package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/slark/internal/pkg/dao"
	"github.com/nextsurfer/slark/internal/pkg/model"
	"github.com/nextsurfer/slark/internal/pkg/util"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func DeferWriteResponse(w http.ResponseWriter, statusCode int, resp Response) {
	respData, err := json.Marshal(resp)
	if err != nil {
		statusCode = http.StatusInternalServerError
		resp.Code = 1
		resp.Message = err.Error()
	}
	w.WriteHeader(statusCode)
	w.Write(respData)
}

func MustMethodPost(r *http.Request, statusCode *int, resp *Response) bool {
	if r.Method != http.MethodPost {
		*statusCode = http.StatusMethodNotAllowed
		resp.Code = 1
		resp.Message = "invalid request method\n"
		return false
	}
	return true
}

func UnmarshalRequestBody(r *http.Request, req interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, req); err != nil {
		return err
	}
	return nil
}

func CheckSSO(sso, sig string) error {
	if sso == "" || sig == "" {
		return errors.New("sso payload and signature are required both")
	}
	if util.HmacSha256Hex(sso, os.Getenv("DISCOURSE_CONNECT_SECRET")) != sig {
		return errors.New("signature verification failure")
	}
	return nil
}

func CheckCookieAndSessionInfo(ctx context.Context, client *redis.Client, r *http.Request) (*http.Cookie, *util.SessionInfo, error) {
	cookie, err := r.Cookie(rpc.DefaultCookieSessionKey)
	if err != nil {
		log.Printf("get cookie [name=%s] failed: %v", rpc.DefaultCookieSessionKey, err)
		return nil, nil, err
	}
	if cookie == nil {
		return nil, nil, errors.New("session cookie is required")
	}
	session, err := util.GetSessionInRedis(ctx, client, cookie.Value)
	if err != nil {
		return nil, nil, err
	}
	return cookie, session, nil
}

func GetRealIP(r *http.Request) string {
	var realIP string
	if realIP = r.Header.Get("x-forwarded-for"); realIP != "" {
		return realIP
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func RefreshSession(ctx context.Context, client *redis.Client, daoManager *dao.Manager, sessionID string, r *http.Request) (*util.SessionInfo, error) {
	// get session info from redis
	session, err := util.GetSessionInRedis(ctx, client, sessionID)
	if err != nil {
		return nil, err
	}
	// update session and cache, because it is a web session
	session.LoginIP = GetRealIP(r)
	session.DeviceID = "NOID"
	if err := daoManager.SessionDAO.UpdateLoginIPInSession(ctx, sessionID, session.LoginIP); err != nil {
		return nil, err
	}
	if err := util.UpdateSessionInRedis(ctx, client, session); err != nil {
		return nil, err
	}
	return session, nil
}

func CheckEmail(ctx context.Context, client *redis.Client, daoManager *dao.Manager, email, passwordHash string, r *http.Request) (*model.SlkUser, error) {
	user, err := daoManager.UserDAO.GetFromEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("unregistered email [%s]", email)
	}
	if user.PasswordHash != passwordHash {
		return nil, errors.New("invalid password")
	}
	return user, nil
}

func NewCookie(ctx context.Context, client *redis.Client, daoManager *dao.Manager, sessionID string, user *model.SlkUser, realIp string) (*http.Cookie, error) {
	sessionObj := &model.SlkSession{
		Application: "slark-http",
		UserID:      user.ID,
		SessionID:   sessionID,
		DeviceID:    "NOID",
		LoginIP:     realIp,
	}
	if err := daoManager.SessionDAO.Create(ctx, sessionObj); err != nil {
		return nil, err
	}
	// cache the session info
	session := util.SessionInfo{
		ExtraInfo: make(map[string]string),
		LoginIP:   realIp,
		UserID:    user.ID,
		DeviceID:  "NOID",
	}
	session.SetSessionID(sessionID)
	if err := util.UpdateSessionInRedis(ctx, client, &session); err != nil {
		return nil, err
	}
	// generate a cookie
	return &http.Cookie{
		HttpOnly: false,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    sessionID,
	}, nil
}

func CheckCookie(ctx context.Context, client *redis.Client, daoManager *dao.Manager, email, passwordHash string, r *http.Request) (*http.Cookie, *model.SlkUser, string, error) {
	// try to get cookie from request
	cookie, err := r.Cookie(rpc.DefaultCookieSessionKey)
	if err != nil {
		log.Printf("get cookie [name=%s] failed\n", rpc.DefaultCookieSessionKey)
	}
	var user *model.SlkUser
	var sessionID string
	// 2-1. validate sessionID if sessionID is not empty
	if cookie != nil {
		sessionID = cookie.Value
		session, err := RefreshSession(ctx, client, daoManager, sessionID, r)
		if err != nil {
			return nil, nil, "", err
		}
		// get user model by user id in session info
		user, err = daoManager.UserDAO.GetFromID(ctx, session.UserID)
		if err != nil {
			return nil, nil, "", err
		}
		if user == nil {
			return nil, nil, "", errors.New("invalid session")
		}
	} else {
		var err error
		// 2-2. validate email and password if sessionID is empty
		// get user model by email
		user, err = CheckEmail(ctx, client, daoManager, email, passwordHash, r)
		if err != nil {
			return nil, nil, "", err
		}
		// add session info
		sessionID = util.NewUUIDHexEncoding()
		cookie, err = NewCookie(ctx, client, daoManager, sessionID, user, GetRealIP(r))
		if err != nil {
			return nil, nil, "", err
		}
	}
	return cookie, user, sessionID, nil
}

func GenerateResponse(ctx context.Context, sso, sig string, user *model.SlkUser, sessionID string) (int, Response) {
	statusCode := http.StatusOK
	var resp Response
	if sso != "" && sig != "" {
		// 3. redirect back
		ssoData, err := base64.StdEncoding.DecodeString(sso)
		if err != nil {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = fmt.Sprintf("failed to decode sso base64 string: %v\n", err)
			return statusCode, resp
		}
		ssoQueries, err := url.ParseQuery(string(ssoData))
		if err != nil {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = fmt.Sprintf("failed to parse sso queries: %v\n", err)
			return statusCode, resp
		}
		var nonce string
		var returnSsoUrl string
		if len(ssoQueries["nonce"]) > 0 {
			nonce = ssoQueries["nonce"][0]
		}
		if nonce == "" {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = "nonce is an empty string\n"
			return statusCode, resp
		}
		if len(ssoQueries["return_sso_url"]) > 0 {
			returnSsoUrl = ssoQueries["return_sso_url"][0]
		}
		if returnSsoUrl == "" {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = "return_sso_url is an empty string\n"
			return statusCode, resp
		}
		// create a new url-encoded payload with at least nonce, email, and external_id.
		payload := url.Values{}
		payload.Set("nonce", nonce)
		payload.Set("email", user.Email)
		payload.Set("external_id", fmt.Sprintf("%d", user.ID))
		payload.Set("username", user.Nickname)
		payloadStr := payload.Encode()
		// encode the url-encoded payload by base 64
		respSso := base64.StdEncoding.EncodeToString([]byte(payloadStr))
		// calculate a HMAC-SHA256 hash of the payload, and base64 encoded payload as text
		respSig := util.HmacSha256Hex(respSso, os.Getenv("DISCOURSE_CONNECT_SECRET"))
		// 6. redirect back
		respQueries := url.Values{}
		respQueries.Set("sso", respSso)
		respQueries.Set("sig", respSig)
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = returnSsoUrl + "?" + respQueries.Encode()
	} else {
		resp.Code = 0
		resp.Message = "success"
		resp.Data = struct {
			Avatar    string `json:"avatar"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			SessionID string `json:"sessionID"`
		}{
			Avatar:    "",
			Username:  user.Nickname,
			Email:     user.Email,
			SessionID: sessionID,
		}
	}
	return statusCode, resp
}

func FetchSessionWithToken(ctx context.Context, client *redis.Client, token string) (*util.SessionInfo, error) {
	get := client.Get(ctx, token)
	if err := get.Err(); err != nil {
		if err == redis.Nil {
			return nil, errors.New("expired token")
		} else {
			return nil, err
		}
	}
	sessionID := get.Val()
	if sessionID == "" {
		return nil, errors.New("have not log in")
	}
	// validate session
	session, err := util.GetSessionInRedis(ctx, client, sessionID)
	if err != nil {
		return nil, err
	}
	return session, nil
}
