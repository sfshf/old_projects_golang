package servers

import (
	"context"

	"github.com/go-playground/validator/v10"
	alchemist_notification_api "github.com/nextsurfer/alchemist/api/notification"
	"github.com/nextsurfer/alchemist/api/response"
	"github.com/nextsurfer/alchemist/internal/app/notification/services"
	"github.com/nextsurfer/alchemist/internal/pkg/dao"
	"github.com/nextsurfer/alchemist/internal/pkg/redis"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type AlchemistNotificationServer struct {
	alchemist_notification_api.UnimplementedAlchemistNotificationServiceServer
	env                          gutil.APPEnvType
	logger                       *zap.Logger
	daoManager                   *dao.Manager
	redisOption                  *redis.Option
	localizeManager              *localize.Manager
	alchemistNotificationService *services.AlchemistNotificationService
	validator                    *validator.Validate
}

func NewAlchemistNotificationServer(env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option, localizeManager *localize.Manager, validator *validator.Validate) *AlchemistNotificationServer {
	alchemistNotificationService := services.NewAlchemistNotificationService(env, logger, daoManager, redisOption)
	return &AlchemistNotificationServer{
		env:                          env,
		logger:                       logger,
		daoManager:                   daoManager,
		redisOption:                  redisOption,
		localizeManager:              localizeManager,
		alchemistNotificationService: alchemistNotificationService,
		validator:                    validator,
	}
}

func (s *AlchemistNotificationServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *AlchemistNotificationServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := util.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *AlchemistNotificationServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := util.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

// service methods --------------------------------------------------------------------------------------------

func (s *AlchemistNotificationServer) AppStoreNotificationsProd(ctx context.Context, req *alchemist_notification_api.AppStoreNotificationsProdRequest) (*alchemist_notification_api.EmptyResponse, error) {
	var resp alchemist_notification_api.EmptyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AppStoreNotificationsProd", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AppStoreNotificationsProd", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	err := s.alchemistNotificationService.AppStoreNotificationsProd(ctx, rpcCtx, req.SignedPayload)
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

func (s *AlchemistNotificationServer) AppStoreNotificationsTest(ctx context.Context, req *alchemist_notification_api.AppStoreNotificationsTestRequest) (*alchemist_notification_api.EmptyResponse, error) {
	var resp alchemist_notification_api.EmptyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "AppStoreNotificationsTest", req)
	defer func() { s.deferLogResponseData(rpcCtx, "AppStoreNotificationsTest", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	err := s.alchemistNotificationService.AppStoreNotificationsTest(ctx, rpcCtx, req.SignedPayload)
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
