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

func (s *InvokerServer) GetPostList(ctx context.Context, req *invoker_api.GetPostListRequest) (*invoker_api.GetPostListResponse, error) {
	startTS := time.Now()
	var resp invoker_api.GetPostListResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetPostList", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetPostList", &resp) }()
	data, appError := s.SiteService.GetPostList(ctx, rpcCtx, req)
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

func (s *InvokerServer) GetPostDetail(ctx context.Context, req *invoker_api.GetPostDetailRequest) (*invoker_api.GetPostDetailResponse, error) {
	startTS := time.Now()
	var resp invoker_api.GetPostDetailResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetPostDetail", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetPostDetail", &resp) }()
	data, appError := s.SiteService.GetPostDetail(ctx, rpcCtx, req)
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

func (s *InvokerServer) AddPost(ctx context.Context, req *invoker_api.AddPostRequest) (*invoker_api.AddPostResponse, error) {
	startTS := time.Now()
	var resp invoker_api.AddPostResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AddPost", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AddPost", &resp) }()
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
	appError := s.SiteService.AddPost(ctx, rpcCtx, req, loginInfo.Data.UserID)
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

func (s *InvokerServer) EditPost(ctx context.Context, req *invoker_api.EditPostRequest) (*invoker_api.EditPostResponse, error) {
	startTS := time.Now()
	var resp invoker_api.EditPostResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "EditPost", req)
	defer func() { s.deferLogResponseData(rpcCtx, "EditPost", &resp) }()
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
	appError := s.SiteService.EditPost(ctx, rpcCtx, req, loginInfo.Data.UserID)
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

func (s *InvokerServer) DeletePost(ctx context.Context, req *invoker_api.DeletePostRequest) (*invoker_api.DeletePostResponse, error) {
	startTS := time.Now()
	var resp invoker_api.DeletePostResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeletePost", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeletePost", &resp) }()
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
	appError := s.SiteService.DeletePost(ctx, rpcCtx, req, loginInfo.Data.UserID)
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

func (s *InvokerServer) AggregatedSitePage(ctx context.Context, req *invoker_api.AggregatedSitePageRequest) (*invoker_api.AggregatedSitePageResponse, error) {
	startTS := time.Now()
	var resp invoker_api.AggregatedSitePageResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AggregatedSitePage", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AggregatedSitePage", &resp) }()
	data, appError := s.SiteService.AggregatedSitePage(ctx, rpcCtx, req)
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

func (s *InvokerServer) AggregatedCategoryPage(ctx context.Context, req *invoker_api.AggregatedCategoryPageRequest) (*invoker_api.AggregatedCategoryPageResponse, error) {
	startTS := time.Now()
	var resp invoker_api.AggregatedCategoryPageResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AggregatedCategoryPage", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AggregatedCategoryPage", &resp) }()
	data, appError := s.SiteService.AggregatedCategoryPage(ctx, rpcCtx, req)
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

func (s *InvokerServer) AggregatedPostPage(ctx context.Context, req *invoker_api.AggregatedPostPageRequest) (*invoker_api.AggregatedPostPageResponse, error) {
	startTS := time.Now()
	var resp invoker_api.AggregatedPostPageResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AggregatedPostPage", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AggregatedPostPage", &resp) }()
	data, appError := s.SiteService.AggregatedPostPage(ctx, rpcCtx, req)
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
