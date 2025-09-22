package services

import (
	"context"

	"github.com/nextsurfer/alchemist/api/response"
	"github.com/nextsurfer/alchemist/internal/pkg/dao"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/internal/pkg/redis"
	"github.com/nextsurfer/alchemist/pkg/consts"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/ground/pkg/util"
	"go.uber.org/zap"
)

type AlchemistNotificationService struct {
	logger      *zap.Logger
	daoManager  *dao.Manager
	redisOption *redis.Option
}

func NewAlchemistNotificationService(env util.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option) *AlchemistNotificationService {
	s := &AlchemistNotificationService{
		logger:      logger,
		daoManager:  daoManager,
		redisOption: redisOption,
	}
	return s
}

func (s *AlchemistNotificationService) AppStoreNotificationsProd(ctx context.Context, rpcCtx *rpc.Context, signedPayload string) *gerror.AppError {
	// insert into raw_transactions table
	if err := s.daoManager.RawTransactionsDAO.Create(ctx, &RawTransaction{
		Data:        `{"signedPayload":"` + signedPayload + `"}`,
		Environment: consts.DESIGNATED_ENVIRONMENT_NUM_PROD,
	}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *AlchemistNotificationService) AppStoreNotificationsTest(ctx context.Context, rpcCtx *rpc.Context, signedPayload string) *gerror.AppError {
	// insert into raw_transactions table
	if err := s.daoManager.RawTransactionsDAO.Create(ctx, &RawTransaction{
		Data:        `{"signedPayload":"` + signedPayload + `"}`,
		Environment: consts.DESIGNATED_ENVIRONMENT_NUM_SANDBOX,
	}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}
