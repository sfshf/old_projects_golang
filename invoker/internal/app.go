package internal

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	gdao "github.com/nextsurfer/ground/pkg/dao"
	"github.com/nextsurfer/ground/pkg/log"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	invoker_api "github.com/nextsurfer/invoker/api"
	"github.com/nextsurfer/invoker/internal/dao"
	grpcserver "github.com/nextsurfer/invoker/internal/server"
	"go.uber.org/zap"
)

// Application app
type Application struct {
	Name                string
	Port                int // grpc port
	GrpcPrometheusPort  int
	MysqlPrometheusPort int
	Host                string
	Env                 gutil.APPEnvType
	Server              *rpc.Server // TODO consul -> rpc
	Logger              *zap.Logger
	RedisClient         *redis.Client
	Validator           *validator.Validate
	InvokerServer       *grpcserver.InvokerServer
	DaoManager          *dao.Manager
}

// NewApplication create application
func NewApplication(ctx context.Context, name string, port int, host string, appEnv int, redisDNS string, tomlPath, mysqlDNS string) (*Application, error) {
	var err error
	app := &Application{
		Name: name,
		Port: port,
		Host: host,
	}
	// env
	app.Env = gutil.EnvForInt(appEnv)
	// logger
	logOptions := log.NewOptions(app.Name, app.Env, true)
	logger := log.New(logOptions)
	app.Logger = logger
	// grpc metric port
	grpcMetricPortStr := os.Getenv("GRPC_METRIC_PORT")
	if grpcMetricPortStr == "" {
		logger.Fatal("must set env variable for 'GRPC_METRIC_PORT'")
	}
	grpcMetricPort, err := strconv.Atoi(grpcMetricPortStr)
	if err != nil {
		logger.Fatal(fmt.Sprintf("grpcPortStr is error : %s", grpcMetricPortStr))
	}
	app.GrpcPrometheusPort = grpcMetricPort
	// mysql metric port
	mysqlMetricPortStr := os.Getenv("MYSQL_METRIC_PORT")
	if mysqlMetricPortStr == "" {
		logger.Fatal("must set env variable for 'MYSQL_METRIC_PORT'")
	}
	mysqlMetricPort, err := strconv.Atoi(mysqlMetricPortStr)
	if err != nil {
		logger.Fatal(fmt.Sprintf("grpcPortStr is error : %s", mysqlMetricPortStr))
	}
	app.MysqlPrometheusPort = mysqlMetricPort
	// dao
	app.DaoManager = dao.NewManager(gdao.NewOption(mysqlDNS, app.Name, app.MysqlPrometheusPort, app.Env))
	// validator
	validate := validator.New()
	app.Validator = validate
	// redis
	rdbOpt, err := redis.ParseURL(redisDNS)
	if err != nil {
		return nil, err
	}
	rdbClient := redis.NewClient(rdbOpt)
	app.RedisClient = rdbClient
	// grpc server
	app.Server, err = rpc.NewServer(app.Name, app.Env, host, port, rpc.NewTracer(name, app.Env))
	if err != nil {
		logger.Error("grpc server create error", zap.NamedError("appError", err))
		return nil, err
	}
	// register servers
	if err := app.RegisterServers(ctx); err != nil {
		return nil, err
	}
	// load i18n files
	if err := app.LoadMessageFiles(tomlPath); err != nil {
		return nil, err
	}
	return app, nil
}

func (app *Application) LoadMessageFiles(tomlPath string) error {
	app.Logger.Info("i18n load toml files ", zap.String("path", tomlPath))
	files, err := ioutil.ReadDir(tomlPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".toml") {
			fullName := fmt.Sprintf("%s/%s", strings.TrimRight(tomlPath, "/"), f.Name())
			app.Logger.Info("i18n load file", zap.String("file path", fullName))
			app.Server.LocolizerManager.Bundle.MustLoadMessageFile(fullName)
		}

	}
	return nil
}

func (app *Application) RegisterServers(ctx context.Context) error {
	// register servers
	invokerServer, err := grpcserver.NewInvokerServer(
		ctx,
		app.Env,
		app.Name,
		app.Logger,
		app.RedisClient,
		app.Server.LocolizerManager,
		app.Validator,
		app.DaoManager,
	)
	if err != nil {
		return err
	}
	invoker_api.RegisterAdminServiceServer(
		app.Server.GrpcServer(),
		invokerServer)
	invoker_api.RegisterSiteServiceServer(
		app.Server.GrpcServer(),
		invokerServer)
	invoker_api.RegisterUserServiceServer(
		app.Server.GrpcServer(),
		invokerServer)
	app.InvokerServer = invokerServer
	app.Logger.Info("RegisterInvokerServiceServer success")
	return nil
}

// Start call server to start
func (app *Application) Start() error {
	app.Logger.Info(
		"application start",
		zap.String("appName", app.Name),
		zap.Int("port", app.Port),
		zap.String("host", app.Host),
		zap.String("env", gutil.LabelForEnv(app.Env)),
	)
	return app.Server.Start()
}

// Stop  call server to stop
func (app *Application) Stop() error {
	return app.Server.Stop()
}

// AwaitSignal wait user to kill the server
func (app *Application) AwaitSignal() {
	app.Server.AwaitSignal()
}
