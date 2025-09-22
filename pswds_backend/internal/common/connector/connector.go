package connector

import (
	"context"
	"errors"

	connector_grpc "github.com/nextsurfer/connector/pkg/grpc"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/pswds_backend/internal/common/simplecrypto"
)

const (
	RoleWrite = "write"
	RoleRead  = "read"
)

var (
	ErrInvalidApiKey = errors.New("invalid api key")
)

func ValidateApiKey(ctx context.Context, rpcCtx *rpc.Context, app, apikey, role string) error {
	passwordHash, err := simplecrypto.Keccak256Hex([]byte(apikey))
	if err != nil {
		return err
	}
	exist, err := connector_grpc.ValidateApiKey(ctx, rpcCtx, app, string(passwordHash), role)
	if err != nil {
		return err
	}
	if !exist {
		return ErrInvalidApiKey
	}
	return nil
}
