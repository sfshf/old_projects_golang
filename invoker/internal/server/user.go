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

func (s *InvokerServer) ThumbupPost(ctx context.Context, req *invoker_api.ThumbupPostRequest) (*invoker_api.ThumbupPostResponse, error) {
	startTS := time.Now()
	var resp invoker_api.ThumbupPostResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ThumbupPost", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ThumbupPost", &resp) }()
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
	appError := s.UserService.ThumbupPost(ctx, rpcCtx, req, loginInfo.Data.UserID)
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

func (s *InvokerServer) ThumbupComment(ctx context.Context, req *invoker_api.ThumbupCommentRequest) (*invoker_api.ThumbupCommentResponse, error) {
	startTS := time.Now()
	var resp invoker_api.ThumbupCommentResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ThumbupComment", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ThumbupComment", &resp) }()
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
	appError := s.UserService.ThumbupComment(ctx, rpcCtx, req, loginInfo.Data.UserID)
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

func (s *InvokerServer) PostHistory(ctx context.Context, req *invoker_api.PostHistoryRequest) (*invoker_api.PostHistoryResponse, error) {
	startTS := time.Now()
	var resp invoker_api.PostHistoryResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "PostHistory", req)
	defer func() { s.deferLogResponseData(rpcCtx, "PostHistory", &resp) }()
	data, appError := s.UserService.PostHistory(ctx, rpcCtx, req)
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

func (s *InvokerServer) CommentHistory(ctx context.Context, req *invoker_api.CommentHistoryRequest) (*invoker_api.CommentHistoryResponse, error) {
	startTS := time.Now()
	var resp invoker_api.CommentHistoryResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CommentHistory", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CommentHistory", &resp) }()
	data, appError := s.UserService.CommentHistory(ctx, rpcCtx, req)
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

func (s *InvokerServer) ThumbupHistory(ctx context.Context, req *invoker_api.ThumbupHistoryRequest) (*invoker_api.ThumbupHistoryResponse, error) {
	startTS := time.Now()
	var resp invoker_api.ThumbupHistoryResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ThumbupHistory", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ThumbupHistory", &resp) }()
	data, appError := s.UserService.ThumbupHistory(ctx, rpcCtx, req)
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
