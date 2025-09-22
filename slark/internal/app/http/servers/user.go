package servers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/nextsurfer/slark/internal/app/http/services"
	"github.com/nextsurfer/slark/internal/pkg/dao"
	"github.com/nextsurfer/slark/internal/pkg/model"
	"github.com/nextsurfer/slark/internal/pkg/redis"
	"github.com/nextsurfer/slark/internal/pkg/util"
)

func DiscourseConnect(w http.ResponseWriter, r *http.Request) {
	discourseConnectSecret := os.Getenv("DISCOURSE_CONNECT_SECRET")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "invalid request method\n")
		return
	}

	queries := r.URL.Query()
	var sso string
	var sig string
	if len(queries["sso"]) > 0 {
		sso = queries["sso"][0]
	}
	if sso == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "sso is an empty string\n")
		return
	}
	if len(queries["sig"]) > 0 {
		sig = queries["sig"][0]
	}
	if sig == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "sig is an empty string\n")
		return
	}

	// 1. validate the signature
	if util.HmacSha256Hex(sso, discourseConnectSecret) != sig {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "signature verificationure\n")
		return
	}

	// 2. redirect to sign-in address
	signInAddr := os.Getenv("DISCOURSE_CONNECT_SIGN_IN_ADDRESS")
	respQueries := url.Values{}
	respQueries.Set("sso", sso)
	respQueries.Set("sig", sig)
	location := signInAddr + "?" + respQueries.Encode()
	w.Header().Add("Location", location)
	w.WriteHeader(http.StatusFound)
	io.WriteString(w, "ok\n")
}

type ValidateSessionIDRequest struct {
	SSO string `json:"sso"`
	Sig string `json:"sig"`
}

func ValidateSessionID(redisOption *redis.Option, daoManager *dao.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       services.Response
			statusCode = http.StatusOK
		)
		defer func() {
			services.DeferWriteResponse(w, statusCode, resp)
		}()
		if !services.MustMethodPost(r, &statusCode, &resp) {
			return
		}
		ctx := r.Context()
		var req ValidateSessionIDRequest
		if err := services.UnmarshalRequestBody(r, &req); err != nil {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = fmt.Sprintf("failed to unmarshal request: %v\n", err)
			return
		}
		// 1. validate the signature
		if err := services.CheckSSO(req.SSO, req.Sig); err != nil {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = err.Error()
			return
		}
		// 2. validate session id
		cookie, session, err := services.CheckCookieAndSessionInfo(ctx, redisOption.Client, r)
		if err != nil {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = err.Error()
			return
		}
		// 3. return user info
		user, err := daoManager.UserDAO.GetFromID(ctx, session.UserID)
		if err != nil {
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.Message = fmt.Sprintf("internal error: %v", err)
			return
		}
		if user == nil {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = fmt.Sprintf("invalid session: %v", err)
			return
		}
		resp.Code = 0
		resp.Message = "ok"
		resp.Data = struct {
			Avatar   string `json:"avatar"`
			Username string `json:"username"`
			Email    string `json:"email"`
		}{
			Username: user.Nickname,
			Email:    user.Email,
		}
		// set cookie back
		http.SetCookie(w, cookie)
	}
}

type SignInRequest struct {
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
	SSO          string `json:"sso"`
	Sig          string `json:"sig"`
}

// two conditions:
// 1. sign in with sso
// 2. sign in without sso
func SignIn(redisOption *redis.Option, daoManager *dao.Manager) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       services.Response
			statusCode = http.StatusOK
		)
		defer func() {
			services.DeferWriteResponse(w, statusCode, resp)
		}()

		if !services.MustMethodPost(r, &statusCode, &resp) {
			return
		}
		ctx := r.Context()
		var req SignInRequest
		if err := services.UnmarshalRequestBody(r, &req); err != nil {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = fmt.Sprintf("failed to unmarshal request: %v\n", err)
			return
		}
		// check sso if it is a request with sso
		if req.SSO != "" || req.Sig != "" {
			if err := services.CheckSSO(req.SSO, req.Sig); err != nil {
				statusCode = http.StatusBadRequest
				resp.Code = 1
				resp.Message = err.Error()
				return
			}
		}
		// check cookie
		var user *model.SlkUser
		var sessionID string
		var cookie *http.Cookie
		cookie, user, sessionID, err := services.CheckCookie(ctx, redisOption.Client, daoManager, req.Email, req.PasswordHash, r)
		if err != nil {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = err.Error()
			return
		}
		// return of sso or general sign in
		statusCode, resp = services.GenerateResponse(ctx, req.SSO, req.Sig, user, sessionID)
		// set cookie back
		if cookie != nil {
			http.SetCookie(w, cookie)
		}
	})
}

func RequestQRLoginToken(redisOption *redis.Option) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       services.Response
			statusCode = http.StatusOK
		)
		defer func() {
			services.DeferWriteResponse(w, statusCode, resp)
		}()

		if !services.MustMethodPost(r, &statusCode, &resp) {
			return
		}

		ctx := r.Context()
		node, err := snowflake.NewNode(1)
		if err != nil {
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.Message = fmt.Sprintf("internal error: %v", err)
			return
		}
		token := node.Generate()

		// cache the token 5 minutes
		if err := redisOption.Client.Set(ctx, token.Base64(), "", time.Minute*5).Err(); err != nil {
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.Message = fmt.Sprintf("internal error: %v", err)
			return
		}

		// return the token
		resp.Code = 0
		resp.Message = "success"
		resp.Data = struct {
			Token string `json:"token"`
		}{
			Token: token.Base64(),
		}
	})
}

type CheckQRLoginRequest struct {
	Token string `json:"token"`
}

func CheckQRLogin(redisOption *redis.Option, daoManager *dao.Manager) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp       services.Response
			statusCode = http.StatusOK
		)
		defer func() {
			services.DeferWriteResponse(w, statusCode, resp)
		}()
		if !services.MustMethodPost(r, &statusCode, &resp) {
			return
		}
		ctx := r.Context()
		var req CheckQRLoginRequest
		if err := services.UnmarshalRequestBody(r, &req); err != nil {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = fmt.Sprintf("failed to unmarshal request: %v\n", err)
			return
		}
		session, err := services.FetchSessionWithToken(ctx, redisOption.Client, req.Token)
		if err != nil {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = fmt.Sprintf("validate session: %v\n", err)
			return
		}
		// get user model
		user, err := daoManager.UserDAO.GetFromID(ctx, session.UserID)
		if err != nil {
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.Message = fmt.Sprintf("internal error: %v", err)
			return
		}
		if user == nil {
			statusCode = http.StatusBadRequest
			resp.Code = 1
			resp.Message = fmt.Sprintf("invalid session: %v", err)
			return
		}
		// delete the token cache after validation
		if err := redisOption.Client.Del(ctx, req.Token).Err(); err != nil {
			statusCode = http.StatusInternalServerError
			resp.Code = 1
			resp.Message = fmt.Sprintf("internal error: %v", err)
			return
		}
		resp.Code = 0
		resp.Message = "success"
		resp.Data = struct {
			Avatar   string `json:"avatar"`
			Username string `json:"username"`
			Email    string `json:"email"`
		}{
			Avatar:   "",
			Username: user.Nickname,
			Email:    user.Email,
		}
	})
}
