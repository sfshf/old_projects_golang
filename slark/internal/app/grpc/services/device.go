package services

import (
	"context"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/ground/pkg/util"
	slark_api "github.com/nextsurfer/slark/api"
	"github.com/nextsurfer/slark/api/response"
	"github.com/nextsurfer/slark/internal/pkg/dao"
	"github.com/nextsurfer/slark/internal/pkg/model"
	"github.com/nextsurfer/slark/internal/pkg/redis"
	"go.uber.org/zap"
)

// DeviceService : service is pure business
type DeviceService struct {
	env         util.APPEnvType
	logger      *zap.Logger
	daoManager  *dao.Manager
	redisOption *redis.Option
}

// NewDeviceService is factory function
func NewDeviceService(env util.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option) *DeviceService {
	return &DeviceService{
		env:         env,
		logger:      logger,
		daoManager:  daoManager,
		redisOption: redisOption,
	}
}

// UploadDeviceInfo upload device info
func (s *DeviceService) UploadDeviceInfo(ctx context.Context, rpcCtx *rpc.Context, in *slark_api.DeviceRequest) *gerror.AppError {
	// check if device info has been stored
	stored, err := s.daoManager.DeviceDAO.GetFromDeviceID(ctx, in.DeviceID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if stored != nil {
		// 一般来说不会冲突， 后续优化。
		if stored.DeviceModel != in.DeviceModel || stored.ResolutionWidth != in.ResolutionWidth {
			rpcCtx.Logger.Error("Collisions occurs on two device !!!! fk ", zap.Int64("previousID", stored.ID))
		}
		return nil
	}
	// store device info in mysql
	device := &model.SlkDevice{
		DeviceID:         in.DeviceID,
		Platform:         in.Platform,
		DeviceModel:      in.DeviceModel,
		ResolutionWidth:  in.ResolutionWidth,
		ResolutionHeight: in.ResolutionHeight,
		ScreenDensity:    in.ScreenDensity,
		Rom:              in.Rom,
		RAM:              in.Ram,
	}
	if err := s.daoManager.DeviceDAO.Create(ctx, device); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}
