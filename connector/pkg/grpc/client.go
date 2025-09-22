package connector

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/klaytn/klaytn/crypto/sha3"
	_ "github.com/mbobakov/grpc-consul-resolver"
	connector_api "github.com/nextsurfer/connector/api"
	"github.com/nextsurfer/ground/pkg/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func DialDefault() (*grpc.ClientConn, error) {
	return DialConnectorGrpc(grpc.WithInsecure(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
}

func DialConnectorGrpc(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		err := errors.New("must set env variable for 'CONSUL_HTTP_ADDR'")
		log.Println(err)
		return nil, err
	}
	return grpc.Dial("consul://"+consulAddr+"/connector", opts...)
}

func hashText(src []byte) ([]byte, error) {
	h := sha3.NewKeccak256()
	if _, err := h.Write(src); err != nil {
		return nil, err
	}
	sum := h.Sum(nil)
	dst := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(dst, sum)
	return dst, nil
}

func CreatePassword(ctx context.Context, rpcCtx *rpc.Context, apiKey, keyID string) (string, error) {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return "", errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorServiceClient(conn)
	respData, err := connectorServiceClient.CreatePassword(ctx, &connector_api.CreatePasswordRequest{
		ApiKey: apiKey,
		KeyID:  keyID,
	})
	if err != nil {
		return "", err
	}
	if respData.Code != 0 {
		return "", fmt.Errorf("save data: %v", respData)
	}
	return respData.Data.Password, nil
}

func CheckPassword(ctx context.Context, rpcCtx *rpc.Context, apiKey, keyID, passwordHash string) (bool, error) {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return false, errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorServiceClient(conn)
	respData, err := connectorServiceClient.CheckPassword(ctx, &connector_api.CheckPasswordRequest{
		ApiKey:       apiKey,
		KeyID:        keyID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return false, err
	}
	if respData.Code != 0 {
		return false, fmt.Errorf("save data: %v", respData)
	}
	return respData.Data.Valid, nil
}

func DeletePassword(ctx context.Context, rpcCtx *rpc.Context, apiKey, keyID string) error {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorServiceClient(conn)
	respData, err := connectorServiceClient.DeletePassword(ctx, &connector_api.DeletePasswordRequest{
		ApiKey: apiKey,
		KeyID:  keyID,
	})
	if err != nil {
		return err
	}
	if respData.Code != 0 {
		return fmt.Errorf("save data: %v", respData)
	}
	return nil
}

func CreatePrivateKey(ctx context.Context, rpcCtx *rpc.Context, apiKey, keyID string) (string, error) {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return "", errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorServiceClient(conn)
	respData, err := connectorServiceClient.CreatePrivateKey(ctx, &connector_api.CreatePrivateKeyRequest{
		ApiKey: apiKey,
		KeyID:  keyID,
	})
	if err != nil {
		return "", err
	}
	if respData.Code != 0 {
		return "", fmt.Errorf("save data: %v", respData)
	}
	return respData.Data.PublicKey, nil
}

func GetPublicKey(ctx context.Context, rpcCtx *rpc.Context, apiKey, keyID string) (string, error) {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return "", errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorServiceClient(conn)
	respData, err := connectorServiceClient.GetPublicKey(ctx, &connector_api.GetPublicKeyRequest{
		ApiKey: apiKey,
		KeyID:  keyID,
	})
	if err != nil {
		return "", err
	}
	if respData.Code != 0 {
		return "", fmt.Errorf("save data: %v", respData)
	}
	return respData.Data.PublicKey, nil
}

func CheckKeyExisting(ctx context.Context, rpcCtx *rpc.Context, apiKey, keyID string) (bool, error) {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return false, errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorServiceClient(conn)
	respData, err := connectorServiceClient.CheckKeyExisting(ctx, &connector_api.CheckKeyExistingRequest{
		ApiKey: apiKey,
		KeyID:  keyID,
	})
	if err != nil {
		return false, err
	}
	if respData.Code != 0 {
		return false, fmt.Errorf("save data: %v", respData)
	}
	return respData.Data.Existing, nil
}

func DeletePrivateKey(ctx context.Context, rpcCtx *rpc.Context, apiKey, keyID string) error {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorServiceClient(conn)
	respData, err := connectorServiceClient.DeletePrivateKey(ctx, &connector_api.DeletePrivateKeyRequest{
		ApiKey: apiKey,
		KeyID:  keyID,
	})
	if err != nil {
		return err
	}
	if respData.Code != 0 {
		return fmt.Errorf("save data: %v", respData)
	}
	return nil
}

func SaveData(ctx context.Context, rpcCtx *rpc.Context, apiKey, keyID, dataID, plaintext, ciphertext string, replaceCurrentItem bool) error {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorServiceClient(conn)
	plaintextHash, err := hashText([]byte(plaintext))
	if err != nil {
		return fmt.Errorf("plaintext hash: %v", err)
	}
	respData, err := connectorServiceClient.SaveData(ctx, &connector_api.SaveDataRequest{
		ApiKey:             apiKey,
		KeyID:              keyID,
		DataID:             dataID,
		ReplaceCurrentItem: replaceCurrentItem,
		Data:               ciphertext,
		PlaintextHash:      string(plaintextHash),
	})
	if err != nil {
		return err
	}
	if respData.Code != 0 {
		return fmt.Errorf("save data: %v", respData)
	}
	return nil
}

func GetData(ctx context.Context, rpcCtx *rpc.Context, apiKey, keyID, dataID string) (string, error) {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return "", errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorServiceClient(conn)
	respData, err := connectorServiceClient.GetData(ctx, &connector_api.GetDataRequest{
		ApiKey: apiKey,
		KeyID:  keyID,
		DataID: dataID,
	})
	if err != nil {
		return "", err
	}
	if respData.Code != 0 {
		return "", fmt.Errorf("save data: %v", respData)
	}
	plaintext, err := base64.StdEncoding.DecodeString(respData.Data.Data)
	if err != nil {
		return "", fmt.Errorf("base64 decode data: %v", err)
	}
	return string(plaintext), nil
}

func DeleteData(ctx context.Context, rpcCtx *rpc.Context, apiKey, keyID, dataID string) error {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorServiceClient(conn)
	respData, err := connectorServiceClient.DeleteData(ctx, &connector_api.DeleteDataRequest{
		ApiKey: apiKey,
		KeyID:  keyID,
		DataID: dataID,
	})
	if err != nil {
		return err
	}
	if respData.Code != 0 {
		return fmt.Errorf("delete data: %v", respData)
	}
	return nil
}

func DecryptData(ctx context.Context, rpcCtx *rpc.Context, apiKey, keyID, data string) (string, error) {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return "", errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorServiceClient(conn)
	respData, err := connectorServiceClient.DecryptData(ctx, &connector_api.DecryptDataRequest{
		ApiKey: apiKey,
		KeyID:  keyID,
		Data:   data,
	})
	if err != nil {
		return "", err
	}
	if respData.Code != 0 {
		return "", fmt.Errorf("save data: %v", respData)
	}
	return respData.Data.DecrypedData, nil
}

func ValidateApiKey(ctx context.Context, rpcCtx *rpc.Context, app, apiKey, role string) (bool, error) {
	cookie := &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    rpcCtx.SessionID,
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"user-agent":      []string{rpcCtx.UserAgent},
		"accept-language": []string{rpcCtx.Localizer.Language()},
		"request-id":      []string{rpcCtx.RequestID},
		"device-id":       []string{rpcCtx.DeviceID},
		"cookie":          []string{cookie.String()},
		"x-real-ip":       []string{rpcCtx.IP},
	})
	conn, err := DialDefault()
	if err != nil {
		return false, errors.New("failed to connect to connector server")
	}
	defer conn.Close()
	connectorServiceClient := connector_api.NewConnectorConsoleServiceClient(conn)
	respData, err := connectorServiceClient.ValidateApiKey(ctx, &connector_api.ValidateApiKeyRequest{
		App:    app,
		ApiKey: apiKey,
		Role:   role,
	})
	if err != nil {
		return false, err
	}
	if respData.Code != 0 {
		return false, fmt.Errorf("save data: %v", respData)
	}
	return respData.Data.Valid, nil
}
