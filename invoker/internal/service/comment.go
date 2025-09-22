package service

import (
	"context"
	"fmt"
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

type CommentService struct {
	*InvokerService
}

func NewCommentService(InvokerService *InvokerService) *CommentService {
	return &CommentService{
		InvokerService: InvokerService,
	}
}

func (s *CommentService) handleComments(ctx context.Context, rpcCtx *rpc.Context, comments []*Comment) ([]*invoker_api.GetCommentsResponse_CommentInfo, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// fetch login info, if has
	loginInfo, _ := slark.LoginInfo(ctx, rpcCtx)
	var list []*invoker_api.GetCommentsResponse_CommentInfo
	for _, comment := range comments {
		// thumbup
		var thumbup bool
		if loginInfo != nil {
			has, err := s.DaoManager.ThumbupDAO.HasThumbupComment(ctx, loginInfo.UserID, comment.ID)
			if err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			thumbup = has
		}
		one := &invoker_api.GetCommentsResponse_CommentInfo{
			Id:            comment.ID,
			PostID:        comment.PostID,
			RootCommentID: comment.RootCommentID,
			PostedAt:      comment.PostedAt,
			PostedBy:      comment.PostedBy,
			AtWho:         comment.AtWho,
			Content:       comment.Content,
			Thumbups:      comment.Thumbups,
			Thumbup:       thumbup,
			Replies:       comment.Replies,
		}
		postedByString, appError := s.fetchUserNickname(ctx, rpcCtx, comment.PostedBy)
		if appError != nil {
			return nil, appError
		}
		one.PostedByString = postedByString
		if comment.AtWho > 0 {
			atWhoString, appError := s.fetchUserNickname(ctx, rpcCtx, comment.AtWho)
			if appError != nil {
				return nil, appError
			}
			one.AtWhoString = atWhoString
		}
		list = append(list, one)
	}
	return list, nil
}

func (s *CommentService) GetComments(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.GetCommentsRequest) (*invoker_api.GetCommentsResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	appError := s.ValidateRequest(ctx, rpcCtx, req)
	if appError != nil {
		return nil, appError
	}
	var total int64
	tx := s.DaoManager.CommentDAO.Table(ctx).Order(`updated_at ASC`)
	// it's a request for the first level comments
	if req.PostID > 0 {
		if req.PageSize > 20 {
			req.PageSize = 20
		}
		tx = tx.Where(`post_id=? AND root_comment_id=0`, req.PostID)
	} else {
		// it's a request for the second level comments
		if req.PageSize > 10 {
			req.PageSize = 10
		}
		tx = tx.Where(`root_comment_id=?`, req.CommentID)
	}
	if err := tx.Count(&total).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var comments []*Comment
	if err := tx.Offset(int(req.PageNumber * req.PageSize)).
		Limit(int(req.PageSize)).
		Find(&comments).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	list, appError := s.handleComments(ctx, rpcCtx, comments)
	if appError != nil {
		return nil, appError
	}
	return &invoker_api.GetCommentsResponse_Data{
		List:  list,
		Total: total,
	}, nil
}

func (s *CommentService) addComment(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.AddCommentRequest, userID int64, post *Post, rootComment *Comment) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		ts := time.Now().UnixMilli()
		if err := daoManager.CommentDAO.Create(ctx, &Comment{
			SiteID:        post.SiteID,
			CategoryID:    post.CategoryID,
			PostID:        post.ID,
			RootCommentID: req.RootCommentID,
			Content:       req.Content,
			PostedAt:      ts,
			PostedBy:      userID,
			AtWho:         req.AtWho,
			UpdatedAt:     ts,
		}); err != nil {
			return err
		}
		if rootComment != nil {
			if err := daoManager.CommentDAO.UpdateByID(ctx, rootComment.ID, &Comment{
				Replies: rootComment.Replies + 1,
			}); err != nil {
				return err
			}
		}
		return daoManager.PostDAO.UpdateByID(ctx, post.ID, &Post{
			Replies:  post.Replies + 1,
			Activity: time.Now().UnixMilli(),
		})
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *CommentService) AddComment(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.AddCommentRequest, userID int64) *gerror.AppError {
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
	var rootComment *Comment
	if req.RootCommentID > 0 {
		rootComment, err = s.DaoManager.CommentDAO.GetByID(ctx, req.RootCommentID)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if rootComment == nil {
			logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
		}
	}
	// if atWho is not empty, need to validate it is not self
	if req.AtWho > 0 {
		// this is not second level comment
		if req.AtWho == userID {
			err := fmt.Errorf("forbid self-commenting")
			logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeForbidden)
		}
	} else {
		// this is the second level comment
		if rootComment != nil && rootComment.PostedBy == userID {
			err := fmt.Errorf("forbid self-commenting")
			logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeForbidden)
		}
	}
	// add replies, update post activity
	return s.addComment(ctx, rpcCtx, req, userID, post, rootComment)
}

func (s *CommentService) EditComment(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.EditCommentRequest, userID int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// validate post
	comment, err := s.DaoManager.CommentDAO.GetByID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if comment == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	if comment.PostedBy != userID {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidUserID")).WithCode(response.StatusCodeForbidden)
	}
	// check post
	post, err := s.DaoManager.PostDAO.GetByID(ctx, comment.PostID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if post == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// edit
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if err := daoManager.CommentDAO.UpdateByID(ctx, comment.ID, &Comment{
			Content:   req.Content,
			UpdatedAt: time.Now().UnixMilli(),
		}); err != nil {
			return err
		}
		// update post activity
		return daoManager.PostDAO.UpdateByID(ctx, post.ID, &Post{
			Activity: time.Now().UnixMilli(),
		})
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *CommentService) deleteComment(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.DeleteCommentRequest, userID int64, post *Post, category *Category, comment *Comment) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	var count int64
	if err := s.DaoManager.CommentDAO.Table(ctx).Where(`root_comment_id=?`, req.Id).Count(&count).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if count > 0 {
		err := fmt.Errorf("comment [id=%d] is a root comment with children comments", req.Id)
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_HasRelatedRecord")).WithCode(response.StatusCodeForbidden)
	}
	var err error
	var rootComment *Comment
	if comment.RootCommentID > 0 {
		rootComment, err = s.DaoManager.CommentDAO.GetByID(ctx, comment.RootCommentID)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if rootComment == nil {
			logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
		}
	}
	// delete
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if err := daoManager.CommentDAO.DeleteByID(ctx, comment.ID); err != nil {
			return err
		}
		if rootComment != nil {
			daoManager.CommentDAO.UpdateByID(ctx, rootComment.ID, &Comment{
				Replies: rootComment.Replies - 1,
			})
		}
		// sub post replies, update post activity
		return daoManager.PostDAO.UpdateByID(ctx, post.ID, &Post{
			Replies:  post.Replies + 1,
			Activity: time.Now().UnixMilli(),
		})
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *CommentService) DeleteComment(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.DeleteCommentRequest, userID int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// validate comment
	comment, err := s.DaoManager.CommentDAO.GetByID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if comment == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// check post
	post, err := s.DaoManager.PostDAO.GetByID(ctx, comment.PostID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if post == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// check category
	category, err := s.DaoManager.CategoryDAO.GetByID(ctx, post.CategoryID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if category == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	if comment.PostedBy != userID {
		// validate site and site admin
		if appError := s.ValidateSiteAdmin(ctx, rpcCtx, category.SiteID, userID); appError != nil {
			logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidUserID")).WithCode(response.StatusCodeForbidden)
		}
	}
	// delete comment
	return s.deleteComment(ctx, rpcCtx, req, userID, post, category, comment)
}
