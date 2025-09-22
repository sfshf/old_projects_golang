package slark

import (
	"context"

	"github.com/nextsurfer/ground/pkg/rpc"
	slark_api "github.com/nextsurfer/slark/api"
	slark_grpc "github.com/nextsurfer/slark/pkg/grpc"
)

func CheckRegistration(ctx context.Context, rpcCtx *rpc.Context, email string) (*slark_api.CheckRegistrationResponse_Data, error) {
	resp, err := slark_grpc.CheckRegistration(ctx, rpcCtx, email)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func ValidateUserIDs(ctx context.Context, rpcCtx *rpc.Context, userIDs []int64) ([]bool, error) {
	resp, err := slark_grpc.ValidateUserIDs(ctx, rpcCtx, userIDs)
	if err != nil {
		return nil, err
	}
	return resp.Data.List, nil
}

func GetUserInfo(ctx context.Context, rpcCtx *rpc.Context, userID int64) (*slark_api.GetUserInfoResponse_Data, error) {
	resp, err := slark_grpc.GetUserInfo(ctx, rpcCtx, userID)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func LoginInfo(ctx context.Context, rpcCtx *rpc.Context) (*slark_api.LoginResponse_Data, error) {
	resp, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
