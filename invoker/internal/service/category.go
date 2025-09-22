package service

import (
	"context"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	invoker_api "github.com/nextsurfer/invoker/api"
	"github.com/nextsurfer/invoker/api/response"
	. "github.com/nextsurfer/invoker/internal/model"
	"go.uber.org/zap"
)

type CategoryService struct {
	*InvokerService
}

func NewCategoryService(InvokerService *InvokerService) *CategoryService {
	return &CategoryService{
		InvokerService: InvokerService,
	}
}

func (s *CategoryService) GetCategoryList(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.GetCategoryListRequest) (*invoker_api.GetCategoryListResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var list []*invoker_api.GetCategoryListResponse_CategoryInfo
	categories, err := s.DaoManager.CategoryDAO.GetBySiteID(ctx, req.SiteID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for _, category := range categories {
		list = append(list, &invoker_api.GetCategoryListResponse_CategoryInfo{
			Id:     category.ID,
			SiteID: category.SiteID,
			Name:   category.Name,
			Posts:  category.Posts,
		})
	}
	return &invoker_api.GetCategoryListResponse_Data{
		List: list,
	}, nil
}

func (s *CategoryService) GetCategory(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.GetCategoryRequest) (*invoker_api.GetCategoryResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	// validate site
	site, err := s.DaoManager.SiteDAO.GetByID(ctx, req.SiteID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if site == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// get
	category, err := s.DaoManager.CategoryDAO.GetBySiteIDAndName(ctx, req.SiteID, req.Name)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if category == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	return &invoker_api.GetCategoryResponse_Data{
		Id:     category.ID,
		SiteID: category.SiteID,
		Name:   category.Name,
	}, nil
}

func (s *CategoryService) AddCategory(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.AddCategoryRequest, userID int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// validate site and site admin
	if appError := s.ValidateSiteAdmin(ctx, rpcCtx, req.SiteID, userID); appError != nil {
		return appError
	}
	// add
	category, err := s.DaoManager.CategoryDAO.GetBySiteIDAndName(ctx, req.SiteID, req.Name)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if category == nil {
		if err := s.DaoManager.CategoryDAO.Create(ctx, &Category{
			SiteID: req.SiteID,
			Name:   req.Name,
		}); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return nil
}

func (s *CategoryService) EditCategory(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.EditCategoryRequest, userID int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	category, err := s.DaoManager.CategoryDAO.GetByID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if category == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// validate site admin
	if appError := s.ValidateSiteAdmin(ctx, rpcCtx, category.SiteID, userID); appError != nil {
		return appError
	}
	// update
	category.Name = req.Name
	if err := s.DaoManager.CategoryDAO.UpdateByID(
		ctx,
		req.Id,
		category,
	); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.DeleteCategoryRequest, userID int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	category, err := s.DaoManager.CategoryDAO.GetByID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if category == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// validate site admin
	if appError := s.ValidateSiteAdmin(ctx, rpcCtx, category.SiteID, userID); appError != nil {
		return appError
	}
	// delete
	if err := s.DaoManager.CategoryDAO.DeleteByID(ctx, req.Id); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}
