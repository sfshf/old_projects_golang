package servers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	slark_grpc "github.com/nextsurfer/slark/pkg/grpc"
	word_api "github.com/nextsurfer/word/api"
	"github.com/nextsurfer/word/api/response"
	"github.com/nextsurfer/word/internal/app/grpc/services"
	"github.com/nextsurfer/word/internal/pkg/dao"
	"github.com/nextsurfer/word/internal/pkg/redis"
	"github.com/nextsurfer/word/internal/pkg/simplejson"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// WordServer , there are some logic about header/session/bff in server layer.
type WordServer struct {
	word_api.UnimplementedWordServiceServer
	env             gutil.APPEnvType
	logger          *zap.Logger
	daoManager      *dao.Manager
	redisOption     *redis.Option
	localizeManager *localize.Manager
	wordService     *services.WordService
	audioService    *services.TextToAudioService
	validator       *validator.Validate
}

// NewWordServer is factory function
func NewWordServer(env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option, localizeManager *localize.Manager, validator *validator.Validate) *WordServer {
	wordService := services.NewWordService(logger, daoManager, redisOption)
	audioService := services.NewTextToAudioService(logger)
	return &WordServer{
		env:             env,
		logger:          logger,
		daoManager:      daoManager,
		redisOption:     redisOption,
		localizeManager: localizeManager,
		wordService:     wordService,
		audioService:    audioService,
		validator:       validator,
	}
}

func (s *WordServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *WordServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *WordServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

// service methods --------------------------------------------------------------------------------------------

func (s *WordServer) FetchAudioURL(ctx context.Context, req *word_api.TextToAudioRequest) (*word_api.TextToAudioResponse, error) {
	var resp word_api.TextToAudioResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "FetchAudioURL", req)
	defer func() { s.deferLogResponseData(rpcCtx, "FetchAudioURL", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	if req.ApiKey != os.Getenv("API_KEY") {
		err := errors.New("parameter error, please update app version")
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	if req.Ssml == "<speak></speak>" {
		// check empty and support for some test cases.
		err := errors.New("GetAudio parameter wrong. ssml is empty")
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()
	data, appError := s.audioService.GetAudio(ctx, rpcCtx, req.ApiKey, req.Text, req.Ssml, req.Accent, req.Voice)
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

func (s *WordServer) FavoriteDefinition(ctx context.Context, req *word_api.FavoriteDefinitionRequest) (*word_api.FavoriteDefinitionResponse, error) {
	var resp word_api.FavoriteDefinitionResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "FavoriteDefinition", req)
	defer func() { s.deferLogResponseData(rpcCtx, "FavoriteDefinition", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != 0 {
			err = fmt.Errorf("session error: %v", rpcCtx)
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	appError := s.wordService.FavoriteDefinition(ctx, rpcCtx, req.DefinitionID, loginInfo.Data.UserID)
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

func (s *WordServer) FavoritedDefinitions(ctx context.Context, req *word_api.Empty) (*word_api.FavoritedDefinitionsResponse, error) {
	var resp word_api.FavoritedDefinitionsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "FavoritedDefinitions", req)
	defer func() { s.deferLogResponseData(rpcCtx, "FavoriteDefinition", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != 0 {
			err = fmt.Errorf("session error: %v", rpcCtx)
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	data, appError := s.wordService.FavoritedDefinitions(ctx, rpcCtx, loginInfo.Data.UserID)
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

func (s *WordServer) ProgressBackupStatus(ctx context.Context, req *word_api.ProgressBackupStatusRequest) (*word_api.ProgressBackupStatusResponse, error) {
	var resp word_api.ProgressBackupStatusResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ProgressBackupStatus", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ProgressBackupStatus", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != 0 {
			err = fmt.Errorf("session error: %v", rpcCtx)
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	data, appError := s.wordService.ProgressBackupStatus(ctx, rpcCtx, req.Timestamp, req.Version, loginInfo.Data.UserID)
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

func (s *WordServer) UploadProgressBackup(ctx context.Context, req *word_api.UploadProgressBackupRequest) (*word_api.UploadProgressBackupResponse, error) {
	var resp word_api.UploadProgressBackupResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UploadProgressBackup", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UploadProgressBackup", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != 0 {
			err = fmt.Errorf("session error: %v", rpcCtx)
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	appError := s.wordService.UploadProgressBackup(ctx, rpcCtx, req.Timestamp, req.Version, req.Data, loginInfo.Data.UserID)
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

func (s *WordServer) DownloadProgressBackup(ctx context.Context, req *word_api.DownloadProgressBackupRequest) (*word_api.DownloadProgressBackupResponse, error) {
	var resp word_api.DownloadProgressBackupResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DownloadProgressBackup", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DownloadProgressBackup", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != 0 {
			err = fmt.Errorf("session error: %v", rpcCtx)
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	data, appError := s.wordService.DownloadProgressBackup(ctx, rpcCtx, req.Timestamp, req.Version, loginInfo.Data.UserID)
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
