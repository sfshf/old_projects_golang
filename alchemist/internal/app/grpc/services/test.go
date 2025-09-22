package services

import (
	"context"
	"errors"

	"github.com/nextsurfer/alchemist/api/response"
	"github.com/nextsurfer/alchemist/internal/pkg/dao"
	"github.com/nextsurfer/alchemist/internal/pkg/redis"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	devicecheck "github.com/rinchsan/device-check-go/v2"
	"go.uber.org/zap"
)

type TestService struct {
	app         string
	env         gutil.APPEnvType
	logger      *zap.Logger
	daoManager  *dao.Manager
	redisOption *redis.Option
}

func NewTestService(appID string, env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option) *TestService {
	s := &TestService{
		app:         appID,
		env:         env,
		logger:      logger,
		daoManager:  daoManager,
		redisOption: redisOption,
	}
	return s
}

func (s *TestService) ResetDeviceCheck(ctx context.Context, rpcCtx *rpc.Context, appID, deviceToken0, deviceToken1 string) *gerror.AppError {
	// device check
	var dcEnv devicecheck.Environment
	switch s.env {
	case gutil.AppEnvDEV, gutil.AppEnvPPE:
		dcEnv = devicecheck.Development
	case gutil.AppEnvPROD:
		dcEnv = devicecheck.Production
	}
	cred := devicecheck.NewCredentialString(util.AppConfig(appID).DeviceCheck.PrivKeyPem)
	cfg := devicecheck.NewConfig(util.AppConfig(appID).DeviceCheck.IssuerID, util.AppConfig(appID).DeviceCheck.KeyID, dcEnv)
	dck := devicecheck.New(cred, cfg)
	var result devicecheck.QueryTwoBitsResult
	if err := devicecheck.New(cred, cfg).QueryTwoBits(ctx, deviceToken0, &result); err != nil && !errors.Is(err, devicecheck.ErrBitStateNotFound) {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeBadRequest)
	}
	if result.Bit1 {
		if err := dck.UpdateTwoBits(ctx, deviceToken1, result.Bit0, false); err != nil && !errors.Is(err, devicecheck.ErrBitStateNotFound) {
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeBadRequest)
		}
	}
	return nil
}

func (s *TestService) RemoveFraudRecord(ctx context.Context, rpcCtx *rpc.Context, userID int64) *gerror.AppError {
	if err := s.daoManager.UserRegisteredOnOldDeviceDAO.DeleteByUserID(ctx, userID); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}
