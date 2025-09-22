package slark

import (
	"context"
	"errors"
	"net/http"
	"os"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/slark/api"
	"github.com/nextsurfer/slark/api/response"
	"github.com/nextsurfer/slark/internal/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func DialDefault() (*grpc.ClientConn, error) {
	return DialSlarkGrpc(grpc.WithInsecure(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
}

func DialSlarkGrpc(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		err := errors.New("must set env variable for 'CONSUL_HTTP_ADDR'")
		return nil, err
	}
	return grpc.Dial("consul://"+consulAddr+"/slark", opts...)
}

func SessionLoginInfo(ctx context.Context, rpcCtx *rpc.Context) (*api.LoginResponse, error) {
	// check slark login
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	conn, err := DialDefault()
	if err != nil {
		return nil, errors.New("failed to connect to slark server")
	}
	defer conn.Close()
	userServiceClient := api.NewUserServiceClient(conn)
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	return userServiceClient.LoginInfo(ctx, &api.Empty{})
}

func CheckRegistration(ctx context.Context, rpcCtx *rpc.Context, email string) (*api.CheckRegistrationResponse, error) {
	// check slark login
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	conn, err := DialDefault()
	if err != nil {
		return nil, errors.New("failed to connect to slark server")
	}
	defer conn.Close()
	userServiceClient := api.NewUserServiceClient(conn)
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	return userServiceClient.CheckRegistration(ctx, &api.CheckRegistrationRequest{
		Email: email,
	})
}

func ValidateUserIDs(ctx context.Context, rpcCtx *rpc.Context, userIDs []int64) (*api.ValidateUserIDsResponse, error) {
	// check slark login
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	conn, err := DialDefault()
	if err != nil {
		return nil, errors.New("failed to connect to slark server")
	}
	defer conn.Close()
	userServiceClient := api.NewUserServiceClient(conn)
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	return userServiceClient.ValidateUserIDs(ctx, &api.ValidateUserIDsRequest{
		UserIDs: userIDs,
	})
}

func GetUserInfo(ctx context.Context, rpcCtx *rpc.Context, userID int64) (*api.GetUserInfoResponse, error) {
	// check slark login
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	conn, err := DialDefault()
	if err != nil {
		return nil, errors.New("failed to connect to slark server")
	}
	defer conn.Close()
	userServiceClient := api.NewUserServiceClient(conn)
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	return userServiceClient.GetUserInfo(ctx, &api.GetUserInfoRequest{
		Id: userID,
	})
}

func CreateSecondaryPassword(ctx context.Context, rpcCtx *rpc.Context, passwordHash string) (*api.EmptyResponse, error) {
	// check slark login
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	conn, err := DialDefault()
	if err != nil {
		return nil, errors.New("failed to connect to slark server")
	}
	defer conn.Close()
	userServiceClient := api.NewUserServiceClient(conn)
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	return userServiceClient.CreateSecondaryPassword(ctx, &api.CreateSecondaryPasswordRequest{
		PasswordHash: passwordHash,
	})
}

func UpdateSecondaryPassword(ctx context.Context, rpcCtx *rpc.Context, oldPasswordHash, newPasswordHash string) (*api.EmptyResponse, error) {
	// check slark login
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	conn, err := DialDefault()
	if err != nil {
		return nil, errors.New("failed to connect to slark server")
	}
	defer conn.Close()
	userServiceClient := api.NewUserServiceClient(conn)
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	return userServiceClient.UpdateSecondaryPassword(ctx, &api.UpdateSecondaryPasswordRequest{
		OldPasswordHash: oldPasswordHash,
		NewPasswordHash: newPasswordHash,
	})
}

func LoginBySecondaryPassword(ctx context.Context, rpcCtx *rpc.Context, email, passwordHash string) (*api.LoginResponse, error) {
	// check slark login
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	conn, err := DialDefault()
	if err != nil {
		return nil, errors.New("failed to connect to slark server")
	}
	defer conn.Close()
	userServiceClient := api.NewUserServiceClient(conn)
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	var header metadata.MD
	resp, err := userServiceClient.LoginBySecondaryPassword(ctx, &api.LoginBySecondaryPasswordRequest{
		Email:        email,
		PasswordHash: passwordHash,
	}, grpc.Header(&header))
	if err == nil && resp.Code == response.StatusCodeOK {
		var sessionID string
		cookieList := header.Get("Set-Cookie")
		if cookieList == nil {
			cookieList = header.Get("X-SessionID")
		}
		if cookieList != nil {
			if len(cookieList) > 0 {
				rawCookies := cookieList[0]
				header := http.Header{}
				header.Add("Cookie", rawCookies)
				request := http.Request{Header: header}
				sessionCookie, err := request.Cookie(rpc.DefaultCookieSessionKey)
				if err == nil {
					sessionID = sessionCookie.Value
				}
			}
			util.SetSessionInCookie(ctx, sessionID)
		}
	}
	return resp, err
}
