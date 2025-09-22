package service

import (
	"context"

	"github.com/go-redis/redis/v8"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/nextsurfer/ground/pkg/localize"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"github.com/nextsurfer/oracle/internal/dao"
	"go.uber.org/zap"
)

type ConsoleService struct {
	*GatewayNodeService
	*ApplicationService
	*ServiceService
	*AcmeResourceService
	*HostnameService
	*RateLimitService
	*AlarmEmailService
	*ProtocolService
	*CronService

	AppID           string
	LocalizeManager *localize.Manager
	ConsulClient    *consulApi.Client
	Env             gutil.APPEnvType
	Logger          *zap.Logger
	RedisClient     *redis.Client
	DaoManager      *dao.Manager
}

func NewConsoleService(ctx context.Context, env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisClient *redis.Client, appID string, localizeManager *localize.Manager) (*ConsoleService, error) {
	s := &ConsoleService{
		Env:             env,
		Logger:          logger,
		DaoManager:      daoManager,
		RedisClient:     redisClient,
		AppID:           appID,
		LocalizeManager: localizeManager,
	}
	// GatewayNodeService
	s.GatewayNodeService = NewGatewayNodeService(ctx, s)
	// ApplicationService
	s.ApplicationService = NewApplicationService(ctx, s)
	// ServiceService
	s.ServiceService = NewServiceService(ctx, s)
	// AcmeResourceService
	s.AcmeResourceService = NewAcmeResourceService(ctx, s)
	// HostnameService
	s.HostnameService = NewHostnameService(ctx, s)
	// RateLimitService
	s.RateLimitService = NewRateLimitService(ctx, s)
	// AlarmEmailService
	s.AlarmEmailService = NewAlarmEmailService(ctx, s)
	// ProtocolService
	s.ProtocolService = NewProtocolService(ctx, s)
	// consul client
	config := consulApi.DefaultConfig()
	client, err := consulApi.NewClient(config)
	if err != nil {
		return nil, err
	}
	s.ConsulClient = client
	// cron service
	cronService, err := NewCronService(ctx, s)
	if err != nil {
		return nil, err
	}
	s.CronService = cronService
	return s, nil
}
