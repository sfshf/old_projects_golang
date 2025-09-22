package servers

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	alchemist_api "github.com/nextsurfer/alchemist/api"
	"github.com/nextsurfer/alchemist/api/response"
	"github.com/nextsurfer/alchemist/internal/app/grpc/services"
	"github.com/nextsurfer/alchemist/internal/pkg/dao"
	"github.com/nextsurfer/alchemist/internal/pkg/redis"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	slark_response "github.com/nextsurfer/slark/api/response"
	slark_grpc "github.com/nextsurfer/slark/pkg/grpc"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type AlchemistServer struct {
	alchemist_api.UnimplementedAlchemistServiceServer
	app              string
	env              gutil.APPEnvType
	logger           *zap.Logger
	daoManager       *dao.Manager
	redisOption      *redis.Option
	localizeManager  *localize.Manager
	alchemistService *services.AlchemistService
	validator        *validator.Validate
}

func NewAlchemistServer(app string, env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option, localizeManager *localize.Manager, validator *validator.Validate) (*AlchemistServer, error) {
	alchemistService, err := services.NewAlchemistService(app, env, logger, daoManager, redisOption)
	if err != nil {
		return nil, err
	}
	return &AlchemistServer{
		app:              app,
		env:              env,
		logger:           logger,
		daoManager:       daoManager,
		redisOption:      redisOption,
		localizeManager:  localizeManager,
		alchemistService: alchemistService,
		validator:        validator,
	}, nil
}

func (s *AlchemistServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *AlchemistServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := util.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *AlchemistServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := util.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

// service methods --------------------------------------------------------------------------------------------

func (s *AlchemistServer) GetAppAccountToken(ctx context.Context, req *alchemist_api.Empty) (*alchemist_api.GetAppAccountTokenResponse, error) {
	var resp alchemist_api.GetAppAccountTokenResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetAppAccountToken", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetAppAccountToken", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	data, appError := s.alchemistService.GetAppAccountToken(ctx, rpcCtx, loginInfo.Data.UserID)
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

func (s *AlchemistServer) JoinReferralProgram(ctx context.Context, req *alchemist_api.JoinReferralProgramRequest) (*alchemist_api.JoinReferralProgramResponse, error) {
	var resp alchemist_api.JoinReferralProgramResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "JoinReferralProgram", req)
	defer func() { s.deferLogResponseData(rpcCtx, "JoinReferralProgram", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.JoinReferralProgram(ctx, rpcCtx, req.AppID, loginInfo.Data.UserID)
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

func (s *AlchemistServer) CheckReferralCode(ctx context.Context, req *alchemist_api.CheckReferralCodeRequest) (*alchemist_api.CheckReferralCodeResponse, error) {
	var resp alchemist_api.CheckReferralCodeResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CheckReferralCode", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CheckReferralCode", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.CheckReferralCode(ctx, rpcCtx, req.ReferralCode)
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

func (s *AlchemistServer) BindReferralCode(ctx context.Context, req *alchemist_api.BindReferralCodeRequest) (*alchemist_api.EmptyResponse, error) {
	var resp alchemist_api.EmptyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "BindReferralCode", req)
	defer func() { s.deferLogResponseData(rpcCtx, "BindReferralCode", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.alchemistService.BindReferralCode(ctx, rpcCtx, req.AppID, req.ReferralCode, loginInfo.Data.UserID)
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

func (s *AlchemistServer) GetRewardPoints(ctx context.Context, req *alchemist_api.GetRewardPointsRequest) (*alchemist_api.GetRewardPointsResponse, error) {
	var resp alchemist_api.GetRewardPointsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetRewardPoints", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetRewardPoints", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.GetRewardPoints(ctx, rpcCtx, req.AppID, loginInfo.Data.UserID)
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

func (s *AlchemistServer) GetNewUserDiscountState(ctx context.Context, req *alchemist_api.GetNewUserDiscountStateRequest) (*alchemist_api.GetNewUserDiscountStateResponse, error) {
	var resp alchemist_api.GetNewUserDiscountStateResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetNewUserDiscountState", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetNewUserDiscountState", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.GetNewUserDiscountState(ctx, rpcCtx, req.AppID, req.BilledTimes, loginInfo.Data.UserID)
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

func (s *AlchemistServer) GetNewUserDiscountSignature(ctx context.Context, req *alchemist_api.GetNewUserDiscountSignatureRequest) (*alchemist_api.GetNewUserDiscountSignatureResponse, error) {
	var resp alchemist_api.GetNewUserDiscountSignatureResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetNewUserDiscountSignature", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetNewUserDiscountSignature", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.GetNewUserDiscountSignature(ctx, rpcCtx, req.AppID, req.Environment, loginInfo.Data.UserID)
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

func (s *AlchemistServer) UseNewUserDiscountOffer(ctx context.Context, req *alchemist_api.UseNewUserDiscountOfferRequest) (*alchemist_api.UseNewUserDiscountOfferResponse, error) {
	var resp alchemist_api.UseNewUserDiscountOfferResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "UseNewUserDiscountOffer", req)
	defer func() { s.deferLogResponseData(rpcCtx, "UseNewUserDiscountOffer", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.alchemistService.UseNewUserDiscountOffer(ctx, rpcCtx, req.AppID, req.OfferID, req.Environment, loginInfo.Data.UserID)
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

func (s *AlchemistServer) GetTrialState(ctx context.Context, req *alchemist_api.GetTrialStateRequest) (*alchemist_api.GetTrialStateResponse, error) {
	var resp alchemist_api.GetTrialStateResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetTrialState", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetTrialState", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.GetTrialState(ctx, rpcCtx, req.AppID, loginInfo.Data.UserID)
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

func (s *AlchemistServer) GetRewardList(ctx context.Context, req *alchemist_api.GetRewardListRequest) (*alchemist_api.GetRewardListResponse, error) {
	var resp alchemist_api.GetRewardListResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetRewardList", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetRewardList", &resp) }()
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.GetRewardList(ctx, rpcCtx, req.AppID)
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

func (s *AlchemistServer) RedeemReward(ctx context.Context, req *alchemist_api.RedeemRewardRequest) (*alchemist_api.RedeemRewardResponse, error) {
	var resp alchemist_api.RedeemRewardResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RedeemReward", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RedeemReward", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.RedeemReward(ctx, rpcCtx, req.AppID, req.RewardID, loginInfo.Data.UserID)
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

func (s *AlchemistServer) RedeemRewardOffer(ctx context.Context, req *alchemist_api.RedeemRewardOfferRequest) (*alchemist_api.RedeemRewardOfferResponse, error) {
	var resp alchemist_api.RedeemRewardOfferResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RedeemRewardOffer", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RedeemRewardOffer", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.RedeemRewardOffer(ctx, rpcCtx, req.AppID, req.RewardID, req.Environment, loginInfo.Data.UserID)
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

func (s *AlchemistServer) FinishRewardOffer(ctx context.Context, req *alchemist_api.FinishRewardOfferRequest) (*alchemist_api.FinishRewardOfferResponse, error) {
	var resp alchemist_api.FinishRewardOfferResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "FinishRewardOffer", req)
	defer func() { s.deferLogResponseData(rpcCtx, "FinishRewardOffer", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.FinishRewardOffer(ctx, rpcCtx, req.AppID, req.RewardID, req.Environment, loginInfo.Data.UserID)
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

func (s *AlchemistServer) ConvertUnusedNewUserDiscount(ctx context.Context, req *alchemist_api.ConvertUnusedNewUserDiscountRequest) (*alchemist_api.ConvertUnusedNewUserDiscountResponse, error) {
	var resp alchemist_api.ConvertUnusedNewUserDiscountResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ConvertUnusedNewUserDiscount", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ConvertUnusedNewUserDiscount", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.alchemistService.ConvertUnusedNewUserDiscount(ctx, rpcCtx, req.AppID, loginInfo.Data.UserID)
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

func (s *AlchemistServer) CheckReferralCodeAndDevice(ctx context.Context, req *alchemist_api.CheckReferralCodeAndDeviceRequest) (*alchemist_api.CheckReferralCodeAndDeviceResponse, error) {
	var resp alchemist_api.CheckReferralCodeAndDeviceResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CheckReferralCodeAndDevice", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CheckReferralCodeAndDevice", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.CheckReferralCodeAndDevice(ctx, rpcCtx, req.AppID, req.DeviceToken, req.ReferralCode, loginInfo.Data.UserID)
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

func (s *AlchemistServer) MarkNewRegistration(ctx context.Context, req *alchemist_api.MarkNewRegistrationRequest) (*alchemist_api.MarkNewRegistrationResponse, error) {
	var resp alchemist_api.MarkNewRegistrationResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "MarkNewRegistration", req)
	defer func() { s.deferLogResponseData(rpcCtx, "MarkNewRegistration", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	appError := s.alchemistService.MarkNewRegistration(ctx, rpcCtx, req.AppID, req.DeviceToken0, req.DeviceToken1, loginInfo.Data.UserID)
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

func (s *AlchemistServer) Unregister(ctx context.Context, req *alchemist_api.Empty) (*alchemist_api.UnregisterResponse, error) {
	var resp alchemist_api.UnregisterResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "Unregister", req)
	defer func() { s.deferLogResponseData(rpcCtx, "Unregister", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	appError := s.alchemistService.Unregister(ctx, rpcCtx, loginInfo.Data.UserID)
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

func (s *AlchemistServer) GetState(ctx context.Context, req *alchemist_api.GetStateRequest) (*alchemist_api.GetStateResponse, error) {
	var resp alchemist_api.GetStateResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetState", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetState", &resp) }()
	// get slark login info
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeInternalServerError
		resp.Message = rpcCtx.Localizer.Localize("FatalErrMsg")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	} else {
		if loginInfo.Code != slark_response.StatusCodeOK {
			err = fmt.Errorf("session error: %v", rpcCtx)
			rpcCtx.Logger.Error("session error", zap.NamedError("appError", err))
			resp.Code = loginInfo.Code
			resp.Message = loginInfo.Message
			resp.DebugMessage = structpb.NewStringValue(err.Error())
			return &resp, nil
		}
	}
	// validate request basically
	if err := s.validator.Struct(req); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		resp.Code = response.StatusCodeWrongParameters
		resp.Message = rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")
		resp.DebugMessage = structpb.NewStringValue(err.Error())
		return &resp, nil
	}
	data, appError := s.alchemistService.GetState(ctx, rpcCtx, req.AppID, req.BilledTimes, loginInfo.Data.UserID)
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
