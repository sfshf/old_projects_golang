package service

import (
	"context"
	"errors"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
	"github.com/nextsurfer/pswds_backend/internal/dao"
	. "github.com/nextsurfer/pswds_backend/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NonPasswordService struct {
	*PswdsService
}

func NewNonPasswordService(ctx context.Context, pswdsService *PswdsService) *NonPasswordService {
	return &NonPasswordService{
		PswdsService: pswdsService,
	}
}

func (s *NonPasswordService) CreateNonPasswordRecord(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.CreateNonPasswordRecordRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// slark login info
	loginInfo, appError := s.ValidateLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	record, err := s.DaoManager.NonPasswordRecordDAO.GetByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if record != nil {
		err = errors.New("non password record exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceExists")).WithCode(response.StatusCodeResourceExists)
	}
	backup, appError := s.validateBackupRecord(ctx, rpcCtx, loginInfo.UserID)
	if appError != nil {
		return appError
	}
	if backup.UpdatedAt > req.UpdatedAt {
		err = errors.New("backend data pull ahead")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_BackendDataPullAhead")).WithCode(response.StatusCodeDataPullAhead)
	}
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		// 1. backup
		if err := daoManager.BackupDAO.UpdateByUserID(ctx, loginInfo.UserID, &Backup{
			UpdatedAt: req.UpdatedAt,
		}); err != nil {
			return err
		}
		// 2. non password record
		if err := daoManager.NonPasswordRecordDAO.Create(ctx, &NonPasswordRecord{
			DataID:    req.DataID,
			UpdatedAt: req.UpdatedAt,
			UserID:    loginInfo.UserID,
			Type:      req.Type,
			Content:   req.Content,
			Version:   1,
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *NonPasswordService) validateNonPasswordRecord(ctx context.Context, rpcCtx *rpc.Context, userID int64, dataID string) (*NonPasswordRecord, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	record, err := s.DaoManager.NonPasswordRecordDAO.GetByUserIDAndDataID(ctx, userID, dataID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if record == nil {
		err = errors.New("empty non password record")
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	return record, nil
}

func (s *NonPasswordService) UpdateNonPasswordRecord(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.UpdateNonPasswordRecordRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// slark login info
	loginInfo, appError := s.ValidateLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	record, appError := s.validateNonPasswordRecord(ctx, rpcCtx, loginInfo.UserID, req.DataID)
	if appError != nil {
		return appError
	}
	if record.UpdatedAt > req.UpdatedAt {
		err := errors.New("backend data pull ahead")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_BackendDataPullAhead")).WithCode(response.StatusCodeDataPullAhead)
	}
	backup, appError := s.validateBackupRecord(ctx, rpcCtx, loginInfo.UserID)
	if appError != nil {
		return appError
	}
	if backup.UpdatedAt > req.UpdatedAt {
		err := errors.New("backend data pull ahead")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_BackendDataPullAhead")).WithCode(response.StatusCodeDataPullAhead)
	}
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		// 1. backup
		if err := daoManager.BackupDAO.UpdateByUserID(ctx, loginInfo.UserID, &Backup{
			UpdatedAt: req.UpdatedAt,
		}); err != nil {
			return err
		}
		// 2. non password record
		if err := daoManager.NonPasswordRecordDAO.UpdateByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID, &NonPasswordRecord{
			UpdatedAt: req.UpdatedAt,
			Content:   req.Content,
		}); err != nil {
			return err
		}
		// 3. shared data
		if req.SharedData != "" {
			if err := daoManager.FamilySharedRecordDAO.UpdateByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID, &FamilySharedRecord{
				UpdatedAt: req.UpdatedAt,
				Content:   req.SharedData,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *NonPasswordService) DeleteNonPasswordRecord(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.DeleteNonPasswordRecordRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// slark login info
	loginInfo, appError := s.ValidateLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	if _, appError = s.validateNonPasswordRecord(ctx, rpcCtx, loginInfo.UserID, req.DataID); appError != nil {
		return appError
	}
	backup, appError := s.validateBackupRecord(ctx, rpcCtx, loginInfo.UserID)
	if appError != nil {
		return appError
	}
	if backup.UpdatedAt > req.UpdatedAt {
		err := errors.New("backend data pull ahead")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_BackendDataPullAhead")).WithCode(response.StatusCodeDataPullAhead)
	}
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		// 1. backup
		if err := daoManager.BackupDAO.UpdateByUserID(ctx, loginInfo.UserID, &Backup{
			UpdatedAt: req.UpdatedAt,
		}); err != nil {
			return err
		}
		// 2. password record
		if err := daoManager.NonPasswordRecordDAO.DeleteByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID); err != nil {
			return err
		}
		// 3. shared data
		if err := daoManager.FamilySharedRecordDAO.DeleteByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}
