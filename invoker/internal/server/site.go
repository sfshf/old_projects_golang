package server

import (
	"context"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	invoker_api "github.com/nextsurfer/invoker/api"
	"github.com/nextsurfer/invoker/api/response"
)

func (s *InvokerServer) GetSiteList(ctx context.Context, req *invoker_api.GetSiteListRequest) (*invoker_api.GetSiteListResponse, error) {
	startTS := time.Now()
	var resp invoker_api.GetSiteListResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetSiteList", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetSiteList", &resp) }()
	data, appError := s.SiteService.GetSiteList(ctx, rpcCtx, req)
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

func (s *InvokerServer) GetSite(ctx context.Context, req *invoker_api.GetSiteRequest) (*invoker_api.GetSiteResponse, error) {
	startTS := time.Now()
	var resp invoker_api.GetSiteResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetSite", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetSite", &resp) }()
	data, appError := s.SiteService.GetSite(ctx, rpcCtx, req)
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

func (s *InvokerServer) AggregatedSearchPage(ctx context.Context, req *invoker_api.AggregatedSearchPageRequest) (*invoker_api.AggregatedSearchPageResponse, error) {
	startTS := time.Now()
	var resp invoker_api.AggregatedSearchPageResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AggregatedSearchPage", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AggregatedSearchPage", &resp) }()
	data, appError := s.SiteService.AggregatedSearchPage(ctx, rpcCtx, req)
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

func (s *InvokerServer) SearchPostComment(ctx context.Context, req *invoker_api.SearchPostCommentRequest) (*invoker_api.SearchPostCommentResponse, error) {
	startTS := time.Now()
	var resp invoker_api.SearchPostCommentResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "SearchPostComment", req)
	defer func() { s.deferLogResponseData(rpcCtx, "SearchPostComment", &resp) }()
	data, appError := s.SiteService.SearchPostComment(ctx, rpcCtx, req)
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
