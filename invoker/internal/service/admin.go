package service

import (
	"context"
	"slices"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	invoker_api "github.com/nextsurfer/invoker/api"
	"github.com/nextsurfer/invoker/api/response"
	"github.com/nextsurfer/invoker/internal/common/connector"
	"github.com/nextsurfer/invoker/internal/common/slark"
	"github.com/nextsurfer/invoker/internal/dao"
	. "github.com/nextsurfer/invoker/internal/model"
	"github.com/shirou/gopsutil/disk"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

type AdminService struct {
	*InvokerService
}

func NewAdminService(InvokerService *InvokerService) *AdminService {
	return &AdminService{
		InvokerService: InvokerService,
	}
}

func (s *AdminService) GetSites(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.GetSitesRequest) (*invoker_api.GetSitesResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidApiKey")).WithCode(response.StatusCodeForbidden)
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	sites, err := s.DaoManager.SiteDAO.GetAll(ctx)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var list []*invoker_api.GetSitesResponse_SiteInfo
	for _, site := range sites {
		list = append(list, &invoker_api.GetSitesResponse_SiteInfo{
			Id:   site.ID,
			Name: site.Name,
		})
	}
	return &invoker_api.GetSitesResponse_Data{List: list}, nil
}

func (s *AdminService) AddSite(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.AddSiteRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidApiKey")).WithCode(response.StatusCodeForbidden)
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	site, err := s.DaoManager.SiteDAO.GetByName(ctx, req.Name)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if site == nil {
		if err := s.DaoManager.SiteDAO.Create(ctx, &Site{
			Name: req.Name,
		}); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return nil
}

func (s *AdminService) EditSite(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.EditSiteRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidApiKey")).WithCode(response.StatusCodeForbidden)
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// check site
	site, err := s.DaoManager.SiteDAO.GetByID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if site == nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	if err := s.DaoManager.SiteDAO.UpdateByID(ctx, site.ID, &Site{
		Name: req.Name,
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *AdminService) DeleteSite(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.DeleteSiteRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidApiKey")).WithCode(response.StatusCodeForbidden)
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if err := daoManager.SiteDAO.DeleteByID(ctx, req.Id); err != nil {
			return err
		}
		// delete related data
		// 1. admins
		if err := daoManager.SiteAdminDAO.DeleteBySiteID(ctx, req.Id); err != nil {
			return err
		}
		// 2. category
		if err := daoManager.CategoryDAO.DeleteBySiteID(ctx, req.Id); err != nil {
			return err
		}
		// 3. post
		if err := daoManager.PostDAO.DeleteBySiteID(ctx, req.Id); err != nil {
			return err
		}
		// 4. comment
		if err := daoManager.CommentDAO.DeleteBySiteID(ctx, req.Id); err != nil {
			return err
		}
		// 5. thumbup
		if err := daoManager.ThumbupDAO.DeleteBySiteID(ctx, req.Id); err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *AdminService) GetSiteAdmins(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.GetSiteAdminsRequest) (*invoker_api.GetSiteAdminsResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidApiKey")).WithCode(response.StatusCodeForbidden)
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	admins, err := s.DaoManager.SiteAdminDAO.GetAdminsBySiteID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if len(admins) == 0 {
		return nil, nil
	}
	var list []*invoker_api.GetSiteAdminsResponse_AdminInfo
	valids, err := slark.ValidateUserIDs(ctx, rpcCtx, admins)
	if err != nil || len(valids) != len(admins) {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for idx, userID := range admins {
		list = append(list, &invoker_api.GetSiteAdminsResponse_AdminInfo{
			UserID: userID,
			Valid:  valids[idx],
		})
	}
	return &invoker_api.GetSiteAdminsResponse_Data{List: list}, nil
}

func (s *AdminService) AddSiteAdmin(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.AddSiteAdminRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidApiKey")).WithCode(response.StatusCodeForbidden)
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	admins, err := s.DaoManager.SiteAdminDAO.GetAdminsBySiteID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// check exists
	if slices.Contains(admins, req.UserID) {
		return nil
	}
	// validate userID
	userIDs := []int64{req.UserID}
	valids, err := slark.ValidateUserIDs(ctx, rpcCtx, userIDs)
	if err != nil || len(valids) != len(userIDs) {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if !valids[0] {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidUserID")).WithCode(response.StatusCodeWrongParameters)
	}
	// add a site admin record
	if err := s.DaoManager.SiteAdminDAO.Create(ctx, &SiteAdmin{
		SiteID: req.Id,
		UserID: req.UserID,
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *AdminService) DeleteSiteAdmin(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.DeleteSiteAdminRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidApiKey")).WithCode(response.StatusCodeForbidden)
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	if err := s.DaoManager.SiteAdminDAO.DeleteBySiteIDAndUserID(ctx, req.Id, req.UserID); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *AdminService) GetStorageInfo(ctx context.Context, rpcCtx *rpc.Context, req *invoker_api.GetStorageInfoRequest) (*invoker_api.GetStorageInfoResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_InvalidApiKey")).WithCode(response.StatusCodeForbidden)
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	// system info
	d, err := disk.Usage("/")
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	systemInfo := &invoker_api.GetStorageInfoResponse_SystemInfo{
		Total: d.Total / 1024 / 1024 / 1024,
		Free:  d.Free / 1024 / 1024 / 1024,
	}
	// database infos
	var databaseInfos []*invoker_api.GetStorageInfoResponse_DatabaseInfo
	if err := s.DaoManager.DB.Raw(`SELECT table_schema AS "database", SUM(data_length + index_length) / 1024 / 1024 AS "size" FROM information_schema.TABLES GROUP BY table_schema;`).
		Scan(&databaseInfos).Error; err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &invoker_api.GetStorageInfoResponse_Data{
		SystemInfo:    systemInfo,
		DatabaseInfos: databaseInfos,
	}, nil
}
