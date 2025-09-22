package gateway

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	gdao "github.com/nextsurfer/ground/pkg/dao"
	"github.com/nextsurfer/ground/pkg/log"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"github.com/nextsurfer/oracle/internal/dao"
	"github.com/nextsurfer/oracle/internal/gateway/server"
	"go.uber.org/zap"
)

type GatewayApp struct {
	Env                 gutil.APPEnvType
	Name                string
	Host                string
	HttpPort            int // http port
	TlsPort             int
	MysqlPrometheusPort int // = http port + 2
	RpcPort             int
	Logger              *zap.Logger
	RedisClient         *redis.Client
	DaoManager          *dao.Manager
	gatewayServer       *server.GatewayServer
	Validator           *validator.Validate
}

func NewGatewayApp(ctx context.Context, appEnv int, appName, host string, httpPort, tlsPort, rpcPort int, mysqlDNS, redisDNS, tomlPath string) (*GatewayApp, error) {
	// env
	env := gutil.EnvForInt(appEnv)
	mysqlPrometheusPort := rpcPort + 2
	app := &GatewayApp{
		Env:                 env,
		Name:                appName,
		Host:                host,
		HttpPort:            httpPort,
		TlsPort:             tlsPort,
		RpcPort:             rpcPort,
		MysqlPrometheusPort: mysqlPrometheusPort,
	}
	// validator
	app.Validator = validator.New()
	// dao manager
	option := gdao.NewOption(mysqlDNS, app.Name, app.MysqlPrometheusPort, env)
	daoManager := dao.NewManager(option)
	app.DaoManager = daoManager
	// logger
	logOptions := log.NewOptions(app.Name, env, true)
	logOptions.EnvDev = app.Env == gutil.AppEnvDEV
	app.Logger = log.New(logOptions)
	// redis
	rdbOpt, err := redis.ParseURL(redisDNS)
	if err != nil {
		return nil, err
	}
	rdbClient := redis.NewClient(rdbOpt)
	app.RedisClient = rdbClient
	// http server
	gatewayHttpServer, err := server.NewGatewayServer(ctx, app.Name, app.Env, app.Logger, app.DaoManager, app.RedisClient, host, httpPort, tlsPort, rpcPort, tomlPath, app.Validator)
	if err != nil {
		return nil, err
	}
	app.gatewayServer = gatewayHttpServer
	return app, nil
}

func (app *GatewayApp) Run(ctx context.Context) error {
	// start servers
	if err := app.gatewayServer.Run(ctx); err != nil {
		app.Logger.Error("Gateway app start http server", zap.NamedError("appError", err))
		return err
	}
	app.Logger.Info("Gateway app start http server success")
	return nil
}

func (app *GatewayApp) Stop(ctx context.Context) error {
	// stop servers
	if err := app.gatewayServer.Stop(ctx); err != nil {
		app.Logger.Error("Gateway app stop servers", zap.NamedError("appError", err))
	}
	app.Logger.Info("Gateway app stop gracefully")
	return nil
}
