package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	invoker_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/internal/common/simplejson"
	"github.com/nextsurfer/pswds_backend/internal/dao"
	"github.com/nextsurfer/pswds_backend/internal/service"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type PswdsServer struct {
	invoker_api.UnimplementedPasswordServiceServer
	*service.PswdsService

	env             gutil.APPEnvType
	appID           string
	logger          *zap.Logger
	localizeManager *localize.Manager
	validator       *validator.Validate
	DaoManager      *dao.Manager
}

func NewPswdsServer(ctx context.Context, env gutil.APPEnvType, appID string, logger *zap.Logger, redisClient *redis.Client, localizeManager *localize.Manager, validator *validator.Validate, daoManager *dao.Manager) (*PswdsServer, error) {
	s := &PswdsServer{
		env:             env,
		appID:           appID,
		logger:          logger,
		localizeManager: localizeManager,
		validator:       validator,
		DaoManager:      daoManager,
	}
	pswdsService, err := service.NewPswdsService(ctx, appID, localizeManager, logger, redisClient, daoManager, validator)
	if err != nil {
		return nil, err
	}
	s.PswdsService = pswdsService
	return s, nil
}

func (s *PswdsServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *PswdsServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *PswdsServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

type OracleFieldType struct {
	Duration int64 `json:"duration"` // millisecond
}

func (s *PswdsServer) oracleField(startTS time.Time) string {
	dur := int64(time.Since(startTS) / time.Millisecond)
	if dur == 0 {
		dur = 1
	}
	data, _ := json.Marshal(OracleFieldType{Duration: dur})
	return string(data)
}
