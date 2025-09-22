package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/pswds_backend/api/response"
	"github.com/nextsurfer/pswds_backend/internal/common/slark"
	"github.com/nextsurfer/pswds_backend/internal/dao"
	slark_api "github.com/nextsurfer/slark/api"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type PswdsService struct {
	*BackupService
	*PasswordService
	*NonPasswordService
	*ShareService
	*CronService
	*TrustedContactService
	*SecurityQuestionsService
	*PrivacyEmailService
	*FamilyService

	AppID           string
	LocalizeManager *localize.Manager
	Logger          *zap.Logger
	RedisClient     *redis.Client
	DaoManager      *dao.Manager
	Validator       *validator.Validate
}

const (
	// 根据用户的userID缓存用户基础信息，加速用户信息获取+检测；
	RedisKeyPrefixSlarkInfo = "PSWDS::SlarkInfo::"
	// 根据用户的登录session缓存用户基础数据，加速用户身份检测；
	RedisKeyPrefixSlarkSession = "PSWDS::SlarkSession::"
)

type SlarkInfo struct {
	UserID   int64  `json:"userID,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

func NewPswdsService(
	ctx context.Context,
	appID string,
	localizeManager *localize.Manager,
	logger *zap.Logger,
	redisClient *redis.Client,
	daoManager *dao.Manager,
	validator *validator.Validate,
) (*PswdsService, error) {
	pswdsService := &PswdsService{
		AppID:           appID,
		LocalizeManager: localizeManager,
		Logger:          logger,
		RedisClient:     redisClient,
		DaoManager:      daoManager,
		Validator:       validator,
	}
	var err error
	// backup service
	pswdsService.BackupService = NewBackupService(ctx, pswdsService)
	// password service
	pswdsService.PasswordService = NewPasswordService(ctx, pswdsService)
	// non password service
	pswdsService.NonPasswordService = NewNonPasswordService(ctx, pswdsService)
	// share service
	pswdsService.ShareService = NewShareService(ctx, pswdsService)
	// cron service
	pswdsService.CronService, err = NewCronService(ctx, pswdsService)
	if err != nil {
		return nil, err
	}
	// security questions service
	pswdsService.SecurityQuestionsService, err = NewSecurityQuestionsService(ctx, pswdsService)
	if err != nil {
		return nil, err
	}
	// trusted contact service
	pswdsService.TrustedContactService, err = NewTrustedContactService(ctx, pswdsService)
	if err != nil {
		return nil, err
	}
	// privacy email service
	pswdsService.PrivacyEmailService, err = NewPrivacyEmailService(ctx, pswdsService)
	if err != nil {
		return nil, err
	}
	// family service
	pswdsService.FamilyService = NewFamilyService(ctx, pswdsService)
	return pswdsService, nil
}

func (s *PswdsService) ValidateRequest(ctx context.Context, rpcCtx *rpc.Context, request interface{}) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	if err := s.Validator.Struct(request); err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	return nil
}

func (s *PswdsService) ValidateLoginInfo(ctx context.Context, rpcCtx *rpc.Context) (*slark_api.LoginResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	var err error
	if rpcCtx.SessionID == "" {
		err = errors.New("empty session")
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidSession")).WithCode(response.StatusCodeUnauthorized)
	}
	var result SlarkInfo
	data, err := s.RedisClient.Get(ctx, RedisKeyPrefixSlarkSession+rpcCtx.SessionID).Result()
	if err != nil {
		if err != redis.Nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			loginInfo, err := slark.LoginInfo(ctx, rpcCtx)
			if err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			if loginInfo == nil {
				err = errors.New("invalid session")
				logger.Error("bad request", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidSession")).WithCode(response.StatusCodeUnauthorized)
			}
			result = SlarkInfo{
				UserID:   loginInfo.UserID,
				Email:    loginInfo.Email,
				Nickname: loginInfo.Nickname,
				Phone:    loginInfo.Phone,
			}
			data, err := json.Marshal(result)
			if err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			if err := s.RedisClient.Set(ctx, RedisKeyPrefixSlarkSession+rpcCtx.SessionID, string(data), 20*time.Minute).Err(); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	} else {
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return &slark_api.LoginResponse_Data{
		UserID:   result.UserID,
		Email:    result.Email,
		Nickname: result.Nickname,
		Phone:    result.Phone,
	}, nil
}

func (s *PswdsService) FetchSlarkInfo(ctx context.Context, rpcCtx *rpc.Context, userID int64) (*SlarkInfo, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	var err error
	var result SlarkInfo
	key := fmt.Sprintf("%s%d", RedisKeyPrefixSlarkInfo, userID)
	data, err := s.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			slarkInfo, err := slark.GetUserInfo(ctx, rpcCtx, userID)
			if err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			if slarkInfo == nil {
				err = errors.New("invalid user id")
				logger.Error("bad request", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
			}
			result = SlarkInfo{
				UserID:   slarkInfo.Id,
				Email:    slarkInfo.Email,
				Nickname: slarkInfo.Nickname,
			}
			data, err := json.Marshal(result)
			if err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			if err := s.RedisClient.Set(ctx, key, string(data), 20*time.Minute).Err(); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	} else {
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return &result, nil
}

func (s *PswdsService) FetchSlarkInfoByEmail(ctx context.Context, rpcCtx *rpc.Context, email string) (*SlarkInfo, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	var err error
	var cursor uint64
	var keys []string
	for {
		keys, cursor, err = s.RedisClient.Scan(ctx, cursor, RedisKeyPrefixSlarkInfo+"*", 10).Result()
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		for _, key := range keys {
			data, err := s.RedisClient.Get(ctx, key).Result()
			if err != nil {
				if err != redis.Nil {
					logger.Error("internal error", zap.NamedError("appError", err))
					return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
			} else {
				var result SlarkInfo
				if err := json.Unmarshal([]byte(data), &result); err != nil {
					logger.Error("internal error", zap.NamedError("appError", err))
					return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
				if result.Email == email {
					return &result, nil
				}
			}

		}
		if cursor == 0 {
			break
		}
	}
	return nil, nil
}
