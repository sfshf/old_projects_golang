package services

import (
	"context"
	"fmt"
	"time"

	alchemist_api "github.com/nextsurfer/alchemist/api"
	"github.com/nextsurfer/alchemist/api/response"
	"github.com/nextsurfer/alchemist/internal/pkg/dao"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/internal/pkg/redis"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AlchemistConsoleService struct {
	app                         string
	env                         gutil.APPEnvType
	logger                      *zap.Logger
	daoManager                  *dao.Manager
	redisOption                 *redis.Option
	cron                        *cron.Cron
	subscriptionCountEntryID    cron.EntryID
	subscriptionCountCronStatus struct {
		Started           bool
		StartedOrStopedAt time.Time
		ScheduleSpec      string
		LastExecError     string
	}
}

func NewAlchemistConsoleService(appID string, env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option) (*AlchemistConsoleService, error) {
	s := &AlchemistConsoleService{
		app:         appID,
		env:         env,
		logger:      logger,
		daoManager:  daoManager,
		redisOption: redisOption,
	}
	cron := cron.New()
	s.cron = cron
	// subscription count cron entry
	subscriptionCountEntryID, err := cron.AddFunc("@daily", func() {
		if err := doSubscriptionCount(s.daoManager, logger); err != nil {
			s.subscriptionCountCronStatus.LastExecError = err.Error()
		}
	})
	if err != nil {
		return nil, err
	}
	s.subscriptionCountEntryID = subscriptionCountEntryID
	// start cron jobs
	s.cron.Start()
	s.subscriptionCountCronStatus.Started = true
	s.subscriptionCountCronStatus.StartedOrStopedAt = time.Now()
	s.subscriptionCountCronStatus.ScheduleSpec = "@daily"
	return s, nil
}

func (s *AlchemistConsoleService) ListConfigs(ctx context.Context, rpcCtx *rpc.Context, password string) (*alchemist_api.ListConfigsResponse_Data, *gerror.AppError) {
	if err := util.ValidateApiKey(ctx, rpcCtx, s.app, password, util.RoleRead); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyIsInvalid")).WithCode(response.StatusCodeUnauthorized)
	}
	allAppConfigs, err := s.daoManager.AppConfigDAO.GetAll(ctx)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	var list []*alchemist_api.ListConfigsResponse_AppConfig
	for _, appConfig := range allAppConfigs {
		list = append(list, &alchemist_api.ListConfigsResponse_AppConfig{
			Id:     appConfig.ID,
			Config: appConfig.Config,
		})
	}

	return &alchemist_api.ListConfigsResponse_Data{List: list}, nil
}

func (s *AlchemistConsoleService) CreateConfig(ctx context.Context, rpcCtx *rpc.Context, password, config string) *gerror.AppError {
	if err := util.ValidateApiKey(ctx, rpcCtx, s.app, password, util.RoleWrite); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyIsInvalid")).WithCode(response.StatusCodeUnauthorized)
	}
	appID, config, err := util.CheckConfigFormat(config)
	if err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_DeformedConfig")).WithCode(response.StatusCodeBadRequest)
	}
	if err := s.daoManager.AppConfigDAO.Create(ctx, &AppConfig{
		App:    appID,
		Config: config,
	}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// refresh app config
	if err := util.RefreshConfig(s.daoManager); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *AlchemistConsoleService) UpdateConfig(ctx context.Context, rpcCtx *rpc.Context, password string, id int64, config string) *gerror.AppError {
	if err := util.ValidateApiKey(ctx, rpcCtx, s.app, password, util.RoleWrite); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyIsInvalid")).WithCode(response.StatusCodeUnauthorized)
	}
	appID, config, err := util.CheckConfigFormat(config)
	if err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_DeformedConfig")).WithCode(response.StatusCodeBadRequest)
	}
	appConfig, err := s.daoManager.AppConfigDAO.GetByID(ctx, id)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongParameters")).WithCode(response.StatusCodeInternalServerError)
	}
	if appConfig == nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeWrongParameters)
	}
	if appID != appConfig.App {
		err = fmt.Errorf("appID [%s] in config not equal to appConfig's appID [%s] of id [%d]", appID, appConfig.App, id)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	if err := s.daoManager.AppConfigDAO.UpdateByID(ctx, id, config); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// refresh app config
	if err := util.RefreshConfig(s.daoManager); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *AlchemistConsoleService) DeleteConfig(ctx context.Context, rpcCtx *rpc.Context, password string, id int64) *gerror.AppError {
	if err := util.ValidateApiKey(ctx, rpcCtx, s.app, password, util.RoleWrite); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyIsInvalid")).WithCode(response.StatusCodeUnauthorized)
	}
	if err := s.daoManager.AppConfigDAO.DeleteByID(ctx, id); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// refresh app config
	if err := util.RefreshConfig(s.daoManager); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func doSubscriptionCount(daoManager *dao.Manager, logger *zap.Logger) error {
	ctx := context.Background()
	loc := time.FixedZone("UTC-5", -5*60*60)
	now := time.Now().In(loc).UnixMilli()
	counts, err := daoManager.SubscriptionStateProdDAO.CountSubscription(ctx)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return err
	}
	var subscriptionCounts []SubscriptionCount
	for _, one := range counts {
		subscriptionCounts = append(subscriptionCounts, SubscriptionCount{
			App:   one.App,
			Time:  now,
			Count: one.Count,
		})
	}
	if len(subscriptionCounts) > 0 {
		if err := daoManager.SubscriptionCountDAO.Create(ctx, subscriptionCounts); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return err
		}
	}
	return nil
}

func (s *AlchemistConsoleService) GetCurrentSubscriptionCount(ctx context.Context, rpcCtx *rpc.Context, password, appID string) (*alchemist_api.GetCurrentSubscriptionCountResponse_Data, *gerror.AppError) {
	if err := util.ValidateApiKey(ctx, rpcCtx, s.app, password, util.RoleRead); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyIsInvalid")).WithCode(response.StatusCodeUnauthorized)
	}
	cnt, err := s.daoManager.SubscriptionStateProdDAO.CountSubscriptionByApp(ctx, appID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &alchemist_api.GetCurrentSubscriptionCountResponse_Data{Count: cnt}, nil
}

func (s *AlchemistConsoleService) ListSubscriptionCounts(ctx context.Context, rpcCtx *rpc.Context, password, appID, startDate, endDate string) (*alchemist_api.ListSubscriptionCountsResponse_Data, *gerror.AppError) {
	if err := util.ValidateApiKey(ctx, rpcCtx, s.app, password, util.RoleRead); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyIsInvalid")).WithCode(response.StatusCodeUnauthorized)
	}
	loc := time.FixedZone("UTC-5", -5*60*60)
	startTS, err := time.ParseInLocation("2006-01-02", startDate, loc)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	endTS, err := time.ParseInLocation("2006-01-02", endDate, loc)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	list, err := s.daoManager.SubscriptionCountDAO.GetList(ctx, appID, startTS.UnixMilli(), endTS.UnixMilli())
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var res []int64
	for _, ele := range list {
		res = append(res, ele.Count)
	}
	return &alchemist_api.ListSubscriptionCountsResponse_Data{List: res}, nil
}

func (s *AlchemistConsoleService) GetAllApps(ctx context.Context, rpcCtx *rpc.Context, password string) (*alchemist_api.GetAllAppsResponse_Data, *gerror.AppError) {
	if err := util.ValidateApiKey(ctx, rpcCtx, s.app, password, util.RoleRead); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyIsInvalid")).WithCode(response.StatusCodeUnauthorized)
	}
	allAppConfigs, err := s.daoManager.AppConfigDAO.GetAll(ctx)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	var list []string
	for _, appConfig := range allAppConfigs {
		list = append(list, appConfig.App)
	}
	return &alchemist_api.GetAllAppsResponse_Data{List: list}, nil
}
