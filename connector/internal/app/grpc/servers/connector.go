package grpcservers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
	connector_api "github.com/nextsurfer/connector/api"
	"github.com/nextsurfer/connector/api/response"
	"github.com/nextsurfer/connector/internal/app/grpc/services"
	"github.com/nextsurfer/connector/internal/pkg/dao"
	"github.com/nextsurfer/connector/internal/pkg/redis"
	"github.com/nextsurfer/connector/internal/pkg/simplejson"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type ConnectorServer struct {
	connector_api.UnimplementedConnectorServiceServer
	env              gutil.APPEnvType
	logger           *zap.Logger
	daoManager       *dao.Manager
	redisOption      *redis.Option
	localizeManager  *localize.Manager
	connectorService *services.ConnectorService
	validator        *validator.Validate
}

func NewConnectorServer(env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option, localizeManager *localize.Manager, validator *validator.Validate) *ConnectorServer {
	connectorService := services.NewConnectorService(env, logger, daoManager, redisOption)
	return &ConnectorServer{
		env:              env,
		logger:           logger,
		daoManager:       daoManager,
		redisOption:      redisOption,
		localizeManager:  localizeManager,
		connectorService: connectorService,
		validator:        validator,
	}
}

func (s *ConnectorServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *ConnectorServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *ConnectorServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

type OracleFieldType struct {
	Duration int64 `json:"duration"` // millisecond
}

func (s *ConnectorServer) oracleField(startTS time.Time) string {
	data, _ := json.Marshal(OracleFieldType{Duration: int64(time.Since(startTS) / time.Millisecond)})
	return string(data)
}

// service methods --------------------------------------------------------------------------------------------

func (s *ConnectorServer) CreatePassword(ctx context.Context, req *connector_api.CreatePasswordRequest) (*connector_api.CreatePasswordResponse, error) {
	startTS := time.Now()
	var resp connector_api.CreatePasswordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreatePassword", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreatePassword", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	// service duration
	data, err := s.connectorService.CreatePassword(ctx, rpcCtx, startTS, req.ApiKey, req.KeyID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *ConnectorServer) CheckPassword(ctx context.Context, req *connector_api.CheckPasswordRequest) (*connector_api.CheckPasswordResponse, error) {
	startTS := time.Now()
	var resp connector_api.CheckPasswordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CheckPassword", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CheckPassword", &resp) }()
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	data, err := s.connectorService.CheckPassword(ctx, rpcCtx, startTS, req.ApiKey, req.KeyID, req.PasswordHash)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *ConnectorServer) DeletePassword(ctx context.Context, req *connector_api.DeletePasswordRequest) (*connector_api.DeletePasswordResponse, error) {
	startTS := time.Now()
	var resp connector_api.DeletePasswordResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeletePassword", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeletePassword", &resp) }()
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	err := s.connectorService.DeletePassword(ctx, rpcCtx, startTS, req.ApiKey, req.KeyID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *ConnectorServer) CreatePrivateKey(ctx context.Context, req *connector_api.CreatePrivateKeyRequest) (*connector_api.CreatePrivateKeyResponse, error) {
	startTS := time.Now()
	var resp connector_api.CreatePrivateKeyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreatePrivateKey", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreatePrivateKey", &resp) }()
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	data, err := s.connectorService.CreatePrivateKey(ctx, rpcCtx, startTS, req.ApiKey, req.KeyID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *ConnectorServer) GetPublicKey(ctx context.Context, req *connector_api.GetPublicKeyRequest) (*connector_api.GetPublicKeyResponse, error) {
	startTS := time.Now()
	var resp connector_api.GetPublicKeyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetPublicKey", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetPublicKey", &resp) }()
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	data, err := s.connectorService.GetPublicKey(ctx, rpcCtx, startTS, req.ApiKey, req.KeyID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *ConnectorServer) DeletePrivateKey(ctx context.Context, req *connector_api.DeletePrivateKeyRequest) (*connector_api.DeletePrivateKeyResponse, error) {
	startTS := time.Now()
	var resp connector_api.DeletePrivateKeyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeletePrivateKey", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeletePrivateKey", &resp) }()
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	err := s.connectorService.DeletePrivateKey(ctx, rpcCtx, startTS, req.ApiKey, req.KeyID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *ConnectorServer) SaveData(ctx context.Context, req *connector_api.SaveDataRequest) (*connector_api.SaveDataResponse, error) {
	startTS := time.Now()
	var resp connector_api.SaveDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "SaveData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "SaveData", &resp) }()
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	err := s.connectorService.SaveData(ctx, rpcCtx, startTS, req.ApiKey,
		req.KeyID, req.DataID, req.ReplaceCurrentItem, req.Data, req.PlaintextHash)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *ConnectorServer) DeleteData(ctx context.Context, req *connector_api.DeleteDataRequest) (*connector_api.DeleteDataResponse, error) {
	startTS := time.Now()
	var resp connector_api.DeleteDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeleteData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeleteData", &resp) }()
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	err := s.connectorService.DeleteData(ctx, rpcCtx, startTS, req.ApiKey, req.KeyID, req.DataID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *ConnectorServer) DecryptData(ctx context.Context, req *connector_api.DecryptDataRequest) (*connector_api.DecryptDataResponse, error) {
	startTS := time.Now()
	var resp connector_api.DecryptDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DecryptData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DecryptData", &resp) }()
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	data, err := s.connectorService.DecryptData(ctx, rpcCtx, startTS, req.ApiKey, req.KeyID, req.Data)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *ConnectorServer) GetData(ctx context.Context, req *connector_api.GetDataRequest) (*connector_api.GetDataResponse, error) {
	startTS := time.Now()
	var resp connector_api.GetDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetData", &resp) }()
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	data, err := s.connectorService.GetData(ctx, rpcCtx, startTS, req.ApiKey, req.KeyID, req.DataID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *ConnectorServer) CheckKeyExisting(ctx context.Context, req *connector_api.CheckKeyExistingRequest) (*connector_api.CheckKeyExistingResponse, error) {
	startTS := time.Now()
	var resp connector_api.CheckKeyExistingResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CheckKeyExisting", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CheckKeyExisting", &resp) }()
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	data, err := s.connectorService.CheckKeyExisting(ctx, rpcCtx, startTS, req.ApiKey, req.KeyID)
	if err != nil {
		resp.Code = err.Code
		resp.Message = err.Message
		resp.DebugMessage = err.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}
