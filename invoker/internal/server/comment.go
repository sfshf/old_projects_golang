package server

import (
	"context"
	"fmt"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	invoker_api "github.com/nextsurfer/invoker/api"
	"github.com/nextsurfer/invoker/api/response"
	slark_response "github.com/nextsurfer/slark/api/response"
	slark_grpc "github.com/nextsurfer/slark/pkg/grpc"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *InvokerServer) GetComments(ctx context.Context, req *invoker_api.GetCommentsRequest) (*invoker_api.GetCommentsResponse, error) {
	startTS := time.Now()
	var resp invoker_api.GetCommentsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetComments", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetComments", &resp) }()
	data, appError := s.CommentService.GetComments(ctx, rpcCtx, req)
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

func (s *InvokerServer) AddComment(ctx context.Context, req *invoker_api.AddCommentRequest) (*invoker_api.AddCommentResponse, error) {
	startTS := time.Now()
	var resp invoker_api.AddCommentResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AddComment", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AddComment", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			// oracle field
			resp.Oracle = s.oracleField(startTS)
			return &resp, nil
		}
	}
	appError := s.CommentService.AddComment(ctx, rpcCtx, req, loginInfo.Data.UserID)
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

func (s *InvokerServer) EditComment(ctx context.Context, req *invoker_api.EditCommentRequest) (*invoker_api.EditCommentResponse, error) {
	startTS := time.Now()
	var resp invoker_api.EditCommentResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "EditComment", req)
	defer func() { s.deferLogResponseData(rpcCtx, "EditComment", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			// oracle field
			resp.Oracle = s.oracleField(startTS)
			return &resp, nil
		}
	}
	appError := s.CommentService.EditComment(ctx, rpcCtx, req, loginInfo.Data.UserID)
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

func (s *InvokerServer) DeleteComment(ctx context.Context, req *invoker_api.DeleteCommentRequest) (*invoker_api.DeleteCommentResponse, error) {
	startTS := time.Now()
	var resp invoker_api.DeleteCommentResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeleteComment", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeleteComment", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			// oracle field
			resp.Oracle = s.oracleField(startTS)
			return &resp, nil
		}
	}
	appError := s.CommentService.DeleteComment(ctx, rpcCtx, req, loginInfo.Data.UserID)
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
