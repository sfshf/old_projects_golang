package grpcservers

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	connector_api "github.com/nextsurfer/connector/api"
	"github.com/nextsurfer/connector/api/response"
	"github.com/nextsurfer/connector/internal/app/grpc/services"
	"github.com/nextsurfer/connector/internal/pkg/dao"
	"github.com/nextsurfer/connector/internal/pkg/redis"
	"github.com/nextsurfer/connector/internal/pkg/simplejson"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"
)

type ConnectorConsoleServer struct {
	connector_api.UnimplementedConnectorConsoleServiceServer
	env             gutil.APPEnvType
	logger          *zap.Logger
	daoManager      *dao.Manager
	redisOption     *redis.Option
	localizeManager *localize.Manager
	consoleService  *services.ConnectorConsoleService
	validator       *validator.Validate
}

func NewConnectorConsoleServer(env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option, localizeManager *localize.Manager, validator *validator.Validate) *ConnectorConsoleServer {
	consoleService := services.NewConnectorConsoleService(env, logger, daoManager, redisOption)
	return &ConnectorConsoleServer{
		env:             env,
		logger:          logger,
		daoManager:      daoManager,
		redisOption:     redisOption,
		localizeManager: localizeManager,
		consoleService:  consoleService,
		validator:       validator,
	}
}

func (s *ConnectorConsoleServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *ConnectorConsoleServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *ConnectorConsoleServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

func (s *ConnectorConsoleServer) GetLogs(ctx context.Context, req *connector_api.GetLogsRequest) (*connector_api.GetLogsResponse, error) {
	startTS := time.Now()
	var resp connector_api.GetLogsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetLogs", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetLogs", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "GetLogs", time.Since(startTS).Milliseconds(), err, req.ApiKey, "", "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.GetLogs(ctx, rpcCtx, startTS, req.ApiKey, req.PageNumber, req.PageSize)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) GetMonitorInfos(ctx context.Context, req *connector_api.GetMonitorInfosRequest) (*connector_api.GetMonitorInfosResponse, error) {
	startTS := time.Now()
	var resp connector_api.GetMonitorInfosResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetMonitorInfos", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetMonitorInfos", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "GetMonitorInfos", time.Since(startTS).Milliseconds(), err, req.ApiKey, "", "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.GetMonitorInfos(ctx, rpcCtx, startTS, req.ApiKey)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) ListConfigs(ctx context.Context, req *connector_api.ListConfigsRequest) (*connector_api.ListConfigsResponse, error) {
	startTS := time.Now()
	var resp connector_api.ListConfigsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ListConfigs", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ListConfigs", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "ListConfigs", time.Since(startTS).Milliseconds(), err, req.ApiKey, "", "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.ListConfigs(ctx, rpcCtx, startTS, req.ApiKey)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) CreateConfig(ctx context.Context, req *connector_api.CreateConfigRequest) (*connector_api.CreateConfigResponse, error) {
	startTS := time.Now()
	var resp connector_api.CreateConfigResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreateConfig", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreateConfig", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "CreateConfig", time.Since(startTS).Milliseconds(), err, req.ApiKey, req.App, "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	err := s.consoleService.CreateConfig(ctx, rpcCtx, startTS, req.ApiKey, req.App, req.Config)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}

func (s *ConnectorConsoleServer) UpdateConfig(ctx context.Context, req *connector_api.UpdateConfigRequest) (*connector_api.UpdateConfigResponse, error) {
	startTS := time.Now()
	var resp connector_api.UpdateConfigResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UpdateConfig", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UpdateConfig", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "UpdateConfig", time.Since(startTS).Milliseconds(), err, req.ApiKey, req.App, "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	err := s.consoleService.UpdateConfig(ctx, rpcCtx, startTS, req.ApiKey, req.App, req.Config)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}

func (s *ConnectorConsoleServer) DeleteConfig(ctx context.Context, req *connector_api.DeleteConfigRequest) (*connector_api.DeleteConfigResponse, error) {
	startTS := time.Now()
	var resp connector_api.DeleteConfigResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeleteConfig", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeleteConfig", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "DeleteConfig", time.Since(startTS).Milliseconds(), err, req.ApiKey, req.App, "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	err := s.consoleService.DeleteConfig(ctx, rpcCtx, startTS, req.ApiKey, req.App)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}

func (s *ConnectorConsoleServer) ListPassword(ctx context.Context, req *connector_api.ListPasswordRequest) (*connector_api.ListPasswordResponse, error) {
	startTS := time.Now()
	var resp connector_api.ListPasswordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ListPassword", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ListPassword", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "ListPassword", time.Since(startTS).Milliseconds(), err, req.ApiKey, req.App, "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.ListPassword(ctx, rpcCtx, startTS, req.ApiKey, req.App)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) AddPassword(ctx context.Context, req *connector_api.AddPasswordRequest) (*connector_api.AddPasswordResponse, error) {
	startTS := time.Now()
	var resp connector_api.AddPasswordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AddPassword", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AddPassword", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "AddPassword", time.Since(startTS).Milliseconds(), err, req.ApiKey, req.App, req.KeyID)
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.AddPassword(ctx, rpcCtx, startTS, req.ApiKey, req.App, req.KeyID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) RemovePassword(ctx context.Context, req *connector_api.RemovePasswordRequest) (*connector_api.RemovePasswordResponse, error) {
	startTS := time.Now()
	var resp connector_api.RemovePasswordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RemovePassword", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RemovePassword", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "RemovePassword", time.Since(startTS).Milliseconds(), err, req.ApiKey, req.App, req.KeyID)
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	err := s.consoleService.RemovePassword(ctx, rpcCtx, startTS, req.ApiKey, req.App, req.KeyID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}

func (s *ConnectorConsoleServer) FetchPassword(ctx context.Context, req *connector_api.FetchPasswordRequest) (*connector_api.FetchPasswordResponse, error) {
	startTS := time.Now()
	var resp connector_api.FetchPasswordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "FetchPassword", req)
	defer func() { s.deferLogResponseData(rpcCtx, "FetchPassword", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "FetchPassword", time.Since(startTS).Milliseconds(), err, req.ApiKey, req.App, req.KeyID)
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.FetchPassword(ctx, rpcCtx, startTS, req.ApiKey, req.App, req.KeyID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) ListPrivateKey(ctx context.Context, req *connector_api.ListPrivateKeyRequest) (*connector_api.ListPrivateKeyResponse, error) {
	startTS := time.Now()
	var resp connector_api.ListPrivateKeyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ListPrivateKey", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ListPrivateKey", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "ListPrivateKey", time.Since(startTS).Milliseconds(), err, req.ApiKey, req.App, "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.ListPrivateKey(ctx, rpcCtx, startTS, req.ApiKey, req.App)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) AddPrivateKey(ctx context.Context, req *connector_api.AddPrivateKeyRequest) (*connector_api.AddPrivateKeyResponse, error) {
	startTS := time.Now()
	var resp connector_api.AddPrivateKeyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AddPrivateKey", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AddPrivateKey", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "AddPrivateKey", time.Since(startTS).Milliseconds(), err, req.ApiKey, req.App, req.KeyID)
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.AddPrivateKey(ctx, rpcCtx, startTS, req.ApiKey, req.App, req.KeyID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) RemovePrivateKey(ctx context.Context, req *connector_api.RemovePrivateKeyRequest) (*connector_api.RemovePrivateKeyResponse, error) {
	startTS := time.Now()
	var resp connector_api.RemovePrivateKeyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RemovePrivateKey", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RemovePrivateKey", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "RemovePrivateKey", time.Since(startTS).Milliseconds(), err, req.ApiKey, req.App, req.KeyID)
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	err := s.consoleService.RemovePrivateKey(ctx, rpcCtx, startTS, req.ApiKey, req.App, req.KeyID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}

func (s *ConnectorConsoleServer) GetAllApps(ctx context.Context, req *connector_api.GetAllAppsRequest) (*connector_api.GetAllAppsResponse, error) {
	startTS := time.Now()
	var resp connector_api.GetAllAppsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetAllApps", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetAllApps", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "GetAllApps", time.Since(startTS).Milliseconds(), err, req.ApiKey, "", "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.GetAllApps(ctx, rpcCtx, startTS, req.ApiKey)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) GetAllKeyIDs(ctx context.Context, req *connector_api.GetAllKeyIDsRequest) (*connector_api.GetAllKeyIDsResponse, error) {
	startTS := time.Now()
	var resp connector_api.GetAllKeyIDsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetAllKeyIDs", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetAllKeyIDs", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		s.consoleService.Error(rpcCtx, "bad request", "GetAllKeyIDs", time.Since(startTS).Milliseconds(), err, req.ApiKey, "", "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.GetAllKeyIDs(ctx, rpcCtx, startTS, req.ApiKey)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) ListApiKey(ctx context.Context, req *connector_api.ListApiKeyRequest) (*connector_api.ListApiKeyResponse, error) {
	startTS := time.Now()
	var resp connector_api.ListApiKeyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ListApiKey", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ListApiKey", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "ListApiKey", time.Since(startTS).Milliseconds(), err, req.ApiKey, "", "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.ListApiKey(ctx, rpcCtx, startTS, req.ApiKey)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) AddApiKey(ctx context.Context, req *connector_api.AddApiKeyRequest) (*connector_api.AddApiKeyResponse, error) {
	startTS := time.Now()
	var resp connector_api.AddApiKeyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AddApiKey", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AddApiKey", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "AddApiKey", time.Since(startTS).Milliseconds(), err, req.ApiKey, req.App, req.KeyID)
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	err := s.consoleService.AddApiKey(ctx, rpcCtx, startTS, req.ApiKey, req.App, req.KeyID, req.Name, req.Permission)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}

func (s *ConnectorConsoleServer) RemoveApiKey(ctx context.Context, req *connector_api.RemoveApiKeyRequest) (*connector_api.RemoveApiKeyResponse, error) {
	startTS := time.Now()
	var resp connector_api.RemoveApiKeyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RemoveApiKey", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RemoveApiKey", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "RemoveApiKey", time.Since(startTS).Milliseconds(), err, req.ApiKey, "", "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	apiKeyObj, err := s.daoManager.ApiKeyDAO.GetByID(ctx, req.Id)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			s.consoleService.Error(rpcCtx, "internal error", "RemoveApiKey", time.Since(startTS).Milliseconds(), err, req.ApiKey, "", "")
			resp.Code = response.StatusCodeEmptyParameters
			resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
			return &resp, nil
		} else {
			s.consoleService.Error(rpcCtx, "bad request", "RemoveApiKey", time.Since(startTS).Milliseconds(), err, req.ApiKey, "", "")
			resp.Code = response.StatusCodeEmptyParameters
			resp.Message = rpcCtx.Localizer.Localize("ParamErrMsg_IDIsInvalid")
			return &resp, nil
		}
	}
	appErr := s.consoleService.RemoveApiKey(ctx, rpcCtx, startTS, req.ApiKey, req.Id, apiKeyObj)
	if err != nil {
		resp.Code = appErr.Code
		resp.Message = appErr.Message
		resp.DebugMessage = appErr.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}

func (s *ConnectorConsoleServer) ValidateApiKey(ctx context.Context, req *connector_api.ValidateApiKeyRequest) (*connector_api.ValidateApiKeyResponse, error) {
	startTS := time.Now()
	var resp connector_api.ValidateApiKeyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ValidateApiKey", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ValidateApiKey", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.ValidateApiKey(ctx, rpcCtx, startTS, req.App, req.ApiKey, req.Role)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *ConnectorConsoleServer) GetManagePlatformLogs(ctx context.Context, req *connector_api.GetManagePlatformLogsRequest) (*connector_api.GetManagePlatformLogsResponse, error) {
	startTS := time.Now()
	var resp connector_api.GetManagePlatformLogsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetManagePlatformLogs", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetManagePlatformLogs", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		s.consoleService.Error(rpcCtx, "bad request", "GetManagePlatformLogs", time.Since(startTS).Milliseconds(), err, req.ApiKey, "", "")
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, err := s.consoleService.GetManagePlatformLogs(ctx, rpcCtx, startTS, req.ApiKey, req.PageSize, req.PageNumber)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}
