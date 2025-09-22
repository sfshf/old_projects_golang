package grpc

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/nextsurfer/connector/internal/pkg/dao"
	"github.com/nextsurfer/connector/internal/pkg/keystore"
	"github.com/nextsurfer/connector/internal/pkg/redis"
	"github.com/nextsurfer/connector/internal/pkg/util"
	gdao "github.com/nextsurfer/ground/pkg/dao"
	"github.com/nextsurfer/ground/pkg/log"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
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
	env := gutil.EnvForInt(appEnv)
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
	// keystore address
	keyStoreAddr := strings.TrimSpace(os.Getenv("KEY_STORE_ADDRESS"))
	if keyStoreAddr == "" {
		logger.Fatal("must set env variable for 'KEY_STORE_ADDRESS'")
	}
	keyStoreIpPort := strings.Split(keyStoreAddr, ":")
	if len(keyStoreIpPort) != 2 {
		logger.Fatal("'KEY_STORE_ADDRESS' is invalid")
	}
	keyStoreIp := strings.TrimSpace(keyStoreIpPort[0])
	if net.ParseIP(keyStoreIp) == nil {
		logger.Fatal("'KEY_STORE_ADDRESS' ip is invalid")
	}
	keyStorePort := strings.TrimSpace(keyStoreIpPort[1])
	if _, err = strconv.Atoi(keyStorePort); err != nil {
		logger.Fatal("'KEY_STORE_ADDRESS' port is invalid", zap.NamedError("appError", err))
	}
	// init config
	util.InitConfig(fmt.Sprintf("%s:%s", keyStoreIp, keyStorePort))
	// upsert connector password
	if err := keystore.UpsertConnectorPassword(daoManager, logger); err != nil {
		return nil, err
	}
	// load all app keys
	if err := util.LoadAllAppKeys(daoManager); err != nil {
		logger.Fatal("LoadAllAppKeys", zap.NamedError("appError", err))
	}
	for _, k := range util.ConfigInfo().AppKeys {
		logger.Info("load app key", zap.String("app", k.App), zap.String("keyID", k.KeyID), zap.String("key", k.ApiKey), zap.String("permission", k.Permission))
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
