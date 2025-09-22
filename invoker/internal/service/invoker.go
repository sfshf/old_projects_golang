package service

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/invoker/api/response"
	"github.com/nextsurfer/invoker/internal/common/slark"
	"github.com/nextsurfer/invoker/internal/dao"
	"go.uber.org/zap"
)

type InvokerService struct {
	*AdminService
	*SiteService
	*CategoryService
	*PostService
	*CommentService
	*UserService

	AppID           string
	LocalizeManager *localize.Manager
	Logger          *zap.Logger
	RedisClient     *redis.Client
	DaoManager      *dao.Manager
	Validator       *validator.Validate
	UserInfoCache   map[int64]*UserInfo
}

func NewInvokerService(
	ctx context.Context,
	appID string,
	localizeManager *localize.Manager,
	logger *zap.Logger,
	redisClient *redis.Client,
	daoManager *dao.Manager,
	validator *validator.Validate,
) (*InvokerService, error) {
	invokerService := &InvokerService{
		AppID:           appID,
		LocalizeManager: localizeManager,
		Logger:          logger,
		RedisClient:     redisClient,
		DaoManager:      daoManager,
		Validator:       validator,
	}
	// var err error
	// admin service
	invokerService.AdminService = NewAdminService(invokerService)
	// site service
	invokerService.SiteService = NewSiteService(invokerService)
	// category service
	invokerService.CategoryService = NewCategoryService(invokerService)
	// post service
	invokerService.PostService = NewPostService(invokerService)
	// comment service
	invokerService.CommentService = NewCommentService(invokerService)
	// user service
	invokerService.UserService = NewUserService(invokerService)
	// cache
	invokerService.UserInfoCache = make(map[int64]*UserInfo)
	return invokerService, nil
}

func (s *InvokerService) ValidateRequest(ctx context.Context, rpcCtx *rpc.Context, request interface{}) *gerror.AppError {
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

func (s *InvokerService) fetchUserNickname(ctx context.Context, rpcCtx *rpc.Context, userID int64) (string, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	if userID <= 0 {
		return "", nil
	}
	if userInfo := s.UserInfoCache[userID]; userInfo == nil {
		atUser, err := slark.GetUserInfo(ctx, rpcCtx, userID)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return "", gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		s.UserInfoCache[userID] = &UserInfo{
			ID:       atUser.Id,
			Nickname: atUser.Nickname,
		}
		return atUser.Nickname, nil
	} else {
		return userInfo.Nickname, nil
	}
}
