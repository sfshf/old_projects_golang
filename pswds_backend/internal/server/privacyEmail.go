package server

import (
	"context"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
)

func (s *PswdsServer) GetPrivacyEmails(ctx context.Context, req *pswds_api.GetPrivacyEmailsRequest) (*pswds_api.GetPrivacyEmailsResponse, error) {
	startTS := time.Now()
	var resp pswds_api.GetPrivacyEmailsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetPrivacyEmails", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetPrivacyEmails", &resp) }()
	data, appError := s.PasswordService.GetPrivacyEmails(ctx, rpcCtx, req)
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

func (s *PswdsServer) GetPrivacyEmail(ctx context.Context, req *pswds_api.GetPrivacyEmailRequest) (*pswds_api.GetPrivacyEmailResponse, error) {
	startTS := time.Now()
	var resp pswds_api.GetPrivacyEmailResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetPrivacyEmail", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetPrivacyEmail", &resp) }()
	data, appError := s.PasswordService.GetPrivacyEmail(ctx, rpcCtx, req)
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

func (s *PswdsServer) DeletePrivacyEmail(ctx context.Context, req *pswds_api.DeletePrivacyEmailRequest) (*pswds_api.DeletePrivacyEmailResponse, error) {
	startTS := time.Now()
	var resp pswds_api.DeletePrivacyEmailResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeletePrivacyEmail", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeletePrivacyEmail", &resp) }()
	appError := s.PasswordService.DeletePrivacyEmail(ctx, rpcCtx, req)
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

func (s *PswdsServer) AddPrivacyEmailAccount(ctx context.Context, req *pswds_api.AddPrivacyEmailAccountRequest) (*pswds_api.AddPrivacyEmailAccountResponse, error) {
	startTS := time.Now()
	var resp pswds_api.AddPrivacyEmailAccountResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AddPrivacyEmailAccount", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AddPrivacyEmailAccount", &resp) }()
	appError := s.PasswordService.AddPrivacyEmailAccount(ctx, rpcCtx, req)
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
