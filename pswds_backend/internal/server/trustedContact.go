package server

import (
	"context"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
)

func (s *PswdsServer) CreateTrustedContact(ctx context.Context, req *pswds_api.CreateTrustedContactRequest) (*pswds_api.CreateTrustedContactResponse, error) {
	startTS := time.Now()
	var resp pswds_api.CreateTrustedContactResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreateTrustedContact", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreateTrustedContact", &resp) }()
	appError := s.PasswordService.CreateTrustedContact(ctx, rpcCtx, req)
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

func (s *PswdsServer) GetTrustedContacts(ctx context.Context, req *pswds_api.GetTrustedContactsRequest) (*pswds_api.GetTrustedContactsResponse, error) {
	startTS := time.Now()
	var resp pswds_api.GetTrustedContactsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetTrustedContacts", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetTrustedContacts", &resp) }()
	data, appError := s.PasswordService.GetTrustedContacts(ctx, rpcCtx, req)
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
	resp.Data = data
	return &resp, nil
}

func (s *PswdsServer) DeleteTrustedContact(ctx context.Context, req *pswds_api.DeleteTrustedContactRequest) (*pswds_api.DeleteTrustedContactResponse, error) {
	startTS := time.Now()
	var resp pswds_api.DeleteTrustedContactResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeleteTrustedContact", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeleteTrustedContact", &resp) }()
	appError := s.PasswordService.DeleteTrustedContact(ctx, rpcCtx, req)
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
