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

func (s *InvokerServer) GetCategoryList(ctx context.Context, req *invoker_api.GetCategoryListRequest) (*invoker_api.GetCategoryListResponse, error) {
	startTS := time.Now()
	var resp invoker_api.GetCategoryListResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetCategoryList", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetCategoryList", &resp) }()
	data, appError := s.CategoryService.GetCategoryList(ctx, rpcCtx, req)
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

func (s *InvokerServer) GetCategory(ctx context.Context, req *invoker_api.GetCategoryRequest) (*invoker_api.GetCategoryResponse, error) {
	startTS := time.Now()
	var resp invoker_api.GetCategoryResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetCategory", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetCategory", &resp) }()
	data, appError := s.CategoryService.GetCategory(ctx, rpcCtx, req)
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

func (s *InvokerServer) AddCategory(ctx context.Context, req *invoker_api.AddCategoryRequest) (*invoker_api.AddCategoryResponse, error) {
	startTS := time.Now()
	var resp invoker_api.AddCategoryResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AddCategory", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AddCategory", &resp) }()
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
	appError := s.CategoryService.AddCategory(ctx, rpcCtx, req, loginInfo.Data.UserID)
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

func (s *InvokerServer) EditCategory(ctx context.Context, req *invoker_api.EditCategoryRequest) (*invoker_api.EditCategoryResponse, error) {
	startTS := time.Now()
	var resp invoker_api.EditCategoryResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "EditCategory", req)
	defer func() { s.deferLogResponseData(rpcCtx, "EditCategory", &resp) }()
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
	appError := s.CategoryService.EditCategory(ctx, rpcCtx, req, loginInfo.Data.UserID)
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

func (s *InvokerServer) DeleteCategory(ctx context.Context, req *invoker_api.DeleteCategoryRequest) (*invoker_api.DeleteCategoryResponse, error) {
	startTS := time.Now()
	var resp invoker_api.DeleteCategoryResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeleteCategory", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeleteCategory", &resp) }()
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
	appError := s.CategoryService.DeleteCategory(ctx, rpcCtx, req, loginInfo.Data.UserID)
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
