package servers

import (
	"context"

	"github.com/go-playground/validator/v10"
	alchemist_api "github.com/nextsurfer/alchemist/api"
	"github.com/nextsurfer/alchemist/api/response"
	"github.com/nextsurfer/alchemist/internal/app/grpc/services"
	"github.com/nextsurfer/alchemist/internal/pkg/dao"
	"github.com/nextsurfer/alchemist/internal/pkg/redis"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type AlchemistConsoleServer struct {
	alchemist_api.UnimplementedAlchemistConsoleServiceServer
	app                     string
	env                     gutil.APPEnvType
	logger                  *zap.Logger
	daoManager              *dao.Manager
	redisOption             *redis.Option
	localizeManager         *localize.Manager
	alchemistConsoleService *services.AlchemistConsoleService
	validator               *validator.Validate
}

func NewAlchemistConsoleServer(app string, env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option, localizeManager *localize.Manager, validator *validator.Validate) (*AlchemistConsoleServer, error) {
	alchemistConsoleService, err := services.NewAlchemistConsoleService(app, env, logger, daoManager, redisOption)
	if err != nil {
		return nil, err
	}
	return &AlchemistConsoleServer{
		app:                     app,
		env:                     env,
		logger:                  logger,
		daoManager:              daoManager,
		redisOption:             redisOption,
		localizeManager:         localizeManager,
		alchemistConsoleService: alchemistConsoleService,
		validator:               validator,
	}, nil
}

func (s *AlchemistConsoleServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *AlchemistConsoleServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := util.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *AlchemistConsoleServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := util.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

// service methods --------------------------------------------------------------------------------------------

func (s *AlchemistConsoleServer) ListConfigs(ctx context.Context, req *alchemist_api.ListConfigsRequest) (*alchemist_api.ListConfigsResponse, error) {
	var resp alchemist_api.ListConfigsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ListConfigs", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ListConfigs", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistConsoleService.ListConfigs(ctx, rpcCtx, req.Password)
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

func (s *AlchemistConsoleServer) CreateConfig(ctx context.Context, req *alchemist_api.CreateConfigRequest) (*alchemist_api.CreateConfigResponse, error) {
	var resp alchemist_api.CreateConfigResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreateConfig", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreateConfig", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.alchemistConsoleService.CreateConfig(ctx, rpcCtx, req.Password, req.Config)
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

func (s *AlchemistConsoleServer) UpdateConfig(ctx context.Context, req *alchemist_api.UpdateConfigRequest) (*alchemist_api.UpdateConfigResponse, error) {
	var resp alchemist_api.UpdateConfigResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UpdateConfig", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UpdateConfig", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.alchemistConsoleService.UpdateConfig(ctx, rpcCtx, req.Password, req.Id, req.Config)
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

func (s *AlchemistConsoleServer) DeleteConfig(ctx context.Context, req *alchemist_api.DeleteConfigRequest) (*alchemist_api.DeleteConfigResponse, error) {
	var resp alchemist_api.DeleteConfigResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeleteConfig", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeleteConfig", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.alchemistConsoleService.DeleteConfig(ctx, rpcCtx, req.Password, req.Id)
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

func (s *AlchemistConsoleServer) GetCurrentSubscriptionCount(ctx context.Context, req *alchemist_api.GetCurrentSubscriptionCountRequest) (*alchemist_api.GetCurrentSubscriptionCountResponse, error) {
	var resp alchemist_api.GetCurrentSubscriptionCountResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetCurrentSubscriptionCount", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetCurrentSubscriptionCount", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistConsoleService.GetCurrentSubscriptionCount(ctx, rpcCtx, req.Password, req.AppID)
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

func (s *AlchemistConsoleServer) ListSubscriptionCounts(ctx context.Context, req *alchemist_api.ListSubscriptionCountsRequest) (*alchemist_api.ListSubscriptionCountsResponse, error) {
	var resp alchemist_api.ListSubscriptionCountsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ListSubscriptionCounts", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ListSubscriptionCounts", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistConsoleService.ListSubscriptionCounts(ctx, rpcCtx, req.Password, req.AppID, req.StartDate, req.EndDate)
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

func (s *AlchemistConsoleServer) GetAllApps(ctx context.Context, req *alchemist_api.GetAllAppsRequest) (*alchemist_api.GetAllAppsResponse, error) {
	var resp alchemist_api.GetAllAppsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetAllApps", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetAllApps", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistConsoleService.GetAllApps(ctx, rpcCtx, req.Password)
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
