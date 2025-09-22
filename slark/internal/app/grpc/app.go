package grpc

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	gdao "github.com/nextsurfer/ground/pkg/dao"
	"github.com/nextsurfer/ground/pkg/log"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	slark_api "github.com/nextsurfer/slark/api"
	grpcservers "github.com/nextsurfer/slark/internal/app/grpc/servers"
	"github.com/nextsurfer/slark/internal/pkg/dao"
	"github.com/nextsurfer/slark/internal/pkg/redis"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
)

// Application app
type Application struct {
	Env                 gutil.APPEnvType
	Name                string
	Port                int // grpc port
	GrpcPrometheusPort  int // = grpc port + 1
	MysqlPrometheusPort int // = gprc port + 2
	Host                string
	Server              *rpc.Server
	Logger              *zap.Logger
	DaoManager          *dao.Manager
	RedisOption         *redis.Option
	Validator           *validator.Validate
	MongoDB             *mongo.Database
}

// NewApplication create application
func NewApplication(ctx context.Context, name string, port int, host string, appEnv int, redisDNS string, mysqlDNS string, tomlPath, mongodbUri string) (*Application, error) {
	var err error
	app := &Application{
		Name: name,
		Host: host,
		Port: port,
	}
	// env
	app.Env = gutil.EnvForInt(appEnv)
	// logger
	logOptions := log.NewOptions(app.Name, app.Env, true)
	logger := log.New(logOptions)
	app.Logger = logger
	// mongo db
	if app.Env == gutil.AppEnvDEV || app.Env == gutil.AppEnvPPE {
		uri, err := url.Parse(mongodbUri)
		if err != nil {
			return nil, err
		}
		mgoCli, err := mongo.Connect(options.Client().ApplyURI(mongodbUri))
		if err != nil {
			return nil, err
		}
		if err := mgoCli.Ping(ctx, nil); err != nil {
			return nil, err
		}
		dbName := uri.Path[1:]
		if dbName == "" {
			dbName = "slark"
		}
		app.MongoDB = mgoCli.Database(dbName)
	}
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

func (app *Application) RegisterServers() {
	server := app.Server.GrpcServer()
	if app.Env == gutil.AppEnvDEV || app.Env == gutil.AppEnvPPE {
		slark_api.RegisterTestServiceServer(server, grpcservers.NewTestServer(app.Logger, app.DaoManager, app.RedisOption, app.Server.LocolizerManager))
	}
	slark_api.RegisterUserServiceServer(server, grpcservers.NewUserServer(app.Env, app.Logger, app.DaoManager, app.RedisOption, app.Server.LocolizerManager, app.Validator, app.MongoDB))
	slark_api.RegisterDeviceServiceServer(server, grpcservers.NewDeviceServer(app.Env, app.Logger, app.DaoManager, app.RedisOption, app.Server.LocolizerManager, app.Validator))
	reflection.Register(server)
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
