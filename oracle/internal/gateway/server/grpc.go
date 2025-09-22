package server

import (
	"context"
	"fmt"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	gateway_api "github.com/nextsurfer/oracle/api/gateway"
	"github.com/nextsurfer/oracle/api/response"
	"github.com/nextsurfer/oracle/internal/common/simplejson"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *GatewayServer) isEnvTest() bool {
	return s.Env == gutil.AppEnvDEV || s.Env == gutil.AppEnvPPE
}

func (s *GatewayServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *GatewayServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

func (s *GatewayServer) RefreshService(ctx context.Context, req *gateway_api.RefreshServiceRequest) (*gateway_api.RefreshServiceResponse, error) {
	var resp gateway_api.RefreshServiceResponse
	rpcCtx := rpc.NewContext(ctx, s.LocalizeManager)
	s.logRequestData(rpcCtx, "RefreshService", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RefreshService", &resp) }()
	// validate request basically
	if err := s.Validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.GatewayService.RefreshService(ctx, rpcCtx, req.Name)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	// refresh rate limit rules
	s.refreshRateLimitRules(ctx, req.Name)
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}

func (s *GatewayServer) RefreshCertificate(ctx context.Context, req *gateway_api.RefreshCertificateRequest) (*gateway_api.RefreshCertificateResponse, error) {
	var resp gateway_api.RefreshCertificateResponse
	rpcCtx := rpc.NewContext(ctx, s.LocalizeManager)
	s.logRequestData(rpcCtx, "RefreshCertificate", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RefreshCertificate", &resp) }()
	// validate request basically
	if err := s.Validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// just need to delete local cache because of using of tls.Config.GetCertificate
	delete(s.TlsCerts, req.Domain)
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}

func (s *GatewayServer) RefreshProxy(ctx context.Context, req *gateway_api.RefreshProxyRequest) (*gateway_api.RefreshProxyResponse, error) {
	var resp gateway_api.RefreshProxyResponse
	rpcCtx := rpc.NewContext(ctx, s.LocalizeManager)
	s.logRequestData(rpcCtx, "RefreshProxy", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RefreshProxy", &resp) }()
	// validate request basically
	if err := s.Validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	hostname, err := s.DaoManager.HostManageDAO.GetByDomain(ctx, req.Domain)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	if hostname == nil {
		err = fmt.Errorf("domain %s not found", req.Domain)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_BadRequest")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	delete(s.HostManage, hostname.Domain)
	s.HostManage[hostname.Domain] = hostname.RawURL
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}
