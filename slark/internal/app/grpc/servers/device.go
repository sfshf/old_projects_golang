package servers

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	slark_api "github.com/nextsurfer/slark/api"
	"github.com/nextsurfer/slark/api/response"
	"github.com/nextsurfer/slark/internal/app/grpc/services"
	"github.com/nextsurfer/slark/internal/pkg/dao"
	"github.com/nextsurfer/slark/internal/pkg/redis"
	"github.com/nextsurfer/slark/internal/pkg/simplejson"
	"go.uber.org/zap"
)

// DeviceServer , there are some logic about device in server layer.
type DeviceServer struct {
	slark_api.UnimplementedDeviceServiceServer
	env             gutil.APPEnvType
	logger          *zap.Logger
	daoManager      *dao.Manager
	redisOption     *redis.Option
	localizeManager *localize.Manager
	deviceService   *services.DeviceService
	validator       *validator.Validate
}

// NewDeviceServer is factory function
func NewDeviceServer(env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option, localizeManager *localize.Manager, validator *validator.Validate) *DeviceServer {
	deviceService := services.NewDeviceService(env, logger, daoManager, redisOption)
	return &DeviceServer{
		env:             env,
		logger:          logger,
		daoManager:      daoManager,
		redisOption:     redisOption,
		localizeManager: localizeManager,
		deviceService:   deviceService,
		validator:       validator,
	}
}

func (s *DeviceServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *DeviceServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *DeviceServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

func (s *DeviceServer) UploadDeviceInfo(ctx context.Context, req *slark_api.DeviceRequest) (*slark_api.DeviceEmptyResponse, error) {
	var resp slark_api.DeviceEmptyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UploadDeviceInfo", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UploadDeviceInfo", &resp) }()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	appError := s.deviceService.UploadDeviceInfo(ctx, rpcCtx, req)
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
