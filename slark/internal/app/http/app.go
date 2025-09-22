package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	gdao "github.com/nextsurfer/ground/pkg/dao"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/log"
	gutil "github.com/nextsurfer/ground/pkg/util"
	httphandlers "github.com/nextsurfer/slark/internal/app/http/servers"
	"github.com/nextsurfer/slark/internal/pkg/dao"
	"github.com/nextsurfer/slark/internal/pkg/redis"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

type Application struct {
	Env              gutil.APPEnvType
	Name             string
	Host             string
	Port             int // http port
	WebPath          string
	Logger           *zap.Logger
	RedisOption      *redis.Option
	DaoManager       *dao.Manager
	Validator        *validator.Validate
	Server           *http.Server
	LocolizerManager *localize.Manager
}

func NewApplication(name string, port int, host string, appEnv int, redisDNS, mysqlDNS, webPath, tomlPath string) (*Application, error) {
	app := &Application{
		Name:    name,
		Host:    host,
		Port:    port,
		WebPath: webPath,
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
	// dao
	app.DaoManager = dao.NewManager(gdao.NewOption(mysqlDNS, app.Name, app.Port+2, app.Env))
	// localizer
	app.LocolizerManager = localize.NewManager()
	gerror.AddErrorLocalizerMessage(app.LocolizerManager.Bundle)
	// http server
	app.Server = &http.Server{Addr: fmt.Sprintf("%s:%v", app.Host, app.Port)}
	// register handlers
	app.RegisterHandlers()
	// load i18n files
	if err := app.LoadMessageFiles(tomlPath); err != nil {
		return nil, err
	}
	return app, nil
}

func (app *Application) RegisterHandlers() {
	// web ui static server
	http.Handle("/", http.FileServer(http.Dir(app.WebPath)))
	// api handles
	http.HandleFunc("/api/discourse/sso/v1", httphandlers.DiscourseConnect)
	http.HandleFunc("/api/session/validate/v1", httphandlers.ValidateSessionID(app.RedisOption, app.DaoManager))
	http.HandleFunc("/api/signIn/v1", httphandlers.SignIn(app.RedisOption, app.DaoManager))
	http.HandleFunc("/api/qrcode/token/v1", httphandlers.RequestQRLoginToken(app.RedisOption))
	http.HandleFunc("/api/qrcode/login/check/v1", httphandlers.CheckQRLogin(app.RedisOption, app.DaoManager))
	// cors
	handler := cors.Default().Handler(http.DefaultServeMux)
	app.Server.Handler = handler
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
			app.LocolizerManager.Bundle.MustLoadMessageFile(fullName)
		}
	}
	return nil
}

func (app *Application) Start(ctx context.Context) error {
	app.Logger.Info(fmt.Sprintf("HTTP application is starting at %s:%d ...\n", app.Host, app.Port))
	go func() {
		if err := app.Server.ListenAndServe(); err != nil {
			app.Logger.Error("Http application start", zap.NamedError("appError", err))
		}
	}()
	return nil
}

func (app *Application) Stop(ctx context.Context) error {
	return app.Server.Shutdown(ctx)
}
