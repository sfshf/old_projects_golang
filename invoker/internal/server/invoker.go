package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	invoker_api "github.com/nextsurfer/invoker/api"
	"github.com/nextsurfer/invoker/internal/common/simplejson"
	"github.com/nextsurfer/invoker/internal/dao"
	"github.com/nextsurfer/invoker/internal/service"
	"go.uber.org/zap"
)

type InvokerServer struct {
	invoker_api.UnimplementedAdminServiceServer
	invoker_api.UnimplementedSiteServiceServer
	invoker_api.UnimplementedUserServiceServer
	*service.InvokerService

	env             gutil.APPEnvType
	appID           string
	logger          *zap.Logger
	localizeManager *localize.Manager
	validator       *validator.Validate
	DaoManager      *dao.Manager
}

func NewInvokerServer(ctx context.Context, env gutil.APPEnvType, appID string, logger *zap.Logger, redisClient *redis.Client, localizeManager *localize.Manager, validator *validator.Validate, daoManager *dao.Manager) (*InvokerServer, error) {
	s := &InvokerServer{
		env:             env,
		appID:           appID,
		logger:          logger,
		localizeManager: localizeManager,
		validator:       validator,
		DaoManager:      daoManager,
	}
	invokerService, err := service.NewInvokerService(ctx, appID, localizeManager, logger, redisClient, daoManager, validator)
	if err != nil {
		return nil, err
	}
	s.InvokerService = invokerService
	return s, nil
}

func (s *InvokerServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *InvokerServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *InvokerServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

type OracleFieldType struct {
	Duration int64 `json:"duration"` // millisecond
}

func (s *InvokerServer) oracleField(startTS time.Time) string {
	data, _ := json.Marshal(OracleFieldType{Duration: int64(time.Since(startTS) / time.Millisecond)})
	return string(data)
}
