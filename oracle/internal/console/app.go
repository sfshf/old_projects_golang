package console

import (
	"context"
	stdlog "log"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	consulApi "github.com/hashicorp/consul/api"
	gdao "github.com/nextsurfer/ground/pkg/dao"
	"github.com/nextsurfer/ground/pkg/log"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"github.com/nextsurfer/oracle/internal/console/server"
	"github.com/nextsurfer/oracle/internal/dao"
	"go.uber.org/zap"
)

// console app
type ConsoleApp struct {
	Env                 gutil.APPEnvType
	Name                string
	Host                string
	HttpPort            int // http port
	GrpcPort            int
	MysqlPrometheusPort int // = http port + 2
	Logger              *zap.Logger
	RedisClient         *redis.Client
	DaoManager          *dao.Manager
	ConsoleServer       *server.ConsoleServer
	consulClient        *consulApi.Client
	WebPath             string
	Validator           *validator.Validate
}

func NewConsoleApp(ctx context.Context, appEnv int, appName string, host string, httpPort, grpcPort int, mysqlDNS, redisDNS, tomlPath string, consoleWebPath string) (*ConsoleApp, error) {
	mysqlPrometheusPort := httpPort + 2
	if mysqlPrometheusPort == grpcPort {
		mysqlPrometheusPort += 1
	}
	app := &ConsoleApp{
		Env:                 gutil.EnvForInt(appEnv),
		Name:                appName,
		Host:                host,
		HttpPort:            httpPort,
		GrpcPort:            grpcPort,
		MysqlPrometheusPort: mysqlPrometheusPort,
		WebPath:             consoleWebPath,
	}
	// consul client
	config := consulApi.DefaultConfig()
	client, err := consulApi.NewClient(config)
	if err != nil {
		stdlog.Fatalf("internal error: %v", err)
	}
	app.consulClient = client
	// validator
	app.Validator = validator.New()
	// dao manager
	option := gdao.NewOption(mysqlDNS, app.Name, app.MysqlPrometheusPort, app.Env)
	daoManager := dao.NewManager(option)
	app.DaoManager = daoManager
	// logger
	logOptions := log.NewOptions(app.Name, app.Env, true)
	logOptions.EnvDev = app.Env == gutil.AppEnvDEV
	app.Logger = log.New(logOptions)
	// redis
	rdbOpt, err := redis.ParseURL(redisDNS)
	if err != nil {
		return nil, err
	}
	rdbClient := redis.NewClient(rdbOpt)
	app.RedisClient = rdbClient
	// console server
	consoleServer, err := server.NewConsoleServer(ctx, app.Name, app.Env, app.Logger, app.DaoManager, app.RedisClient, host, grpcPort, httpPort, tomlPath, app.Validator, app.WebPath)
	if err != nil {
		return nil, err
	}
	app.ConsoleServer = consoleServer
	return app, nil
}

func (s *ConsoleApp) Run(ctx context.Context) {
	s.ConsoleServer.Run(ctx)
}

func (s *ConsoleApp) Stop(ctx context.Context) error {
	return s.ConsoleServer.Stop(ctx)
}
