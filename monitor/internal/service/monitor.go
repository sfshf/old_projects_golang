package service

import (
	"context"

	"github.com/nextsurfer/monitor/internal/common/redis"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type MonitorService struct {
	*MarketService
	*CronService
	*Web3Service

	Logger      *zap.Logger
	redisOption *redis.Option
	MongoDB     *mongo.Database
}

func NewMonitorService(ctx context.Context, logger *zap.Logger, redisOption *redis.Option, MongoDB *mongo.Database) (*MonitorService, error) {
	monitorService := &MonitorService{
		Logger:      logger,
		MongoDB:     MongoDB,
		redisOption: redisOption,
	}
	var err error
	// market service
	monitorService.MarketService, err = NewMarketService(ctx, monitorService)
	if err != nil {
		return nil, err
	}
	// cron service
	monitorService.CronService, err = NewCronService(ctx, monitorService)
	if err != nil {
		return nil, err
	}
	// web3 service
	monitorService.Web3Service, err = NewWeb3Service(ctx, monitorService)
	if err != nil {
		return nil, err
	}
	return monitorService, nil
}
