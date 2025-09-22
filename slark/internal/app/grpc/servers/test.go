package servers

import (
	"context"

	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	slark_api "github.com/nextsurfer/slark/api"
	"github.com/nextsurfer/slark/api/response"
	"github.com/nextsurfer/slark/internal/app/grpc/services"
	"github.com/nextsurfer/slark/internal/pkg/dao"
	"github.com/nextsurfer/slark/internal/pkg/redis"
	"github.com/nextsurfer/slark/internal/pkg/simplejson"
	"go.uber.org/zap"
)

type TestServer struct {
	slark_api.UnimplementedTestServiceServer
	logger          *zap.Logger
	daoManager      *dao.Manager
	redisOption     *redis.Option
	localizeManager *localize.Manager
	testService     *services.TestService
}

func NewTestServer(logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option, localizeManager *localize.Manager) *TestServer {
	testService := services.NewTestService(logger, daoManager, redisOption)
	return &TestServer{
		logger:          logger,
		daoManager:      daoManager,
		redisOption:     redisOption,
		localizeManager: localizeManager,
		testService:     testService,
	}
}

func (s *TestServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	data, _ := simplejson.DigestToJson(request)
	rpcCtx.Logger.Info(method, zap.String("request_data", data))
}

func (s *TestServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	data, _ := simplejson.DigestToJson(response)
	rpcCtx.Logger.Info(method, zap.String("response_data", data))
}

func (s *TestServer) GetRegistrationEmailCaptchas(ctx context.Context, req *slark_api.Empty) (*slark_api.GetRegistrationEmailCaptchasResponse, error) {
	var resp slark_api.GetRegistrationEmailCaptchasResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetRegistrationEmailCaptchas", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetRegistrationEmailCaptchas", &resp) }()
	data, appError := s.testService.GetRegistrationEmailCaptchas(ctx, rpcCtx)
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

func (s *TestServer) GetLoginEmailCaptchas(ctx context.Context, req *slark_api.Empty) (*slark_api.GetLoginEmailCaptchasResponse, error) {
	var resp slark_api.GetLoginEmailCaptchasResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetLoginEmailCaptchas", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetLoginEmailCaptchas", &resp) }()
	data, appError := s.testService.GetLoginEmailCaptchas(ctx, rpcCtx)
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
