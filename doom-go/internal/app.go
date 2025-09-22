package internal

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	connector_grpc "github.com/nextsurfer/connector/pkg/grpc"
	doom_api "github.com/nextsurfer/doom-go/api"
	"github.com/nextsurfer/doom-go/internal/common/config"
	"github.com/nextsurfer/doom-go/internal/common/eth"
	grpcserver "github.com/nextsurfer/doom-go/internal/server"
	"github.com/nextsurfer/ground/pkg/log"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Application app
type Application struct {
	Name               string
	Port               int // grpc port
	GrpcPrometheusPort int
	Host               string
	Env                gutil.APPEnvType
	Server             *rpc.Server // TODO consul -> rpc
	Logger             *zap.Logger
	RedisClient        *redis.Client
	ConnectorApiKey    string
	ConnectorKeyID     string
	Validator          *validator.Validate
	MongoDB            *mongo.Database
	DoomServer         *grpcserver.DoomServer
}

// NewApplication create application
func NewApplication(ctx context.Context, name, connectorApiKey, connectorKeyID string, port int, host string, appEnv int, redisDNS string, tomlPath, mongodbUri string) (*Application, error) {
	var err error
	app := &Application{
		Name:            name,
		Port:            port,
		Host:            host,
		ConnectorApiKey: connectorApiKey,
		ConnectorKeyID:  connectorKeyID,
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
	// validator
	validate := validator.New()
	validate.RegisterValidation("ethaddr", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if len(val) != 40 && len(val) != 42 {
			return false
		}
		if len(val) == 42 {
			if !strings.HasPrefix(val, "0x") && !strings.HasPrefix(val, "0X") {
				return false
			}
			val = val[2:]
		}
		// hexadecimal
		for _, r := range val {
			if (r >= 97 && r <= 102) || // a~f
				(r >= 65 && r <= 70) || // A~F
				(r >= 48 && r <= 57) { // 0~9
				continue
			} else {
				return false
			}
		}
		fl.Field().SetString(eth.MixedcaseAddress(val))
		return true
	})
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
	// load default configs
	if err := config.LoadReputableTokens(ctx, app.MongoDB); err != nil {
		return nil, err
	}
	if err := config.LoadDefaultChainConfig(); err != nil {
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
	// upsert doom-001 key
	if err := app.UpsertDoom001(ctx); err != nil {
		return nil, err
	}
	return app, nil
}

func (app *Application) UpsertDoom001(ctx context.Context) error {
	// check whether connectorKeyID exists
	rpcCtx := rpc.NewContext(ctx, app.Server.LocolizerManager)
	exist, err := connector_grpc.CheckKeyExisting(ctx, rpcCtx, app.ConnectorApiKey, app.ConnectorKeyID)
	if err != nil {
		return err
	}
	// check whether to update, if connectorKeyID exists
	if exist {
		update, err := strconv.ParseBool(strings.TrimSpace(os.Getenv("UPDATE_CONNECTOR_KEY")))
		if err != nil {
			return err
		}
		// get publicKey, if no need to update
		if !update {
			publicKey, err := connector_grpc.GetPublicKey(ctx, rpcCtx, app.ConnectorApiKey, app.ConnectorKeyID)
			if err != nil {
				return err
			}
			app.Logger.Info("connector key",
				zap.String("connectorApiKey", app.ConnectorApiKey),
				zap.String("connectorKeyID", app.ConnectorKeyID),
				zap.String("publicKey", publicKey),
			)
			return nil
		}
		// delete private key, if need to update
		if err := connector_grpc.DeletePrivateKey(ctx, rpcCtx, app.ConnectorApiKey, app.ConnectorKeyID); err != nil {
			return err
		}
	}
	// create one
	publicKey, err := connector_grpc.CreatePrivateKey(ctx, rpcCtx, app.ConnectorApiKey, app.ConnectorKeyID)
	if err != nil {
		return err
	}
	app.Logger.Info("connector key",
		zap.String("connectorApiKey", app.ConnectorApiKey),
		zap.String("connectorKeyID", app.ConnectorKeyID),
		zap.String("publicKey", publicKey),
	)
	return nil
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
	doomServer, err := grpcserver.NewDoomServer(
		ctx,
		app.Env,
		app.Logger,
		app.RedisClient,
		app.Server.LocolizerManager,
		app.ConnectorApiKey,
		app.ConnectorKeyID,
		app.Validator,
		app.MongoDB,
	)
	if err != nil {
		return err
	}
	doom_api.RegisterDoomServiceServer(
		app.Server.GrpcServer(),
		doomServer)
	app.DoomServer = doomServer
	app.Logger.Info("RegisterDoomServiceServer success")
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
