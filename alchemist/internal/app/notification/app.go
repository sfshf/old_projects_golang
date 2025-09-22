package notification

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/nextsurfer/alchemist/internal/pkg/dao"
	"github.com/nextsurfer/alchemist/internal/pkg/redis"
	gdao "github.com/nextsurfer/ground/pkg/dao"
	"github.com/nextsurfer/ground/pkg/log"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/ground/pkg/util"
	"go.uber.org/zap"
)

// Application app
type Application struct {
	Name                string
	Port                int // grpc port
	GrpcPrometheusPort  int
	MysqlPrometheusPort int
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
	env := util.EnvForInt(appEnv)
	// logger
	logOptions := log.NewOptions(name, app.Env, true)
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
	daoManager := dao.NewManager(gdao.NewOption(mysqlDNS, app.Name, app.MysqlPrometheusPort, app.Env))
	app.DaoManager = daoManager
	// grpc server
	app.Server, err = rpc.NewServer(name, env, host, port, rpc.NewTracer(name, env))
	if err != nil {
		logger.Fatal("grpc server create error", zap.NamedError("appError", err))
	}
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

// AwaitSignal wait user to kill the server
func (app *Application) AwaitSignal() {
	app.Server.AwaitSignal()
}
