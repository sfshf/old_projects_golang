package util

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nextsurfer/ground/pkg/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	PrefixSession = "Session:"
)

// SessionInfo save session info in redis
type SessionInfo struct {
	LoginIP   string
	UserID    int64
	DeviceID  string
	sessionID string
	ExtraInfo map[string]string
}

// SetSessionID set session id
func (s *SessionInfo) SetSessionID(sessionID string) {
	s.sessionID = sessionID
}

// UpdateSessionInRedis update redis
func UpdateSessionInRedis(ctx context.Context, rdsClient *redis.Client, session *SessionInfo) error {
	str, err := json.Marshal(session)
	if err != nil {
		return err
	}
	if err := rdsClient.Set(ctx, PrefixSession+session.sessionID, str, time.Hour*24*365).Err(); err != nil {
		return err
	}
	return nil
}

func GetSessionInRedis(ctx context.Context, rdsClient *redis.Client, sessionID string) (*SessionInfo, error) {
	strCmd := rdsClient.Get(ctx, PrefixSession+sessionID)
	if err := strCmd.Err(); err != nil {
		return nil, err
	}
	data, err := strCmd.Bytes()
	if err != nil {
		return nil, err
	}
	var session SessionInfo
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteSessionInRedis delete session in redis
func DeleteSessionInRedis(ctx context.Context, rpcCtx *rpc.Context, rdsClient *redis.Client) error {
	return rdsClient.Del(ctx, PrefixSession+rpcCtx.SessionID).Err()
}

// SetSessionInCookie pass set-cookie in header
func SetSessionInCookie(ctx context.Context, sessionID string) {
	cookie := &http.Cookie{
		HttpOnly: false,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    sessionID,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	grpc.SendHeader(ctx, metadata.New(map[string]string{
		"Set-Cookie": cookie.String(),
	}))
}

// RemoveSessionInCookie remove cookie
func RemoveSessionInCookie(ctx context.Context) {
	cookie := &http.Cookie{
		HttpOnly: false,
		Path:     "/",
		Expires:  time.Unix(0, 0),
		Name:     rpc.DefaultCookieSessionKey,
		Value:    "",
	}
	grpc.SendHeader(ctx, metadata.New(map[string]string{
		"Set-Cookie": cookie.String(),
	}))
}

const (
	PrefixLoginInfo = "LoginInfo:"
)

type LoginInfo struct {
	SessionID string
	UserID    int64
	Nickname  string
	Email     string
	Phone     string
}

func UpdateLoginInfoInRedis(ctx context.Context, rdsClient *redis.Client, loginInfo *LoginInfo) error {
	str, err := json.Marshal(loginInfo)
	if err != nil {
		return err
	}
	if err := rdsClient.Set(ctx, PrefixLoginInfo+loginInfo.SessionID, str, time.Hour*24).Err(); err != nil {
		return err
	}
	return nil
}

func GetLoginInfoInRedis(ctx context.Context, rdsClient *redis.Client, sessionID string) (*LoginInfo, error) {
	strCmd := rdsClient.Get(ctx, PrefixLoginInfo+sessionID)
	if err := strCmd.Err(); err != nil {
		return nil, err
	}
	data, err := strCmd.Bytes()
	if err != nil {
		return nil, err
	}
	var loginInfo LoginInfo
	if err := json.Unmarshal(data, &loginInfo); err != nil {
		return nil, err
	}
	return &loginInfo, nil
}
