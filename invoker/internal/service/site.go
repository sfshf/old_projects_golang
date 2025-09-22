package service

import (
	"context"
	"fmt"
	"strings"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	invoker_api "github.com/nextsurfer/invoker/api"
	"github.com/nextsurfer/invoker/api/response"
	"go.uber.org/zap"
)

type SiteService struct {
	*InvokerService
}

func NewSiteService(InvokerService *InvokerService) *SiteService {
	return &SiteService{
		InvokerService: InvokerService,
	}
}

type UserInfo struct {
	ID       int64
	Nickname string
}

func (s *SiteService) GetSiteList(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.GetSiteListRequest) (*invoker_api.GetSiteListResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var list []*invoker_api.GetSiteListResponse_SiteInfo
	sites, err := s.DaoManager.SiteDAO.GetAll(ctx)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for _, site := range sites {
		list = append(list, &invoker_api.GetSiteListResponse_SiteInfo{
			Id:   site.ID,
			Name: site.Name,
		})
	}
	return &invoker_api.GetSiteListResponse_Data{List: list}, nil
}

func (s *SiteService) GetSite(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.GetSiteRequest) (*invoker_api.GetSiteResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	site, err := s.DaoManager.SiteDAO.GetByName(ctx, req.Name)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if site == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	// site admins
	admins, err := s.DaoManager.SiteAdminDAO.GetAdminsBySiteID(ctx, site.ID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &invoker_api.GetSiteResponse_Data{
		Id:     site.ID,
		Name:   site.Name,
		Admins: admins,
	}, nil
}

func (s *SiteService) ValidateSiteAdmin(ctx context.Context, rpcCtx *rpc.Context, siteID, userID int64) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	is, err := s.DaoManager.SiteAdminDAO.ValidateSiteAdmin(ctx, siteID, userID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if !is {
		err := fmt.Errorf("UserID [%d] is not the site [id=%d] administrator", userID, siteID)
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidUserID")).WithCode(response.StatusCodeForbidden)
	}
	return nil
}

type SearchInfo struct {
	PostID                int64  `gorm:"column:postID;"`
	Title                 string `gorm:"column:title;"`
	PostPostedAt          int64  `gorm:"column:postPostedAt;"`
	PostPostedBy          int64  `gorm:"column:postPostedBy;"`
	PostContent           string `gorm:"column:postContent;"`
	CommentID             int64  `gorm:"column:commentID;"`
	CommentContent        string `gorm:"column:commentContent;"`
	CommentPostedAt       int64  `gorm:"column:commentPostedAt;"`
	CommentPostedBy       int64  `gorm:"column:commentPostedBy;"`
	PostTitleMatched      bool   `gorm:"column:postTitleMatched;"`
	PostContentMatched    bool   `gorm:"column:postContentMatched;"`
	CommentContentMatched bool   `gorm:"column:commentContentMatched;"`
}

const (
	SearchInfoType_Post    = "post"
	SearchInfoType_Comment = "comment"
)

func (s *SiteService) shortTitle(src string) string {
	if len(src) > 140 {
		return src[:140] + "......"
	}
	return src
}

func (s *SiteService) shortText(text string, matched string) string {
	index := strings.Index(text, matched)
	var startEllipsis bool
	var endEllipsis bool
	if index < 0 {
		l := len(text)
		if l > 200 {
			l = 200
			endEllipsis = true
		}
		if endEllipsis {
			return text[:l] + "......"
		}
		return text[:l]
	}
	var start int
	if index-100 >= 0 {
		startEllipsis = true
		start = index - 100
	}
	end := len(text)
	if index+len(matched)+100 < len(text) {
		endEllipsis = true
		end = index + len(matched) + 100
	}
	res := text[start:end]
	if startEllipsis {
		res = "......" + res
	}
	if endEllipsis {
		res = res + "......"
	}
	return res
}

func (s *SiteService) generateAggregatedSearchInfoList(ctx context.Context, rpcCtx *rpc.Context, searchInfos []*SearchInfo, searchText string) ([]*invoker_api.AggregatedSearchPageResponse_MatchedInfo, *gerror.AppError) {
	var res []*invoker_api.AggregatedSearchPageResponse_MatchedInfo
	for _, searchInfo := range searchInfos {
		one := &invoker_api.AggregatedSearchPageResponse_MatchedInfo{
			PostID:                searchInfo.PostID,
			Title:                 s.shortTitle(searchInfo.Title),
			PostPostedAt:          searchInfo.PostPostedAt,
			PostPostedBy:          searchInfo.PostPostedBy,
			PostContent:           s.shortText(searchInfo.PostContent, searchText),
			CommentID:             searchInfo.CommentID,
			CommentContent:        s.shortText(searchInfo.CommentContent, searchText),
			CommentPostedAt:       searchInfo.CommentPostedAt,
			CommentPostedBy:       searchInfo.CommentPostedBy,
			PostTitleMatched:      searchInfo.PostTitleMatched,
			PostContentMatched:    searchInfo.PostContentMatched,
			CommentContentMatched: searchInfo.CommentContentMatched,
		}
		postPostedByString, appError := s.fetchUserNickname(ctx, rpcCtx, searchInfo.PostPostedBy)
		if appError != nil {
			return nil, appError
		}
		one.PostPostedByString = postPostedByString
		commentPostedByString, appError := s.fetchUserNickname(ctx, rpcCtx, searchInfo.CommentPostedBy)
		if appError != nil {
			return nil, appError
		}
		one.CommentPostedByString = commentPostedByString
		res = append(res, one)
	}
	return res, nil
}

func (s *SiteService) AggregatedSearchPage(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.AggregatedSearchPageRequest) (*invoker_api.AggregatedSearchPageResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var res invoker_api.AggregatedSearchPageResponse_Data
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
	if req.SearchText == "" {
		return &res, nil
	}
	// search posts and comments
	subquery := s.DaoManager.PostDAO.Table(ctx).
		Select(`post.id AS postID, post.title, post.activity, post.posted_at AS postPostedAt, post.posted_by AS postPostedBy, post.content AS postContent, INSTR(post.title, ?)>0 AS postTitleMatched, INSTR(post.content, ?)>0 AS postContentMatched, INSTR(comment.content, ?)>0 AS commentContentMatched, comment.id AS commentID, comment.content AS commentContent, comment.posted_at AS commentPostedAt, comment.posted_by AS commentPostedBy`,
			req.SearchText, req.SearchText, req.SearchText).
		Joins(`LEFT JOIN comment ON comment.post_id=post.id`).
		Where(`post.site_id=?`, res.SiteInfo.Id).
		Where(s.DaoManager.DB.Or(`post.title LIKE ?`, "%"+req.SearchText+"%").Or(`post.content LIKE ?`, "%"+req.SearchText+"%").Or(`comment.content LIKE ?`, "%"+req.SearchText+"%"))

	postTx := s.DaoManager.DB.Table(`(?) AS searched_post`, subquery).
		Distinct(`postID`, `title`, `activity`, `postPostedAt`, `postPostedBy`, `postContent`, `postTitleMatched`, `postContentMatched`, `commentContentMatched`, `0 AS commentID`, `'' AS commentContent`, `0 AS commentPostedAt`, `0 AS commentPostedBy`).
		Where(`commentContentMatched=false`)
	commentTx := s.DaoManager.DB.Table(`(?) AS searched_comment`, subquery).Select("*").Where(`commentContentMatched=true`)
	tx := s.DaoManager.DB.Table(`(? UNION ALL ?) AS searched_records`, postTx, commentTx).
		Order(`activity DESC`)
	// total
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	res.Total = total
	var searchInfos []*SearchInfo
	if err := tx.Limit(10).Find(&searchInfos).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// handle SearchInfo list
	res.MatchedInfos, appError = s.generateAggregatedSearchInfoList(ctx, rpcCtx, searchInfos, req.SearchText)
	if appError != nil {
		return nil, appError
	}
	return &res, nil
}

func (s *SiteService) generateSearchInfoList(ctx context.Context, rpcCtx *rpc.Context, searchInfos []*SearchInfo, searchText string) ([]*invoker_api.SearchPostCommentResponse_MatchedInfo, *gerror.AppError) {
	var res []*invoker_api.SearchPostCommentResponse_MatchedInfo
	for _, searchInfo := range searchInfos {
		one := &invoker_api.SearchPostCommentResponse_MatchedInfo{
			PostID:                searchInfo.PostID,
			Title:                 s.shortTitle(searchInfo.Title),
			PostPostedAt:          searchInfo.PostPostedAt,
			PostPostedBy:          searchInfo.PostPostedBy,
			PostContent:           s.shortText(searchInfo.PostContent, searchText),
			CommentID:             searchInfo.CommentID,
			CommentContent:        s.shortText(searchInfo.CommentContent, searchText),
			CommentPostedAt:       searchInfo.CommentPostedAt,
			CommentPostedBy:       searchInfo.CommentPostedBy,
			PostTitleMatched:      searchInfo.PostTitleMatched,
			PostContentMatched:    searchInfo.PostContentMatched,
			CommentContentMatched: searchInfo.CommentContentMatched,
		}
		postPostedByString, appError := s.fetchUserNickname(ctx, rpcCtx, searchInfo.PostPostedBy)
		if appError != nil {
			return nil, appError
		}
		one.PostPostedByString = postPostedByString
		commentPostedByString, appError := s.fetchUserNickname(ctx, rpcCtx, searchInfo.CommentPostedBy)
		if appError != nil {
			return nil, appError
		}
		one.CommentPostedByString = commentPostedByString
		res = append(res, one)
	}
	return res, nil
}

func (s *SiteService) SearchPostComment(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.SearchPostCommentRequest) (*invoker_api.SearchPostCommentResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	siteInfo, appError := s.aggregatedSiteInfo(ctx, rpcCtx, req.Site)
	if appError != nil {
		return nil, appError
	}
	var res invoker_api.SearchPostCommentResponse_Data
	subquery := s.DaoManager.PostDAO.Table(ctx).
		Select(`post.id AS postID, post.title, post.activity, post.posted_at AS postPostedAt, post.posted_by AS postPostedBy, post.content AS postContent, INSTR(post.title, ?)>0 AS postTitleMatched, INSTR(post.content, ?)>0 AS postContentMatched, INSTR(comment.content, ?)>0 AS commentContentMatched, comment.id AS commentID, comment.content AS commentContent, comment.posted_at AS commentPostedAt, comment.posted_by AS commentPostedBy`,
			req.SearchText, req.SearchText, req.SearchText).
		Joins(`LEFT JOIN comment ON comment.post_id=post.id`).
		Where(`post.site_id=?`, siteInfo.Id).
		Where(s.DaoManager.DB.Or(`post.title LIKE ?`, "%"+req.SearchText+"%").Or(`post.content LIKE ?`, "%"+req.SearchText+"%").Or(`comment.content LIKE ?`, "%"+req.SearchText+"%"))

	postTx := s.DaoManager.DB.Table(`(?) AS searched_post`, subquery).
		Distinct(`postID`, `title`, `activity`, `postPostedAt`, `postPostedBy`, `postContent`, `postTitleMatched`, `postContentMatched`, `commentContentMatched`, `0 AS commentID`, `'' AS commentContent`, `0 AS commentPostedAt`, `0 AS commentPostedBy`).
		Where(`commentContentMatched=false`)
	commentTx := s.DaoManager.DB.Table(`(?) AS searched_comment`, subquery).Select("*").Where(`commentContentMatched=true`)
	tx := s.DaoManager.DB.Table(`(? UNION ALL ?) AS searched_records`, postTx, commentTx).
		Order(`activity DESC`)
	// total
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	res.Total = total
	var searchInfos []*SearchInfo
	if err := tx.Offset(int(req.PageNumber * req.PageSize)).
		Limit(int(req.PageSize)).Find(&searchInfos).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// handle SearchInfo list
	res.List, appError = s.generateSearchInfoList(ctx, rpcCtx, searchInfos, req.SearchText)
	if appError != nil {
		return nil, appError
	}
	return &res, nil
}
