package servers

import (
	"context"
	"fmt"

	alchemist_api "github.com/nextsurfer/alchemist/api"
	"github.com/nextsurfer/alchemist/api/response"
	"github.com/nextsurfer/alchemist/internal/app/grpc/services"
	"github.com/nextsurfer/alchemist/internal/pkg/dao"
	"github.com/nextsurfer/alchemist/internal/pkg/redis"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	slark_response "github.com/nextsurfer/slark/api/response"
	slark_grpc "github.com/nextsurfer/slark/pkg/grpc"
	"go.uber.org/zap"
)

type TestServer struct {
	alchemist_api.UnimplementedTestServiceServer
	app             string
	env             gutil.APPEnvType
	logger          *zap.Logger
	daoManager      *dao.Manager
	redisOption     *redis.Option
	localizeManager *localize.Manager
	testService     *services.TestService
}

func NewTestServer(app string, env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option, localizeManager *localize.Manager) *TestServer {
	testService := services.NewTestService(app, env, logger, daoManager, redisOption)
	return &TestServer{
		app:             app,
		env:             env,
		logger:          logger,
		daoManager:      daoManager,
		redisOption:     redisOption,
		localizeManager: localizeManager,
		testService:     testService,
	}
}

func (s *TestServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *TestServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := util.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *TestServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := util.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

func (s *TestServer) ResetDeviceCheck(ctx context.Context, req *alchemist_api.ResetDeviceCheckRequest) (*alchemist_api.ResetDeviceCheckResponse, error) {
	var resp alchemist_api.ResetDeviceCheckResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ResetDeviceCheck", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ResetDeviceCheck", &resp) }()
	appError := s.testService.ResetDeviceCheck(ctx, rpcCtx, req.AppID, req.DeviceToken0, req.DeviceToken1)
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

func (s *TestServer) RemoveFraudRecord(ctx context.Context, req *alchemist_api.Empty) (*alchemist_api.RemoveFraudRecordResponse, error) {
	var resp alchemist_api.RemoveFraudRecordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RemoveFraudRecord", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RemoveFraudRecord", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			return &resp, nil
		}
	}
	userInfo := loginInfo.Data
	appError := s.testService.RemoveFraudRecord(ctx, rpcCtx, userInfo.UserID)
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
