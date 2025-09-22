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
	doom_api "github.com/nextsurfer/doom-console/api"
	grpcserver "github.com/nextsurfer/doom-console/internal/app/grpc/server"
	"github.com/nextsurfer/ground/pkg/log"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	KongAddress         string
	Validator           *validator.Validate
	MongoDB             *mongo.Database
}

// NewApplication create application
func NewApplication(ctx context.Context, name, kongAddress string, port int, host string, appEnv int, tomlPath, mongodbUri string) (*Application, error) {
	var err error
	app := &Application{
		Name:        name,
		Port:        port,
		Host:        host,
		KongAddress: kongAddress,
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
	// validator
	app.Validator = validator.New()
	// grpc server
	app.Server, err = rpc.NewServer(app.Name, app.Env, host, port, rpc.NewTracer(name, app.Env))
	if err != nil {
		logger.Error("grpc server create error", zap.NamedError("appError", err))
		return nil, err
	}
	// mongo db
	uri, err := url.Parse(mongodbUri)
	if err != nil {
		return nil, err
	}
	mgoCli, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbUri))
	if err != nil {
		return nil, err
	}
	if err := mgoCli.Ping(ctx, nil); err != nil {
		return nil, err
	}
	dbName := uri.Path[1:]
	if dbName == "" {
		dbName = "doom"
	}
	app.MongoDB = mgoCli.Database(dbName)
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
	doomServer, err := grpcserver.NewDoomConsoleServer(
		ctx,
		app.Name,
		app.Env,
		app.Logger,
		app.Server.LocolizerManager,
		app.KongAddress,
		app.Validator,
		app.MongoDB,
	)
	if err != nil {
		return err
	}
	doom_api.RegisterDoomConsoleServiceServer(
		app.Server.GrpcServer(),
		doomServer)
	app.Logger.Info("RegisterDoomConsoleServiceServer success")
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
