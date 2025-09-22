package server

import (
	"context"

	"github.com/go-playground/validator/v10"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"github.com/nextsurfer/monitor/internal/common/redis"
	"github.com/nextsurfer/monitor/internal/service"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type MonitorServer struct {
	*service.MonitorService
	env            gutil.APPEnvType
	logger         *zap.Logger
	validator      *validator.Validate
	monitorService *service.MonitorService
}

func NewMonitorServer(ctx context.Context, env gutil.APPEnvType, logger *zap.Logger, redisOption *redis.Option, MongoDB *mongo.Database, validator *validator.Validate) (*MonitorServer, error) {
	s := &MonitorServer{
		env:       env,
		logger:    logger,
		validator: validator,
	}
	// services
	monitorService, err := service.NewMonitorService(ctx, logger, redisOption, MongoDB)
	if err != nil {
		return nil, err
	}
	s.monitorService = monitorService
	return s, nil
}
