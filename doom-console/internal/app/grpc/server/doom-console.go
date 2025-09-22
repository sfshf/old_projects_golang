package server

import (
	"context"

	"github.com/go-playground/validator/v10"
	doom_console_api "github.com/nextsurfer/doom-console/api"
	"github.com/nextsurfer/doom-console/api/response"
	"github.com/nextsurfer/doom-console/internal/app/grpc/service"
	"github.com/nextsurfer/doom-console/internal/pkg/eth"
	"github.com/nextsurfer/doom-console/internal/pkg/riki"
	"github.com/nextsurfer/doom-console/internal/pkg/simplejson"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type DoomConsoleServer struct {
	doom_console_api.UnimplementedDoomConsoleServiceServer
	service         *service.DoomConsoleService
	appID           string
	env             gutil.APPEnvType
	logger          *zap.Logger
	localizeManager *localize.Manager
	validator       *validator.Validate
}

func NewDoomConsoleServer(ctx context.Context, appID string, env gutil.APPEnvType, logger *zap.Logger, localizeManager *localize.Manager, kongAddress string, validator *validator.Validate, mongoDB *mongo.Database) (*DoomConsoleServer, error) {
	s := &DoomConsoleServer{
		appID:           appID,
		env:             env,
		logger:          logger,
		localizeManager: localizeManager,
		validator:       validator,
	}
	// market service
	DoomConsoleService, err := service.NewDoomConsoleService(ctx, logger, mongoDB)
	if err != nil {
		return nil, err
	}
	s.service = DoomConsoleService
	return s, nil
}

func (s *DoomConsoleServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *DoomConsoleServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *DoomConsoleServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

func (s *DoomConsoleServer) ListReputableTokens(ctx context.Context, req *doom_console_api.ListReputableTokensRequest) (*doom_console_api.ListReputableTokensResponse, error) {
	var resp doom_console_api.ListReputableTokensResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ListReputableTokens", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ListReputableTokens", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleRead); err != nil {
		s.logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeUnauthorized
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_ApiKeyIsInvalid")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.service.ListReputableTokens(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *DoomConsoleServer) UniswapInfo(ctx context.Context, req *doom_console_api.UniswapInfoRequest) (*doom_console_api.UniswapInfoResponse, error) {
	var resp doom_console_api.UniswapInfoResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UniswapInfo", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UniswapInfo", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleRead); err != nil {
		s.logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeUnauthorized
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_ApiKeyIsInvalid")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.service.UniswapInfo(ctx, rpcCtx)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *DoomConsoleServer) Erc20TokensInfo(ctx context.Context, req *doom_console_api.Erc20TokensInfoRequest) (*doom_console_api.Erc20TokensInfoResponse, error) {
	var resp doom_console_api.Erc20TokensInfoResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "Erc20TokensInfo", req)
	defer func() { s.deferLogResponseData(rpcCtx, "Erc20TokensInfo", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleRead); err != nil {
		s.logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeUnauthorized
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_ApiKeyIsInvalid")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.service.Erc20TokensInfo(ctx, rpcCtx)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *DoomConsoleServer) Erc20TokensQuery(ctx context.Context, req *doom_console_api.Erc20TokensQueryRequest) (*doom_console_api.Erc20TokensQueryResponse, error) {
	var resp doom_console_api.Erc20TokensQueryResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "Erc20TokensQuery", req)
	defer func() { s.deferLogResponseData(rpcCtx, "Erc20TokensQuery", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleRead); err != nil {
		s.logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeUnauthorized
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_ApiKeyIsInvalid")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	req.ContractAddress = eth.MixedcaseAddress(req.ContractAddress)
	data, appError := s.service.Erc20TokensQuery(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *DoomConsoleServer) ListErrorTokens(ctx context.Context, req *doom_console_api.ListErrorTokensRequest) (*doom_console_api.ListErrorTokensResponse, error) {
	var resp doom_console_api.ListErrorTokensResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ListErrorTokens", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ListErrorTokens", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleRead); err != nil {
		s.logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeUnauthorized
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_ApiKeyIsInvalid")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.service.ListErrorTokens(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *DoomConsoleServer) CheckErrorToken(ctx context.Context, req *doom_console_api.CheckErrorTokenRequest) (*doom_console_api.CheckErrorTokenResponse, error) {
	var resp doom_console_api.CheckErrorTokenResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CheckErrorToken", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CheckErrorToken", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleRead); err != nil {
		s.logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeUnauthorized
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_ApiKeyIsInvalid")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.service.CheckErrorToken(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}

func (s *DoomConsoleServer) DetectErrorToken(ctx context.Context, req *doom_console_api.DetectErrorTokenRequest) (*doom_console_api.DetectErrorTokenResponse, error) {
	var resp doom_console_api.DetectErrorTokenResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DetectErrorToken", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DetectErrorToken", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleRead); err != nil {
		s.logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeUnauthorized
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_ApiKeyIsInvalid")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.service.DetectErrorToken(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *DoomConsoleServer) RpcServerDetection(ctx context.Context, req *doom_console_api.RpcServerDetectionRequest) (*doom_console_api.RpcServerDetectionResponse, error) {
	var resp doom_console_api.RpcServerDetectionResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RpcServerDetection", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RpcServerDetection", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleRead); err != nil {
		s.logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeUnauthorized
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_ApiKeyIsInvalid")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.service.RpcServerDetection(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}
