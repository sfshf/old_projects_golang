package grpc

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	gdao "github.com/nextsurfer/ground/pkg/dao"
	"github.com/nextsurfer/ground/pkg/log"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/ground/pkg/util"
	word_api "github.com/nextsurfer/word/api"
	grpcservers "github.com/nextsurfer/word/internal/app/grpc/servers"
	"github.com/nextsurfer/word/internal/pkg/dao"
	"github.com/nextsurfer/word/internal/pkg/redis"
	"go.uber.org/zap"
)

// Application app
type Application struct {
	Name                string
	Port                int // grpc port
	GrpcPrometheusPort  int // = grpc port + 1
	MysqlPrometheusPort int // = gprc port + 2
	Host                string
	Env                 util.APPEnvType
	Server              *rpc.Server // TODO consul -> rpc
	Logger              *zap.Logger
	DaoManager          *dao.Manager
	RedisOption         *redis.Option
	Validator           *validator.Validate
}

// NewApplication create application
func NewApplication(name string, port int, host string, appEnv int, redisDNS string, mysqlDNS string, tomlPath string) (*Application, error) {
	var err error
	app := &Application{
		Name: name,
		Port: port,
		Host: host,
	}
	// env
	app.Env = util.EnvForInt(appEnv)
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
	// validator
	app.Validator = validator.New()
	// redis
	redisOption, err := redis.NewOption(redisDNS, app.Logger)
	if err != nil {
		return nil, err
	}
	app.RedisOption = redisOption
	// dao
	app.DaoManager = dao.NewManager(gdao.NewOption(mysqlDNS, app.Name, app.MysqlPrometheusPort, app.Env))
	// grpc server
	app.Server, err = rpc.NewServer(app.Name, app.Env, host, port, rpc.NewTracer(name, app.Env))
	if err != nil {
		logger.Error("grpc server create error", zap.NamedError("appError", err))
		return nil, err
	}
	// register servers
	app.RegisterServers()
	// load i18n files
	err = app.LoadMessageFiles(tomlPath)
	if err != nil {
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

func (app *Application) RegisterServers() {
	word_api.RegisterWordServiceServer(app.Server.GrpcServer(), grpcservers.NewWordServer(app.Env, app.Logger, app.DaoManager, app.RedisOption, app.Server.LocolizerManager, app.Validator))
	app.Logger.Info("api.RegisterWordServiceServer success")
}

// Start call server to start
func (app *Application) Start() error {
	app.Logger.Info(
		"application start",
		zap.String("appName", app.Name),
		zap.Int("port", app.Port),
		zap.String("host", app.Host),
		zap.String("env", util.LabelForEnv(app.Env)),
	)
	return app.Server.Start()
}

// Stop  call server to stop
func (app *Application) Stop() error {
	return app.Server.Stop()
}
