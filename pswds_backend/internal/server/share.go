package server

import (
	"context"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
)

func (s *PswdsServer) GetAirdropID(ctx context.Context, req *pswds_api.GetAirdropIDRequest) (*pswds_api.GetAirdropIDResponse, error) {
	startTS := time.Now()
	var resp pswds_api.GetAirdropIDResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetAirdropID", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetAirdropID", &resp) }()
	data, appError := s.ShareService.GetAirdropID(ctx, rpcCtx, req)
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
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) RequestAirdropData(ctx context.Context, req *pswds_api.RequestAirdropDataRequest) (*pswds_api.RequestAirdropDataResponse, error) {
	startTS := time.Now()
	var resp pswds_api.RequestAirdropDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RequestAirdropData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RequestAirdropData", &resp) }()
	data, appError := s.ShareService.RequestAirdropData(ctx, rpcCtx, req)
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
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) UploadAirdropData(ctx context.Context, req *pswds_api.UploadAirdropDataRequest) (*pswds_api.UploadAirdropDataResponse, error) {
	startTS := time.Now()
	var resp pswds_api.UploadAirdropDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UploadAirdropData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UploadAirdropData", &resp) }()
	appError := s.ShareService.UploadAirdropData(ctx, rpcCtx, req)
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
