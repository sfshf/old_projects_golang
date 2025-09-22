package service

import (
	"context"
	"time"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	invoker_api "github.com/nextsurfer/invoker/api"
	"github.com/nextsurfer/invoker/api/response"
	"github.com/nextsurfer/invoker/internal/common/slark"
	"github.com/nextsurfer/invoker/internal/dao"
	. "github.com/nextsurfer/invoker/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PostService struct {
	*InvokerService
}

func NewPostService(InvokerService *InvokerService) *PostService {
	return &PostService{
		InvokerService: InvokerService,
	}
}

func (s *PostService) GetPostList(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.GetPostListRequest) (*invoker_api.GetPostListResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	tx := s.DaoManager.PostDAO.Table(ctx)
	tx = tx.Where(`state=?`, dao.PostState_Posted)
	if req.SiteID > 0 {
		tx = tx.Where(`site_id=?`, req.SiteID)
	}
	if req.CategoryID > 0 {
		tx = tx.Where(`category_id=?`, req.CategoryID)
	}
	if req.SortedByActivity {
		tx = tx.Order(`activity DESC`)
	} else if req.SortedByViews {
		tx = tx.Order(`views DESC`)
	} else if req.SortedByReplies {
		tx = tx.Order(`replies DESC`)
	}
	// total
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var posts []*Post
	if err := tx.Select("id", "title", "posted_at", "posted_by", "views", "replies", "thumbups", "activity").
		Offset(int(req.PageNumber * req.PageSize)).Limit(int(req.PageSize)).Find(&posts).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// list
	var list []*invoker_api.GetPostListResponse_PostInfo
	for _, post := range posts {
		postedByString, appError := s.fetchUserNickname(ctx, rpcCtx, post.PostedBy)
		if appError != nil {
			return nil, appError
		}
		list = append(list, &invoker_api.GetPostListResponse_PostInfo{
			Id:             post.ID,
			Title:          post.Title,
			PostedAt:       post.PostedAt,
			PostedBy:       post.PostedBy,
			PostedByString: postedByString,
			Views:          post.Views,
			Replies:        post.Replies,
			Activity:       post.Activity,
		})
	}
	return &invoker_api.GetPostListResponse_Data{List: list, Total: total}, nil
}

func (s *PostService) GetPostDetail(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.GetPostDetailRequest) (*invoker_api.GetPostDetailResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	// fetch login info, if has
	loginInfo, _ := s.aggregatedLoginInfo(ctx, rpcCtx, true)
	post, err := s.DaoManager.PostDAO.GetByID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if post == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	postedByString, appError := s.fetchUserNickname(ctx, rpcCtx, post.PostedBy)
	if appError != nil {
		return nil, appError
	}
	// hasThumbup
	var hasThumbup bool
	if loginInfo != nil {
		hasThumbup, err = s.DaoManager.ThumbupDAO.HasThumbupPost(ctx, loginInfo.UserID, post.ID)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return &invoker_api.GetPostDetailResponse_Data{
		Id:             post.ID,
		Title:          post.Title,
		PostedAt:       post.PostedAt,
		PostedBy:       post.PostedBy,
		PostedByString: postedByString,
		Content:        post.Content,
		Image:          post.Image,
		State:          post.State,
		Views:          post.Views,
		Replies:        post.Replies,
		Activity:       post.Activity,
		Thumbups:       post.Thumbups,
		Thumbup:        hasThumbup,
	}, nil
}

func (s *PostService) AddPost(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.AddPostRequest, userID int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// validate site
	site, err := s.DaoManager.SiteDAO.GetByID(ctx, req.SiteID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if site == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// validate category
	category, err := s.DaoManager.CategoryDAO.GetByID(ctx, req.CategoryID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if category == nil || category.SiteID != req.SiteID {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// add
	now := time.Now().UnixMilli()
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if err := daoManager.PostDAO.Create(ctx, &Post{
			SiteID:     category.SiteID,
			CategoryID: category.ID,
			Title:      req.Title,
			PostedBy:   userID,
			PostedAt:   now,
			Content:    req.Content,
			Image:      req.Image,
			State:      1,
			Activity:   now,
		}); err != nil {
			return err
		}
		// update category posts field
		return daoManager.CategoryDAO.UpdateByID(ctx, category.ID, &Category{
			Posts: category.Posts + 1,
		})
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *PostService) EditPost(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.EditPostRequest, userID int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// validate post
	post, err := s.DaoManager.PostDAO.GetByID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if post == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	if post.PostedBy != userID {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidUserID")).WithCode(response.StatusCodeForbidden)
	}
	// edit
	if err := s.DaoManager.PostDAO.UpdateByID(ctx, post.ID, &Post{
		Title:    req.Title,
		Content:  req.Content,
		Image:    req.Image,
		Activity: time.Now().UnixMilli(),
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *PostService) DeletePost(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.DeletePostRequest, userID int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// validate post
	post, err := s.DaoManager.PostDAO.GetByID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if post == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// validate category
	category, err := s.DaoManager.CategoryDAO.GetByID(ctx, post.CategoryID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if category == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	if post.PostedBy != userID {
		// validate site and site admin
		if appError := s.ValidateSiteAdmin(ctx, rpcCtx, category.SiteID, userID); appError != nil {
			logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidUserID")).WithCode(response.StatusCodeForbidden)
		}
	}
	// delete
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if err := daoManager.PostDAO.DeleteByID(ctx, req.Id); err != nil {
			return err
		}
		return daoManager.CategoryDAO.UpdateByID(ctx, category.ID, &Category{Posts: category.Posts - 1})
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *PostService) aggregatedLoginInfo(ctx context.Context, rpcCtx *rpc.Context, isTry bool) (*invoker_api.AggregatedLoginInfo, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	loginInfo, err := slark.LoginInfo(ctx, rpcCtx)
	if err != nil {
		if isTry {
			return nil, nil
		}
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if loginInfo == nil {
		if isTry {
			return nil, nil
		}
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidSession")).WithCode(response.StatusCodeUnauthorized)
	}
	return &invoker_api.AggregatedLoginInfo{
		UserID:   loginInfo.UserID,
		Nickname: loginInfo.Nickname,
		Email:    loginInfo.Email,
		Phone:    loginInfo.Phone,
	}, nil
}

func (s *PostService) aggregatedSiteInfo(ctx context.Context, rpcCtx *rpc.Context, siteName string) (*invoker_api.AggregatedSiteInfo, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	site, err := s.DaoManager.SiteDAO.GetByName(ctx, siteName)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if site == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	admins, err := s.DaoManager.SiteAdminDAO.GetAdminsBySiteID(ctx, site.ID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &invoker_api.AggregatedSiteInfo{
		Id:     site.ID,
		Name:   site.Name,
		Admins: admins,
	}, nil
}

func (s *PostService) aggregatedCategoryInfo(ctx context.Context, rpcCtx *rpc.Context, siteID int64) ([]*invoker_api.AggregatedCategoryInfo, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	var categorieInfos []*invoker_api.AggregatedCategoryInfo
	categories, err := s.DaoManager.CategoryDAO.GetBySiteID(ctx, siteID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var categoryIDs []int64
	for _, category := range categories {
		categoryIDs = append(categoryIDs, category.ID)
		categorieInfos = append(categorieInfos, &invoker_api.AggregatedCategoryInfo{
			Id:     category.ID,
			SiteID: category.SiteID,
			Name:   category.Name,
			Posts:  category.Posts,
		})
	}
	return categorieInfos, nil
}

func (s *PostService) handlePostList(ctx context.Context, rpcCtx *rpc.Context, posts []*Post) ([]*invoker_api.AggregatedPostListInfo, *gerror.AppError) {
	var postList []*invoker_api.AggregatedPostListInfo
	for _, post := range posts {
		var postedByString string
		postedByString, appError := s.fetchUserNickname(ctx, rpcCtx, post.PostedBy)
		if appError != nil {
			return nil, appError
		}
		postList = append(postList, &invoker_api.AggregatedPostListInfo{
			Id:             post.ID,
			Title:          post.Title,
			PostedAt:       post.PostedAt,
			PostedBy:       post.PostedBy,
			PostedByString: postedByString,
			Views:          post.Views,
			Replies:        post.Replies,
			Activity:       post.Activity,
			Thumbups:       post.Thumbups,
		})
	}
	return postList, nil
}

func (s *PostService) AggregatedSitePage(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.AggregatedSitePageRequest) (*invoker_api.AggregatedSitePageResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var res invoker_api.AggregatedSitePageResponse_Data
	var appError *gerror.AppError
	// fetch login info, if has
	res.LoginInfo, _ = s.aggregatedLoginInfo(ctx, rpcCtx, true)
	// fetch site info
	res.SiteInfo, appError = s.aggregatedSiteInfo(ctx, rpcCtx, req.Site)
	if appError != nil {
		return nil, appError
	}
	// fetch categories
	res.Categories, appError = s.aggregatedCategoryInfo(ctx, rpcCtx, res.SiteInfo.Id)
	if appError != nil {
		return nil, appError
	}
	var categoryIDs []int64
	for _, category := range res.Categories {
		categoryIDs = append(categoryIDs, category.Id)
	}
	// fetch posts
	tx := s.DaoManager.PostDAO.Table(ctx)
	tx = tx.Where(`category_id IN (?) AND state=?`, categoryIDs, dao.PostState_Posted)
	// total
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	res.PostTotal = total
	var posts []*Post
	if err := tx.Order(`activity DESC`).
		Limit(10).
		Find(&posts).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	res.Posts, appError = s.handlePostList(ctx, rpcCtx, posts)
	if appError != nil {
		return nil, appError
	}
	return &res, nil
}

func (s *PostService) AggregatedCategoryPage(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.AggregatedCategoryPageRequest) (*invoker_api.AggregatedCategoryPageResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var res invoker_api.AggregatedCategoryPageResponse_Data
	var appError *gerror.AppError
	// fetch login info, if has
	res.LoginInfo, _ = s.aggregatedLoginInfo(ctx, rpcCtx, true)
	// fetch site info
	res.SiteInfo, appError = s.aggregatedSiteInfo(ctx, rpcCtx, req.Site)
	if appError != nil {
		return nil, appError
	}
	// fetch categories
	res.Categories, appError = s.aggregatedCategoryInfo(ctx, rpcCtx, res.SiteInfo.Id)
	if appError != nil {
		return nil, appError
	}
	var reqCategory *invoker_api.AggregatedCategoryInfo
	for _, category := range res.Categories {
		if req.Category == category.Name {
			reqCategory = category
			break
		}
	}
	// fetch posts
	tx := s.DaoManager.PostDAO.Table(ctx)
	tx = tx.Where(`category_id=? AND state=?`, reqCategory.Id, dao.PostState_Posted)
	// total
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	res.PostTotal = total
	var posts []*Post
	if err := tx.
		Order(`activity DESC`).
		Limit(10).
		Find(&posts).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	res.Posts, appError = s.handlePostList(ctx, rpcCtx, posts)
	if appError != nil {
		return nil, appError
	}
	return &res, nil
}

func (s *PostService) handlePostRelatedData(ctx context.Context, rpcCtx *rpc.Context, post *Post, postInfo *invoker_api.AggregatedPostPageResponse_PostInfo, res *invoker_api.AggregatedPostPageResponse_Data, anchorCommentID int64) *gerror.AppError {
	// category
	for _, category := range res.Categories {
		if post.CategoryID == category.Id {
			res.CategoryInfo = category
			break
		}
	}
	if appError := s.handleAnchorInfo(ctx, rpcCtx, post, postInfo, res, anchorCommentID); appError != nil {
		return appError
	}
	return nil
}

func (s *PostService) handleAnchorInfo(ctx context.Context, rpcCtx *rpc.Context, post *Post, postInfo *invoker_api.AggregatedPostPageResponse_PostInfo, res *invoker_api.AggregatedPostPageResponse_Data, anchorCommentID int64) *gerror.AppError {
	var anchorInfo *invoker_api.AggregatedPostPageResponse_AnchorInfo
	var firstLevelPageNumber int64
	// handle anchorCommentID
	if anchorCommentID > 0 {
		anchorComment, err := s.DaoManager.CommentDAO.GetByID(ctx, anchorCommentID)
		if err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if anchorComment == nil {
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
		}
		anchorInfo = &invoker_api.AggregatedPostPageResponse_AnchorInfo{
			RootCommentID: anchorComment.RootCommentID, // if rootCommentID > 0, should anchor at its second level comments list
		}
		// check first level or second level
		if anchorComment.RootCommentID == 0 {
			// it is first level, get firstLevelPageNumber
			countIndex, err := s.DaoManager.CommentDAO.GetFirstLevelCountAndIndex(ctx, post.ID, anchorComment.ID)
			if err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			if countIndex == nil {
				rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
			}
			firstLevelPageNumber = countIndex.RowNumber / 20
			anchorInfo.FirstLevelPageNumber = firstLevelPageNumber
		} else {
			// get firstLevelPageNumber
			firstCountIndex, err := s.DaoManager.CommentDAO.GetFirstLevelCountAndIndex(ctx, post.ID, anchorComment.RootCommentID)
			if err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			if firstCountIndex == nil {
				rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
			}
			firstLevelPageNumber = firstCountIndex.RowNumber / 20
			anchorInfo.FirstLevelPageNumber = firstLevelPageNumber
			// it is second level, get secondLevelPageNumber and secondLevelTotal
			secondCountIndex, err := s.DaoManager.CommentDAO.GetSecondLevelCountAndIndex(ctx, anchorComment.RootCommentID, anchorComment.ID)
			if err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			if secondCountIndex == nil {
				rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
			}
			anchorInfo.SecondLevelPageNumber = secondCountIndex.RowNumber / 10
			if appError := s.handleSecondLevelComments(ctx, rpcCtx, anchorInfo); appError != nil {
				return appError
			}
		}
	}
	res.AnchorInfo = anchorInfo
	if appError := s.handlePostComments(ctx, rpcCtx, post, postInfo, res, firstLevelPageNumber); appError != nil {
		return appError
	}
	return nil
}

func (s *PostService) handleSecondLevelComments(ctx context.Context, rpcCtx *rpc.Context, anchorInfo *invoker_api.AggregatedPostPageResponse_AnchorInfo) *gerror.AppError {
	// comments
	data, appError := s.CommentService.GetComments(ctx, rpcCtx, &invoker_api.GetCommentsRequest{
		CommentID:  anchorInfo.RootCommentID,
		PageSize:   10,
		PageNumber: anchorInfo.SecondLevelPageNumber,
	})
	if appError != nil {
		return appError
	}
	anchorInfo.SecondLevelTotal = data.Total
	var comments []*invoker_api.AggregatedPostPageResponse_CommentInfo
	if len(data.List) > 0 {
		for _, item := range data.List {
			comments = append(comments, &invoker_api.AggregatedPostPageResponse_CommentInfo{
				Id:             item.Id,
				PostID:         item.PostID,
				RootCommentID:  item.RootCommentID,
				PostedAt:       item.PostedAt,
				PostedBy:       item.PostedBy,
				PostedByString: item.PostedByString,
				AtWho:          item.AtWho,
				AtWhoString:    item.AtWhoString,
				Content:        item.Content,
				Replies:        item.Replies,
				Thumbups:       item.Thumbups,
				Thumbup:        item.Thumbup,
			})
		}
	}
	anchorInfo.SecondLevelComments = comments
	return nil
}

func (s *PostService) handlePostComments(ctx context.Context, rpcCtx *rpc.Context, post *Post, postInfo *invoker_api.AggregatedPostPageResponse_PostInfo, res *invoker_api.AggregatedPostPageResponse_Data, firstLevelPageNumber int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// comments
	data, appError := s.CommentService.GetComments(ctx, rpcCtx, &invoker_api.GetCommentsRequest{
		PostID:     post.ID,
		PageSize:   20,
		PageNumber: firstLevelPageNumber,
	})
	if appError != nil {
		return appError
	}
	postInfo.CommentTotal = data.Total
	var comments []*invoker_api.AggregatedPostPageResponse_CommentInfo
	if len(data.List) > 0 {
		for _, item := range data.List {
			comments = append(comments, &invoker_api.AggregatedPostPageResponse_CommentInfo{
				Id:             item.Id,
				PostID:         item.PostID,
				RootCommentID:  item.RootCommentID,
				PostedAt:       item.PostedAt,
				PostedBy:       item.PostedBy,
				PostedByString: item.PostedByString,
				AtWho:          item.AtWho,
				AtWhoString:    item.AtWhoString,
				Content:        item.Content,
				Replies:        item.Replies,
				Thumbups:       item.Thumbups,
				Thumbup:        item.Thumbup,
			})
		}
	}
	postInfo.Comments = comments
	// update views
	if err := s.DaoManager.PostDAO.UpdateByID(ctx, post.ID, &Post{
		Views: post.Views + 1,
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	postInfo.Views = post.Views + 1
	res.Post = postInfo
	return nil
}

func (s *PostService) AggregatedPostPage(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.AggregatedPostPageRequest) (*invoker_api.AggregatedPostPageResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var res invoker_api.AggregatedPostPageResponse_Data
	var appError *gerror.AppError
	// fetch login info, if has
	res.LoginInfo, _ = s.aggregatedLoginInfo(ctx, rpcCtx, true)
	// fetch site info
	res.SiteInfo, appError = s.aggregatedSiteInfo(ctx, rpcCtx, req.Site)
	if appError != nil {
		return nil, appError
	}
	// fetch categories
	res.Categories, appError = s.aggregatedCategoryInfo(ctx, rpcCtx, res.SiteInfo.Id)
	if appError != nil {
		return nil, appError
	}
	// fetch post
	post, err := s.DaoManager.PostDAO.GetByID(ctx, req.PostID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if post == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// get slark user info
	postedByString, appError := s.fetchUserNickname(ctx, rpcCtx, post.PostedBy)
	if appError != nil {
		return nil, appError
	}
	// hasThumbup
	var hasThumbup bool
	if res.LoginInfo != nil {
		hasThumbup, err = s.DaoManager.ThumbupDAO.HasThumbupPost(ctx, res.LoginInfo.UserID, post.ID)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	postInfo := &invoker_api.AggregatedPostPageResponse_PostInfo{
		Id:             post.ID,
		Title:          post.Title,
		PostedAt:       post.PostedAt,
		PostedBy:       post.PostedBy,
		PostedByString: postedByString,
		Content:        post.Content,
		Image:          post.Image,
		State:          post.State,
		Replies:        post.Replies,
		Activity:       post.Activity,
		Thumbups:       post.Thumbups,
		Thumbup:        hasThumbup,
	}
	if appError = s.handlePostRelatedData(ctx, rpcCtx, post, postInfo, &res, req.AnchorCommentID); appError != nil {
		return nil, appError
	}
	return &res, nil
}
