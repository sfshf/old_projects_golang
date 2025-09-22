package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	console_api "github.com/nextsurfer/oracle/api/console"
	"github.com/nextsurfer/oracle/internal/console/service"
	"github.com/nextsurfer/oracle/internal/dao"
	"go.uber.org/zap"
)

type ConsoleServer struct {
	console_api.UnimplementedConsoleServiceServer

	Name            string
	Host            string
	Env             gutil.APPEnvType
	Logger          *zap.Logger
	DaoManager      *dao.Manager
	RedisClient     *redis.Client
	LocalizeManager *localize.Manager
	RpcPort         int
	RpcServer       *rpc.Server
	HttpPort        int
	HttpServer      *http.Server
	WebPath         string
	ConsoleService  *service.ConsoleService
	Validator       *validator.Validate
}

func NewConsoleServer(ctx context.Context, name string, env gutil.APPEnvType, Logger *zap.Logger, daoManager *dao.Manager, redisClient *redis.Client, host string, grpcPort, httpPort int, tomlPath string, validator *validator.Validate, webPath string) (*ConsoleServer, error) {
	s := &ConsoleServer{
		Name:        name,
		Host:        host,
		RpcPort:     grpcPort,
		HttpPort:    httpPort,
		Env:         env,
		Logger:      Logger,
		DaoManager:  daoManager,
		RedisClient: redisClient,
		WebPath:     webPath,
		Validator:   validator,
	}
	// 1. grpc
	// register server
	if err := s.registerRpcServer(); err != nil {
		return nil, err
	}
	// localize manager load message files
	s.LocalizeManager = s.RpcServer.LocolizerManager
	if err := s.loadMessageFiles(tomlPath); err != nil {
		return nil, err
	}
	// console service
	consoleService, err := service.NewConsoleService(ctx, env, Logger, daoManager, redisClient, name, s.LocalizeManager)
	if err != nil {
		return nil, err
	}
	s.ConsoleService = consoleService
	// clear registered gateway node records
	if err := s.clearRegisteredGatewayNodes(); err != nil {
		return nil, err
	}
	// 2. http
	// register http routes
	s.registerRoutes()
	// check prerequisite proto files in db
	if err := s.checkPrerequisiteProtos(); err != nil {
		return nil, err
	}
	// check alarm emails
	if err := s.checkAlarmEmails(); err != nil {
		return nil, err
	}
	// check GATEWAY_API_HOSTNAME host
	if err := s.checkHostnames(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *ConsoleServer) Run(ctx context.Context) error {
	// 1. http server
	go func() {
		s.Logger.Info(fmt.Sprintf("Console http server is running at %s ...", s.HttpServer.Addr))
		if err := s.HttpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				s.Logger.Fatal("Console http server error", zap.NamedError("appError", err))
			} else {
				s.Logger.Info("Console http server closed")
			}
		}
	}()

	// 2. grpc server
	go func() {
		s.Logger.Info(fmt.Sprintf("Console app grpc server is running at %s:%d ...", s.Host, s.RpcPort))
		if err := s.RpcServer.Start(); err != nil {
			s.Logger.Fatal("Console app grpc server error", zap.NamedError("appError", err))
		}
	}()
	return nil
}

func (s *ConsoleServer) Stop(ctx context.Context) error {
	s.Logger.Info("Console servers are stopping ...")
	err := s.RpcServer.Stop()
	if err := s.HttpServer.Shutdown(ctx); err != nil {
		return err
	}
	return err
}
