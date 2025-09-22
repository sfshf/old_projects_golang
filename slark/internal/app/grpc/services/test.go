package services

import (
	"context"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	slark_api "github.com/nextsurfer/slark/api"
	"github.com/nextsurfer/slark/api/response"
	"github.com/nextsurfer/slark/internal/pkg/dao"
	"github.com/nextsurfer/slark/internal/pkg/redis"
	util "github.com/nextsurfer/slark/internal/pkg/util"
	"go.uber.org/zap"
)

type TestService struct {
	logger      *zap.Logger
	daoManager  *dao.Manager
	redisOption *redis.Option
}

func NewTestService(logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option) *TestService {
	return &TestService{
		logger:      logger,
		daoManager:  daoManager,
		redisOption: redisOption,
	}
}

func (s *TestService) GetRegistrationEmailCaptchas(ctx context.Context, rpcCtx *rpc.Context) (*slark_api.GetRegistrationEmailCaptchasResponse_Data, *gerror.AppError) {
	emailCaptchas, err := util.GetRegistrationEmailCaptchasInRedis(ctx, rpcCtx, s.redisOption.Client)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var list []*slark_api.GetRegistrationEmailCaptchasResponse_EmailCaptcha
	for _, one := range emailCaptchas {
		list = append(list, &slark_api.GetRegistrationEmailCaptchasResponse_EmailCaptcha{
			Email: one.Email,
			Code:  one.Captcha,
		})
	}
	return &slark_api.GetRegistrationEmailCaptchasResponse_Data{List: list}, nil
}

func (s *TestService) GetLoginEmailCaptchas(ctx context.Context, rpcCtx *rpc.Context) (*slark_api.GetLoginEmailCaptchasResponse_Data, *gerror.AppError) {
	emailCaptchas, err := util.GetLoginEmailCaptchasInRedis(ctx, rpcCtx, s.redisOption.Client)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var list []*slark_api.GetLoginEmailCaptchasResponse_EmailCaptcha
	for _, one := range emailCaptchas {
		list = append(list, &slark_api.GetLoginEmailCaptchasResponse_EmailCaptcha{
			Email: one.Email,
			Code:  one.Captcha,
		})
	}
	return &slark_api.GetLoginEmailCaptchasResponse_Data{List: list}, nil
}
