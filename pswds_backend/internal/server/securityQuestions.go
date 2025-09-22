package server

import (
	"context"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
)

func (s *PswdsServer) UploadSecurityQuestions(ctx context.Context, req *pswds_api.UploadSecurityQuestionsRequest) (*pswds_api.UploadSecurityQuestionsResponse, error) {
	startTS := time.Now()
	var resp pswds_api.UploadSecurityQuestionsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UploadSecurityQuestions", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UploadSecurityQuestions", &resp) }()
	appError := s.PasswordService.UploadSecurityQuestions(ctx, rpcCtx, req)
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
