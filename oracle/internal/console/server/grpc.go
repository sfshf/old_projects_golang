package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	console_api "github.com/nextsurfer/oracle/api/console"
	"github.com/nextsurfer/oracle/api/response"
	"github.com/nextsurfer/oracle/internal/common/simplejson"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *ConsoleServer) registerRpcServer() error {
	tracer := rpc.NewTracer(s.Name, s.Env)
	server, err := rpc.NewServer(s.Name, s.Env, s.Host, s.RpcPort, tracer)
	if err != nil {
		return err
	}
	grpcServer := server.GrpcServer()
	console_api.RegisterConsoleServiceServer(grpcServer, s)
	reflection.Register(grpcServer)
	s.RpcServer = server
	return nil
}

func (s *ConsoleServer) loadMessageFiles(tomlPath string) error {
	files, err := ioutil.ReadDir(tomlPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".toml") {
			fullName := fmt.Sprintf("%s/%s", strings.TrimRight(tomlPath, "/"), f.Name())
			s.LocalizeManager.Bundle.MustLoadMessageFile(fullName)
		}
	}
	return nil
}

func (s *ConsoleServer) clearRegisteredGatewayNodes() error {
	return s.DaoManager.GatewayNodeDAO.RemoveAll(context.Background())
}

func (s *ConsoleServer) isEnvTest() bool {
	return s.Env == gutil.AppEnvDEV || s.Env == gutil.AppEnvPPE
}

func (s *ConsoleServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *ConsoleServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

func (s *ConsoleServer) RegisterGatewayNode(ctx context.Context, req *console_api.RegisterGatewayNodeRequest) (*console_api.RegisterGatewayNodeResponse, error) {
	var resp console_api.RegisterGatewayNodeResponse
	rpcCtx := rpc.NewContext(ctx, s.LocalizeManager)
	s.logRequestData(rpcCtx, "RegisterGatewayNode", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RegisterGatewayNode", &resp) }()
	// validate request basically
	if err := s.Validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.ConsoleService.RegisterGatewayNode(ctx, rpcCtx, req.Name, req.Ipv4, req.RpcPort)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}

func (s *ConsoleServer) UpsertService(ctx context.Context, req *console_api.UpsertServiceRequest) (*console_api.UpsertServiceResponse, error) {
	var resp console_api.UpsertServiceResponse
	rpcCtx := rpc.NewContext(ctx, s.LocalizeManager)
	s.logRequestData(rpcCtx, "UpsertService", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UpsertService", &resp) }()
	// validate request basically
	if err := s.Validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.ConsoleService.UpsertService(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	return &resp, nil
}
