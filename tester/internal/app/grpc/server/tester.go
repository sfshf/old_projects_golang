package server

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	tester_api "github.com/nextsurfer/tester/api"
	"github.com/nextsurfer/tester/api/response"
	"github.com/nextsurfer/tester/internal/app/grpc/service"
	"github.com/nextsurfer/tester/internal/pkg/dao"
	"github.com/nextsurfer/tester/internal/pkg/redis"
	"github.com/nextsurfer/tester/internal/pkg/riki"
	"github.com/nextsurfer/tester/internal/pkg/simplejson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type TesterServer struct {
	tester_api.UnimplementedTesterServiceServer
	service         *service.TesterService
	appID           string
	env             gutil.APPEnvType
	logger          *zap.Logger
	localizeManager *localize.Manager
	validator       *validator.Validate
}

func NewTesterServer(ctx context.Context, appID string, env gutil.APPEnvType, logger *zap.Logger, localizeManager *localize.Manager, validator *validator.Validate, mongoDB *mongo.Database, daoManager *dao.Manager, redisOption *redis.Option) (*TesterServer, error) {
	s := &TesterServer{
		appID:           appID,
		env:             env,
		logger:          logger,
		localizeManager: localizeManager,
		validator:       validator,
	}
	// market service
	DoomConsoleService, err := service.NewTesterService(ctx, logger, mongoDB, daoManager, redisOption)
	if err != nil {
		return nil, err
	}
	s.service = DoomConsoleService
	return s, nil
}

func (s *TesterServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *TesterServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *TesterServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

func (s *TesterServer) UpdateAPITestcase(ctx context.Context, req *tester_api.UpdateAPITestcaseRequest) (*tester_api.UpdateAPITestcaseResponse, error) {
	var resp tester_api.UpdateAPITestcaseResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UpdateAPITestcase", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UpdateAPITestcase", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleWrite); err != nil {
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
	appError := s.service.UpdateAPITestcase(ctx, rpcCtx, req)
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

func (s *TesterServer) GetApps(ctx context.Context, req *tester_api.GetAppsRequest) (*tester_api.GetAppsResponse, error) {
	var resp tester_api.GetAppsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetApps", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetApps", &resp) }()
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
	data, appError := s.service.GetApps(ctx, rpcCtx, req)
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

func (s *TesterServer) GetAPITestcases(ctx context.Context, req *tester_api.GetAPITestcasesRequest) (*tester_api.GetAPITestcasesResponse, error) {
	var resp tester_api.GetAPITestcasesResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetAPITestcases", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetAPITestcases", &resp) }()
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
	data, appError := s.service.GetAPITestcases(ctx, rpcCtx, req)
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

func (s *TesterServer) GetMysqlInfo(ctx context.Context, req *tester_api.GetMysqlInfoRequest) (*tester_api.GetMysqlInfoResponse, error) {
	var resp tester_api.GetMysqlInfoResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetMysqlInfo", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetMysqlInfo", &resp) }()
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
	data, appError := s.service.GetMysqlInfo(ctx, rpcCtx, req)
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

func (s *TesterServer) GetMongoInfo(ctx context.Context, req *tester_api.GetMongoInfoRequest) (*tester_api.GetMongoInfoResponse, error) {
	var resp tester_api.GetMongoInfoResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetMongoInfo", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetMongoInfo", &resp) }()
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
	data, appError := s.service.GetMongoInfo(ctx, rpcCtx, req)
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

func (s *TesterServer) GetBtcDiff(ctx context.Context, req *tester_api.GetBtcDiffRequest) (*tester_api.GetBtcDiffResponse, error) {
	var resp tester_api.GetBtcDiffResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetBtcDiff", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetBtcDiff", &resp) }()
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
	data, appError := s.service.GetBtcDiff(ctx, rpcCtx, req)
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

func (s *TesterServer) GetTxsMempoolInfo(ctx context.Context, req *tester_api.GetTxsMempoolInfoRequest) (*tester_api.GetTxsMempoolInfoResponse, error) {
	var resp tester_api.GetTxsMempoolInfoResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetTxsMempoolInfo", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetTxsMempoolInfo", &resp) }()
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
	data, appError := s.service.GetTxsMempoolInfo(ctx, rpcCtx, req)
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

func (s *TesterServer) GetSlarkRegistrationCaptchas(ctx context.Context, req *tester_api.GetSlarkRegistrationCaptchasRequest) (*tester_api.GetSlarkRegistrationCaptchasResponse, error) {
	var resp tester_api.GetSlarkRegistrationCaptchasResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetSlarkRegistrationCaptchas", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetSlarkRegistrationCaptchas", &resp) }()
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
	data, appError := s.service.GetSlarkRegistrationCaptchas(ctx, rpcCtx, req)
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

func (s *TesterServer) GetSlarkLoginCaptchas(ctx context.Context, req *tester_api.GetSlarkLoginCaptchasRequest) (*tester_api.GetSlarkLoginCaptchasResponse, error) {
	var resp tester_api.GetSlarkLoginCaptchasResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetSlarkLoginCaptchas", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetSlarkLoginCaptchas", &resp) }()
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
	data, appError := s.service.GetSlarkLoginCaptchas(ctx, rpcCtx, req)
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

func (s *TesterServer) UploadApp(ctx context.Context, req *tester_api.UploadAppRequest) (*tester_api.UploadAppResponse, error) {
	var resp tester_api.UploadAppResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UploadApp", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UploadApp", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleWrite); err != nil {
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
	appError := s.service.UploadApp(ctx, rpcCtx, req)
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

func (s *TesterServer) GetUploadedApps(ctx context.Context, req *tester_api.GetUploadedAppsRequest) (*tester_api.GetUploadedAppsResponse, error) {
	var resp tester_api.GetUploadedAppsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetUploadedApps", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetUploadedApps", &resp) }()
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
	data, appError := s.service.GetUploadedApps(ctx, rpcCtx, req)
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

func (s *TesterServer) GetMessageNotificationConfig(ctx context.Context, req *tester_api.GetMessageNotificationConfigRequest) (*tester_api.GetMessageNotificationConfigResponse, error) {
	var resp tester_api.GetMessageNotificationConfigResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetMessageNotificationConfig", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetMessageNotificationConfig", &resp) }()
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
	data, appError := s.service.GetMessageNotificationConfig(ctx, rpcCtx, req)
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

func (s *TesterServer) UpdateMessageNotificationConfig(ctx context.Context, req *tester_api.UpdateMessageNotificationConfigRequest) (*tester_api.UpdateMessageNotificationConfigResponse, error) {
	var resp tester_api.UpdateMessageNotificationConfigResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UpdateMessageNotificationConfig", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UpdateMessageNotificationConfig", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleWrite); err != nil {
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
	appError := s.service.UpdateMessageNotificationConfig(ctx, rpcCtx, req)
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

func (s *TesterServer) SendMessageNotification(ctx context.Context, req *tester_api.SendMessageNotificationRequest) (*tester_api.SendMessageNotificationResponse, error) {
	var resp tester_api.SendMessageNotificationResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "SendMessageNotification", req)
	defer func() { s.deferLogResponseData(rpcCtx, "SendMessageNotification", &resp) }()
	// validate api key
	if err := riki.ValidateApiKey(ctx, rpcCtx, s.appID, req.ApiKey, riki.RoleWrite); err != nil {
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
	appError := s.service.SendMessageNotification(ctx, rpcCtx, req)
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

func (s *TesterServer) GetMessageNotificationLogs(ctx context.Context, req *tester_api.GetMessageNotificationLogsRequest) (*tester_api.GetMessageNotificationLogsResponse, error) {
	var resp tester_api.GetMessageNotificationLogsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetMessageNotificationLogs", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetMessageNotificationLogs", &resp) }()
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
	data, appError := s.service.GetMessageNotificationLogs(ctx, rpcCtx, req)
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

func (s *TesterServer) GetPrivacyEmailAccounts(ctx context.Context, req *tester_api.GetPrivacyEmailAccountsRequest) (*tester_api.GetPrivacyEmailAccountsResponse, error) {
	var resp tester_api.GetPrivacyEmailAccountsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetPrivacyEmailAccounts", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetPrivacyEmailAccounts", &resp) }()
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
	data, appError := s.service.GetPrivacyEmailAccounts(ctx, rpcCtx, req)
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
