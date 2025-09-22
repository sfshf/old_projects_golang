package server

import (
	"context"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
)

func (s *PswdsServer) CreatePasswordRecord(ctx context.Context, req *pswds_api.CreatePasswordRecordRequest) (*pswds_api.CreatePasswordRecordResponse, error) {
	startTS := time.Now()
	var resp pswds_api.CreatePasswordRecordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreatePasswordRecord", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreatePasswordRecord", &resp) }()
	appError := s.PasswordService.CreatePasswordRecord(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) UpdatePasswordRecord(ctx context.Context, req *pswds_api.UpdatePasswordRecordRequest) (*pswds_api.UpdatePasswordRecordResponse, error) {
	startTS := time.Now()
	var resp pswds_api.UpdatePasswordRecordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UpdatePasswordRecord", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UpdatePasswordRecord", &resp) }()
	appError := s.PasswordService.UpdatePasswordRecord(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) DeletePasswordRecord(ctx context.Context, req *pswds_api.DeletePasswordRecordRequest) (*pswds_api.DeletePasswordRecordResponse, error) {
	startTS := time.Now()
	var resp pswds_api.DeletePasswordRecordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeletePasswordRecord", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeletePasswordRecord", &resp) }()
	appError := s.PasswordService.DeletePasswordRecord(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}
