package server

import (
	"context"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	invoker_api "github.com/nextsurfer/invoker/api"
	"github.com/nextsurfer/invoker/api/response"
)

func (s *InvokerServer) GetSites(ctx context.Context, req *invoker_api.GetSitesRequest) (*invoker_api.GetSitesResponse, error) {
	startTS := time.Now()
	var resp invoker_api.GetSitesResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetSites", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetSites", &resp) }()
	data, appError := s.AdminService.GetSites(ctx, rpcCtx, req)
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

func (s *InvokerServer) AddSite(ctx context.Context, req *invoker_api.AddSiteRequest) (*invoker_api.AddSiteResponse, error) {
	startTS := time.Now()
	var resp invoker_api.AddSiteResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AddSite", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AddSite", &resp) }()
	appError := s.AdminService.AddSite(ctx, rpcCtx, req)
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

func (s *InvokerServer) EditSite(ctx context.Context, req *invoker_api.EditSiteRequest) (*invoker_api.EditSiteResponse, error) {
	startTS := time.Now()
	var resp invoker_api.EditSiteResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "EditSite", req)
	defer func() { s.deferLogResponseData(rpcCtx, "EditSite", &resp) }()
	appError := s.AdminService.EditSite(ctx, rpcCtx, req)
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

func (s *InvokerServer) DeleteSite(ctx context.Context, req *invoker_api.DeleteSiteRequest) (*invoker_api.DeleteSiteResponse, error) {
	startTS := time.Now()
	var resp invoker_api.DeleteSiteResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeleteSite", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeleteSite", &resp) }()
	appError := s.AdminService.DeleteSite(ctx, rpcCtx, req)
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

func (s *InvokerServer) GetSiteAdmins(ctx context.Context, req *invoker_api.GetSiteAdminsRequest) (*invoker_api.GetSiteAdminsResponse, error) {
	startTS := time.Now()
	var resp invoker_api.GetSiteAdminsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetSiteAdmins", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetSiteAdmins", &resp) }()
	data, appError := s.AdminService.GetSiteAdmins(ctx, rpcCtx, req)
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

func (s *InvokerServer) AddSiteAdmin(ctx context.Context, req *invoker_api.AddSiteAdminRequest) (*invoker_api.AddSiteAdminResponse, error) {
	startTS := time.Now()
	var resp invoker_api.AddSiteAdminResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AddSiteAdmin", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AddSiteAdmin", &resp) }()
	appError := s.AdminService.AddSiteAdmin(ctx, rpcCtx, req)
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

func (s *InvokerServer) DeleteSiteAdmin(ctx context.Context, req *invoker_api.DeleteSiteAdminRequest) (*invoker_api.DeleteSiteAdminResponse, error) {
	startTS := time.Now()
	var resp invoker_api.DeleteSiteAdminResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeleteSiteAdmin", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeleteSiteAdmin", &resp) }()
	appError := s.AdminService.DeleteSiteAdmin(ctx, rpcCtx, req)
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

func (s *InvokerServer) GetStorageInfo(ctx context.Context, req *invoker_api.GetStorageInfoRequest) (*invoker_api.GetStorageInfoResponse, error) {
	startTS := time.Now()
	var resp invoker_api.GetStorageInfoResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetStorageInfo", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetStorageInfo", &resp) }()
	data, appError := s.AdminService.GetStorageInfo(ctx, rpcCtx, req)
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
