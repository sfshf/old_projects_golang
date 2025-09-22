package server

import (
	"context"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
)

func (s *PswdsServer) CreateNonPasswordRecord(ctx context.Context, req *pswds_api.CreateNonPasswordRecordRequest) (*pswds_api.CreateNonPasswordRecordResponse, error) {
	startTS := time.Now()
	var resp pswds_api.CreateNonPasswordRecordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreateNonPasswordRecord", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreateNonPasswordRecord", &resp) }()
	appError := s.NonPasswordService.CreateNonPasswordRecord(ctx, rpcCtx, req)
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

func (s *PswdsServer) UpdateNonPasswordRecord(ctx context.Context, req *pswds_api.UpdateNonPasswordRecordRequest) (*pswds_api.UpdateNonPasswordRecordResponse, error) {
	startTS := time.Now()
	var resp pswds_api.UpdateNonPasswordRecordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UpdateNonPasswordRecord", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UpdateNonPasswordRecord", &resp) }()
	appError := s.NonPasswordService.UpdateNonPasswordRecord(ctx, rpcCtx, req)
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

func (s *PswdsServer) DeleteNonPasswordRecord(ctx context.Context, req *pswds_api.DeleteNonPasswordRecordRequest) (*pswds_api.DeleteNonPasswordRecordResponse, error) {
	startTS := time.Now()
	var resp pswds_api.DeleteNonPasswordRecordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeleteNonPasswordRecord", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeleteNonPasswordRecord", &resp) }()
	appError := s.NonPasswordService.DeleteNonPasswordRecord(ctx, rpcCtx, req)
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
