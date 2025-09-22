package server

import (
	"context"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
)

func (s *PswdsServer) CheckUpdates(ctx context.Context, req *pswds_api.CheckUpdatesRequest) (*pswds_api.CheckUpdatesResponse, error) {
	startTS := time.Now()
	var resp pswds_api.CheckUpdatesResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CheckUpdates", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CheckUpdates", &resp) }()
	data, appError := s.PasswordService.CheckUpdates(ctx, rpcCtx, req)
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

func (s *PswdsServer) UploadData(ctx context.Context, req *pswds_api.UploadDataRequest) (*pswds_api.UploadDataResponse, error) {
	startTS := time.Now()
	var resp pswds_api.UploadDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UploadData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UploadData", &resp) }()
	appError := s.PasswordService.UploadData(ctx, rpcCtx, req)
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

func (s *PswdsServer) DownloadData(ctx context.Context, req *pswds_api.DownloadDataRequest) (*pswds_api.DownloadDataResponse, error) {
	startTS := time.Now()
	var resp pswds_api.DownloadDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DownloadData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DownloadData", &resp) }()
	data, appError := s.PasswordService.DownloadData(ctx, rpcCtx, req)
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

func (s *PswdsServer) RecoverUnlockPassword(ctx context.Context, req *pswds_api.RecoverUnlockPasswordRequest) (*pswds_api.RecoverUnlockPasswordResponse, error) {
	startTS := time.Now()
	var resp pswds_api.RecoverUnlockPasswordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RecoverUnlockPassword", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RecoverUnlockPassword", &resp) }()
	appError := s.PasswordService.RecoverUnlockPassword(ctx, rpcCtx, req)
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

func (s *PswdsServer) GetBackupCiphertext(ctx context.Context, req *pswds_api.GetBackupCiphertextRequest) (*pswds_api.GetBackupCiphertextResponse, error) {
	startTS := time.Now()
	var resp pswds_api.GetBackupCiphertextResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetBackupCiphertext", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetBackupCiphertext", &resp) }()
	appError := s.PasswordService.GetBackupCiphertext(ctx, rpcCtx, req)
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

func (s *PswdsServer) CheckUnlockPasswordBackups(ctx context.Context, req *pswds_api.CheckUnlockPasswordBackupsRequest) (*pswds_api.CheckUnlockPasswordBackupsResponse, error) {
	startTS := time.Now()
	var resp pswds_api.CheckUnlockPasswordBackupsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CheckUnlockPasswordBackups", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CheckUnlockPasswordBackups", &resp) }()
	data, appError := s.PasswordService.CheckUnlockPasswordBackups(ctx, rpcCtx, req)
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
