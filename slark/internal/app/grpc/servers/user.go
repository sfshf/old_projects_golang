package servers

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	slark_api "github.com/nextsurfer/slark/api"
	"github.com/nextsurfer/slark/api/response"
	"github.com/nextsurfer/slark/internal/app/grpc/services"
	"github.com/nextsurfer/slark/internal/pkg/dao"
	"github.com/nextsurfer/slark/internal/pkg/redis"
	"github.com/nextsurfer/slark/internal/pkg/simplejson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// UserServer , there are some logic about header/session/bff in server layer.
type UserServer struct {
	slark_api.UnimplementedUserServiceServer
	env             gutil.APPEnvType
	logger          *zap.Logger
	daoManager      *dao.Manager
	redisOption     *redis.Option
	localizeManager *localize.Manager
	userService     *services.UserService
	validator       *validator.Validate
	MongoDB         *mongo.Database
}

// NewUserServer is factory function
func NewUserServer(env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option, localizeManager *localize.Manager, validator *validator.Validate, mongoDB *mongo.Database) *UserServer {
	userService := services.NewUserService(env, logger, daoManager, redisOption, mongoDB)
	return &UserServer{
		env:             env,
		logger:          logger,
		daoManager:      daoManager,
		redisOption:     redisOption,
		localizeManager: localizeManager,
		userService:     userService,
		validator:       validator,
		MongoDB:         mongoDB,
	}
}

func (s *UserServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *UserServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *UserServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

// service methods --------------------------------------------------------------------------------------------

func (s *UserServer) LoginByPhone(ctx context.Context, req *slark_api.PhoneLoginRequest) (*slark_api.LoginResponse, error) {
	var resp slark_api.LoginResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "LoginByPhone", req)
	defer func() { s.deferLogResponseData(rpcCtx, "LoginByPhone", &resp) }()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.userService.LoginByPhone(ctx, rpcCtx, req.Phone, req.PasswordHash)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) LoginByEmail(ctx context.Context, req *slark_api.EmailLoginRequest) (*slark_api.LoginResponse, error) {
	var resp slark_api.LoginResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "LoginByEmail", req)
	defer func() { s.deferLogResponseData(rpcCtx, "LoginByEmail", &resp) }()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	data, appError := s.userService.LoginByEmail(ctx, rpcCtx, req.Email, req.PasswordHash)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) LoginBySession(ctx context.Context, req *slark_api.Empty) (*slark_api.LoginResponse, error) {
	var resp slark_api.LoginResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "LoginBySession", req)
	defer func() { s.deferLogResponseData(rpcCtx, "LoginBySession", &resp) }()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	data, appError := s.userService.LoginBySession(ctx, rpcCtx)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) SendRegistrationEmailCaptcha(ctx context.Context, req *slark_api.SendRegistrationEmailCaptchaRequest) (*slark_api.SendRegistrationEmailCaptchaResponse, error) {
	var resp slark_api.SendRegistrationEmailCaptchaResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "SendRegistrationEmailCaptcha", req)
	defer func() { s.deferLogResponseData(rpcCtx, "SendRegistrationEmailCaptcha", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	_, appError := s.userService.SendRegistrationEmailCaptcha(ctx, rpcCtx, req.Email)
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

func (s *UserServer) RegisterByEmail(ctx context.Context, req *slark_api.RegisterByEmailRequest) (*slark_api.LoginResponse, error) {
	var resp slark_api.LoginResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RegisterByEmail", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RegisterByEmail", &resp) }()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.userService.RegisterByEmail(ctx, rpcCtx, req.Email, req.Nickname, req.PasswordHash, req.Captcha)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) RegisterByEmailCaptcha(ctx context.Context, req *slark_api.RegisterByEmailCaptchaRequest) (*slark_api.LoginResponse, error) {
	var resp slark_api.LoginResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RegisterByEmailCaptcha", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RegisterByEmailCaptcha", &resp) }()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.userService.RegisterByEmailCaptcha(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) LoginByApple(ctx context.Context, req *slark_api.LoginByAppleRequest) (*slark_api.LoginResponse, error) {
	var resp slark_api.LoginResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "LoginByApple", req)
	defer func() { s.deferLogResponseData(rpcCtx, "LoginByApple", &resp) }()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.userService.LoginByApple(ctx, rpcCtx, req.Email, req.UserIdentifier)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) LogOutBySession(ctx context.Context, req *slark_api.Empty) (*slark_api.EmptyResponse, error) {
	var resp slark_api.EmptyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "LogOutBySession", req)
	defer func() { s.deferLogResponseData(rpcCtx, "LogOutBySession", &resp) }()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	appError := s.userService.LogOutBySession(ctx, rpcCtx)
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

func (s *UserServer) CheckLogin(ctx context.Context, req *slark_api.Empty) (*slark_api.EmptyResponse, error) {
	var resp slark_api.EmptyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CheckLogin", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CheckLogin", &resp) }()
	appError := s.userService.CheckLogin(ctx, rpcCtx)
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

func (s *UserServer) LoginInfo(ctx context.Context, req *slark_api.Empty) (*slark_api.LoginResponse, error) {
	var resp slark_api.LoginResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "LoginInfo", req)
	defer func() { s.deferLogResponseData(rpcCtx, "LoginInfo", &resp) }()
	data, appError := s.userService.LoginInfo(ctx, rpcCtx)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) CheckRegistration(ctx context.Context, req *slark_api.CheckRegistrationRequest) (*slark_api.CheckRegistrationResponse, error) {
	var resp slark_api.CheckRegistrationResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CheckRegistration", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CheckRegistration", &resp) }()
	data, appError := s.userService.CheckRegistration(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) ValidateUserIDs(ctx context.Context, req *slark_api.ValidateUserIDsRequest) (*slark_api.ValidateUserIDsResponse, error) {
	var resp slark_api.ValidateUserIDsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ValidateUserIDs", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ValidateUserIDs", &resp) }()
	data, appError := s.userService.ValidateUserIDs(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) GetUserInfo(ctx context.Context, req *slark_api.GetUserInfoRequest) (*slark_api.GetUserInfoResponse, error) {
	var resp slark_api.GetUserInfoResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetUserInfo", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetUserInfo", &resp) }()
	data, appError := s.userService.GetUserInfo(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) QRLogin(ctx context.Context, req *slark_api.QRLoginRequest) (*slark_api.QRLoginResponse, error) {
	var resp slark_api.QRLoginResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "QRLogin", req)
	defer func() { s.deferLogResponseData(rpcCtx, "QRLogin", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.userService.QRLogin(ctx, rpcCtx, req.Token)
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

func (s *UserServer) SendLoginEmailCode(ctx context.Context, req *slark_api.SendLoginEmailCodeRequest) (*slark_api.SendLoginEmailCodeResponse, error) {
	var resp slark_api.SendLoginEmailCodeResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "SendLoginEmailCode", req)
	defer func() { s.deferLogResponseData(rpcCtx, "SendLoginEmailCode", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	_, appError := s.userService.SendLoginEmailCode(ctx, rpcCtx, req.Email)
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

func (s *UserServer) LoginByEmailCode(ctx context.Context, req *slark_api.LoginByEmailCodeRequest) (*slark_api.LoginResponse, error) {
	var resp slark_api.LoginResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "LoginByEmailCode", req)
	defer func() { s.deferLogResponseData(rpcCtx, "LoginByEmailCode", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.userService.LoginByEmailCode(ctx, rpcCtx, req.Email, req.Code)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) UpdateNickname(ctx context.Context, req *slark_api.UpdateNicknameRequest) (*slark_api.UpdateNicknameResponse, error) {
	var resp slark_api.UpdateNicknameResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UpdateNickname", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UpdateNickname", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.userService.UpdateNickname(ctx, rpcCtx, req.Nickname)
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

func (s *UserServer) Unregister(ctx context.Context, req *slark_api.Empty) (*slark_api.UnregisterResponse, error) {
	var resp slark_api.UnregisterResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "Unregister", req)
	defer func() { s.deferLogResponseData(rpcCtx, "Unregister", &resp) }()
	appError := s.userService.Unregister(ctx, rpcCtx)
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

func (s *UserServer) RandomNickname(ctx context.Context, req *slark_api.Empty) (*slark_api.RandomNicknameResponse, error) {
	var resp slark_api.RandomNicknameResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RandomNickname", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RandomNickname", &resp) }()
	data, appError := s.userService.RandomNickname(ctx, rpcCtx)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}

func (s *UserServer) CreateSecondaryPassword(ctx context.Context, req *slark_api.CreateSecondaryPasswordRequest) (*slark_api.EmptyResponse, error) {
	var resp slark_api.EmptyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreateSecondaryPassword", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreateSecondaryPassword", &resp) }()
	appError := s.userService.CreateSecondaryPassword(ctx, rpcCtx, req)
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

func (s *UserServer) UpdateSecondaryPassword(ctx context.Context, req *slark_api.UpdateSecondaryPasswordRequest) (*slark_api.EmptyResponse, error) {
	var resp slark_api.EmptyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UpdateSecondaryPassword", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UpdateSecondaryPassword", &resp) }()
	appError := s.userService.UpdateSecondaryPassword(ctx, rpcCtx, req)
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

func (s *UserServer) LoginBySecondaryPassword(ctx context.Context, req *slark_api.LoginBySecondaryPasswordRequest) (*slark_api.LoginResponse, error) {
	var resp slark_api.LoginResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "LoginBySecondaryPassword", req)
	defer func() { s.deferLogResponseData(rpcCtx, "LoginBySecondaryPassword", &resp) }()
	data, appError := s.userService.LoginBySecondaryPassword(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	return &resp, nil
}
