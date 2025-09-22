package service

import (
	"context"
	"time"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	invoker_api "github.com/nextsurfer/invoker/api"
	"github.com/nextsurfer/invoker/api/response"
	"github.com/nextsurfer/invoker/internal/dao"
	. "github.com/nextsurfer/invoker/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserService struct {
	*InvokerService
}

func NewUserService(InvokerService *InvokerService) *UserService {
	return &UserService{
		InvokerService: InvokerService,
	}
}

func (s *UserService) thumbupPost(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.ThumbupPostRequest, userID int64, post *Post, hasThumbup bool) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	if !hasThumbup {
		if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
			daoManager := dao.ManagerWithDB(tx)
			// add a thumbup
			if err := daoManager.ThumbupDAO.Create(ctx, &Thumbup{
				SiteID:     post.SiteID,
				CategoryID: post.CategoryID,
				Type:       dao.ThumbupType_Post,
				PostID:     post.ID,
				PostedAt:   time.Now().UnixMilli(),
				PostedBy:   userID,
			}); err != nil {
				return err
			}
			// add post thumbups
			return daoManager.PostDAO.UpdateByID(
				ctx,
				post.ID,
				&Post{
					Thumbups: post.Thumbups + 1,
				})
		}); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		return nil
	}
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		// remove the thumbup
		if err := daoManager.ThumbupDAO.DeletePostThumbup(ctx, userID, post.ID); err != nil {
			return err
		}
		// sub post thumbups
		return daoManager.PostDAO.UpdateByID(
			ctx,
			post.ID,
			&Post{
				Thumbups: post.Thumbups - 1,
			})
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *UserService) ThumbupPost(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.ThumbupPostRequest, userID int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// check post
	post, err := s.DaoManager.PostDAO.GetByID(ctx, req.PostID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if post == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// check and update thumbup
	hasThumbup, err := s.DaoManager.ThumbupDAO.HasThumbupPost(ctx, userID, post.ID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return s.thumbupPost(ctx, rpcCtx, req, userID, post, hasThumbup)
}

func (s *UserService) thumbupComment(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.ThumbupCommentRequest, userID int64, comment *Comment, hasThumbup bool) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	if !hasThumbup {
		if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
			daoManager := dao.ManagerWithDB(tx)
			// add a thumbup
			if err := daoManager.ThumbupDAO.Create(ctx, &Thumbup{
				SiteID:     comment.SiteID,
				CategoryID: comment.CategoryID,
				PostID:     comment.PostID,
				Type:       dao.ThumbupType_Comment,
				CommentID:  comment.ID,
				PostedAt:   time.Now().UnixMilli(),
				PostedBy:   userID,
			}); err != nil {
				return err
			}
			// add comment thumbups
			return daoManager.CommentDAO.UpdateByID(
				ctx,
				comment.ID,
				&Comment{
					Thumbups: comment.Thumbups + 1,
				})
		}); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		return nil
	}
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		// remove the thumbup
		if err := daoManager.ThumbupDAO.DeleteCommentThumbup(ctx, userID, comment.ID); err != nil {
			return err
		}
		// sub comment thumbups
		return daoManager.PostDAO.UpdateByID(
			ctx,
			comment.ID,
			&Post{
				Thumbups: comment.Thumbups - 1,
			})
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *UserService) ThumbupComment(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.ThumbupCommentRequest, userID int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// check comment
	comment, err := s.DaoManager.CommentDAO.GetByID(ctx, req.CommentID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if comment == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// check and update thumbup
	hasThumbup, err := s.DaoManager.ThumbupDAO.HasThumbupComment(ctx, userID, comment.ID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return s.thumbupComment(ctx, rpcCtx, req, userID, comment, hasThumbup)
}

func (s *UserService) handleHistoryPost(posts []*Post, postedByString string) []*invoker_api.PostHistoryResponse_PostInfo {
	var list []*invoker_api.PostHistoryResponse_PostInfo
	for _, post := range posts {
		list = append(list, &invoker_api.PostHistoryResponse_PostInfo{
			Id:             post.ID,
			CategoryID:     post.CategoryID,
			Title:          s.shortTitle(post.Title),
			PostedAt:       post.PostedAt,
			PostedBy:       post.PostedBy,
			PostedByString: postedByString,
			Views:          post.Views,
			Replies:        post.Replies,
			Activity:       post.Activity,
		})
	}
	return list
}

func (s *UserService) PostHistory(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.PostHistoryRequest) (*invoker_api.PostHistoryResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var res invoker_api.PostHistoryResponse_Data
	var appError *gerror.AppError
	// fetch login info, if has
	res.LoginInfo, appError = s.aggregatedLoginInfo(ctx, rpcCtx, false)
	if appError != nil {
		return nil, appError
	}
	// fetch site info
	res.SiteInfo, appError = s.aggregatedSiteInfo(ctx, rpcCtx, req.Site)
	if appError != nil {
		return nil, appError
	}
	// post list
	tx := s.DaoManager.PostDAO.Table(ctx).
		Where(`site_id=?`, res.SiteInfo.Id).
		Where(`state=?`, dao.PostState_Posted).
		Where(`posted_by=?`, res.LoginInfo.UserID)
	// total
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	res.Total = total
	var posts []*Post
	if err := tx.Order(`posted_at DESC`).
		Offset(int(req.PageNumber * req.PageSize)).
		Limit(int(req.PageSize)).
		Find(&posts).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// list
	res.List = s.handleHistoryPost(posts, res.LoginInfo.Nickname)
	return &res, nil
}

func (s *UserService) shortTitle(src string) string {
	if len(src) > 140 {
		return src[:140] + "......"
	}
	return src
}

func (s *UserService) shortText(src string) string {
	if len(src) > 200 {
		return src[:200] + "......"
	}
	return src
}

func (s *UserService) handleHistoryComment(ctx context.Context, rpcCtx *rpc.Context, comments []*HistoryComment, postedByString string) ([]*invoker_api.CommentHistoryResponse_CommentInfo, *gerror.AppError) {
	var list []*invoker_api.CommentHistoryResponse_CommentInfo
	for _, comment := range comments {
		var atWhoString string
		if comment.AtWho > 0 {
			data, appError := s.fetchUserNickname(ctx, rpcCtx, comment.AtWho)
			if appError != nil {
				return nil, appError
			}
			atWhoString = data
		}
		list = append(list, &invoker_api.CommentHistoryResponse_CommentInfo{
			Id:             comment.ID,
			PostID:         comment.PostID,
			Title:          s.shortTitle(comment.Title),
			RootCommentID:  comment.RootCommentID,
			PostedAt:       comment.PostedAt,
			PostedBy:       comment.PostedBy,
			PostedByString: postedByString,
			AtWho:          comment.AtWho,
			AtWhoString:    atWhoString,
			Content:        s.shortText(comment.Content),
			Replies:        comment.Replies,
		})
	}
	return list, nil
}

type HistoryComment struct {
	ID            int64  `gorm:"column:id;"`
	PostID        int64  `gorm:"column:post_id;"`
	Title         string `gorm:"column:title;"`
	RootCommentID int64  `gorm:"column:root_comment_id;"`
	Content       string `gorm:"column:content;"`
	PostedAt      int64  `gorm:"column:posted_at;"`
	PostedBy      int64  `gorm:"column:posted_by;"`
	AtWho         int64  `gorm:"column:at_who;"`
	UpdatedAt     int64  `gorm:"column:updated_at;"`
	Replies       int64  `gorm:"column:replies;"`
	Thumbups      int64  `gorm:"column:thumbups;"`
}

func (s *UserService) CommentHistory(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.CommentHistoryRequest) (*invoker_api.CommentHistoryResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var res invoker_api.CommentHistoryResponse_Data
	var appError *gerror.AppError
	// fetch login info, if has
	res.LoginInfo, appError = s.aggregatedLoginInfo(ctx, rpcCtx, false)
	if appError != nil {
		return nil, appError
	}
	// fetch site info
	res.SiteInfo, appError = s.aggregatedSiteInfo(ctx, rpcCtx, req.Site)
	if appError != nil {
		return nil, appError
	}
	// comment list
	tx := s.DaoManager.CommentDAO.Table(ctx).
		Select(`comment.id, comment.post_id, comment.root_comment_id, comment.content, comment.posted_at, comment.posted_by, comment.at_who, comment.updated_at, comment.replies, comment.thumbups, post.title`).
		Joins(`LEFT JOIN post ON post.id=comment.post_id`).
		Where(`comment.site_id=?`, res.SiteInfo.Id).
		Where(`comment.posted_by=?`, res.LoginInfo.UserID)
	// total
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	res.Total = total
	var comments []*HistoryComment
	if err := tx.Order(`comment.posted_at DESC`).
		Offset(int(req.PageNumber * req.PageSize)).
		Limit(int(req.PageSize)).
		Find(&comments).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	list, appError := s.handleHistoryComment(ctx, rpcCtx, comments, res.LoginInfo.Nickname)
	if appError != nil {
		return nil, appError
	}
	res.List = list
	return &res, nil
}

type HistoryThumbup struct {
	ID        int64  `gorm:"column:id;"`
	PostID    int64  `gorm:"column:post_id;"`
	Title     string `gorm:"column:title;"`
	CommentID int64  `gorm:"column:comment_id;"`
	Type      int32  `gorm:"column:type;"`
	PostedAt  int64  `gorm:"column:posted_at;"`
	PostedBy  int64  `gorm:"column:posted_by;"`
	Content   string `gorm:"column:content;"`
}

func (s *UserService) ThumbupHistory(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.ThumbupHistoryRequest) (*invoker_api.ThumbupHistoryResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var res invoker_api.ThumbupHistoryResponse_Data
	var appError *gerror.AppError
	// fetch login info, if has
	res.LoginInfo, appError = s.aggregatedLoginInfo(ctx, rpcCtx, false)
	if appError != nil {
		return nil, appError
	}
	// fetch site info
	res.SiteInfo, appError = s.aggregatedSiteInfo(ctx, rpcCtx, req.Site)
	if appError != nil {
		return nil, appError
	}
	// thumbup list
	postTx := s.DaoManager.ThumbupDAO.Table(ctx).
		Select(`thumbup.id, thumbup.post_id, thumbup.comment_id, thumbup.type, thumbup.posted_at, thumbup.posted_by, post.title, post.content AS content`).
		Joins(`LEFT JOIN post ON post.id=thumbup.post_id`).
		Where(`thumbup.type=? AND thumbup.site_id=? AND thumbup.posted_by=?`,
			dao.ThumbupType_Post, res.SiteInfo.Id, res.LoginInfo.UserID)
	commentTx := s.DaoManager.ThumbupDAO.Table(ctx).
		Select(`thumbup.id, thumbup.post_id, thumbup.comment_id, thumbup.type, thumbup.posted_at, thumbup.posted_by, post.title, comment.content AS content`).
		Joins(`LEFT JOIN comment ON comment.id=thumbup.comment_id LEFT JOIN post ON post.id=comment.post_id`).
		Where(`thumbup.type=? AND thumbup.site_id=? AND thumbup.posted_by=?`,
			dao.ThumbupType_Comment, res.SiteInfo.Id, res.LoginInfo.UserID)
	tx := s.DaoManager.DB.Table(`(? UNION ALL ?) AS thumbup_records`, postTx, commentTx).
		Order(`posted_at DESC`)
		// total
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	res.Total = total
	var thumbups []*HistoryThumbup
	if err := tx.Offset(int(req.PageNumber * req.PageSize)).
		Limit(int(req.PageSize)).
		Find(&thumbups).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// list
	var list []*invoker_api.ThumbupHistoryResponse_ThumbupInfo
	for _, thumbup := range thumbups {
		list = append(list, &invoker_api.ThumbupHistoryResponse_ThumbupInfo{
			Id:             thumbup.ID,
			PostID:         thumbup.PostID,
			Title:          s.shortTitle(thumbup.Title),
			Type:           dao.ThumbupTypeString[thumbup.Type],
			CommentID:      thumbup.CommentID,
			PostedAt:       thumbup.PostedAt,
			PostedBy:       thumbup.PostedBy,
			PostedByString: res.LoginInfo.Nickname,
			Content:        s.shortText(thumbup.Content),
		})
	}
	res.List = list
	return &res, nil
}
