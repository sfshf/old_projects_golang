package internal

import (
	"context"
	"net/url"

	"github.com/go-playground/validator/v10"
	"github.com/nextsurfer/ground/pkg/log"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"github.com/nextsurfer/monitor/internal/common/redis"
	"github.com/nextsurfer/monitor/internal/server"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

// Application app
type Application struct {
	Name          string
	Env           gutil.APPEnvType
	Logger        *zap.Logger
	RedisOption   *redis.Option
	MongoDB       *mongo.Database
	Validator     *validator.Validate
	monitorServer *server.MonitorServer
}

// NewApplication create application
func NewApplication(ctx context.Context, name string, appEnv int, redisDNS, mongodbUri string) (*Application, error) {
	var err error
	app := &Application{
		Name: name,
	}
	// env
	app.Env = gutil.EnvForInt(appEnv)
	// logger
	logOptions := log.NewOptions(app.Name, app.Env, true)
	logger := log.New(logOptions)
	app.Logger = logger
	// validator
	app.Validator = validator.New()
	// redis
	redisOption, err := redis.NewOption(redisDNS, app.Logger)
	if err != nil {
		return nil, err
	}
	app.RedisOption = redisOption
	// mongo db
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
		dbName = "tester"
	}
	app.MongoDB = mgoCli.Database(dbName) // internal servers
	monitorServer, err := server.NewMonitorServer(ctx, app.Env, app.Logger, app.RedisOption, app.MongoDB, app.Validator)
	if err != nil {
		return nil, err
	}
	app.monitorServer = monitorServer
	return app, err
}
