package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
	"github.com/nextsurfer/pswds_backend/internal/common/slark"
	"github.com/nextsurfer/pswds_backend/internal/dao"
	. "github.com/nextsurfer/pswds_backend/internal/model"
	slark_api "github.com/nextsurfer/slark/api"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BackupService struct {
	*PswdsService
}

func NewBackupService(ctx context.Context, pswdsService *PswdsService) *BackupService {
	return &BackupService{
		PswdsService: pswdsService,
	}
}

const (
	PasswordBackupState_Nothing  = "nothing"
	PasswordBackupState_Upload   = "upload"
	PasswordBackupState_Download = "download"
)

func (s *BackupService) CheckUpdates(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.CheckUpdatesRequest) (*pswds_api.CheckUpdatesResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	// slark login info
	loginInfo, appError := s.ValidateLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	backup, err := s.DaoManager.BackupDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	res := pswds_api.CheckUpdatesResponse_Data{
		State: PasswordBackupState_Nothing,
	}
	if backup == nil {
		res.State = PasswordBackupState_Upload
	} else {
		if backup.UpdatedAt < req.UpdatedAt {
			res.State = PasswordBackupState_Upload
			res.UpdatedAt = backup.UpdatedAt
		} else if backup.UpdatedAt > req.UpdatedAt {
			res.State = PasswordBackupState_Download
		}
	}
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		return &res, nil
	}
	res.HasFamily = true
	sharedDataChecksum, err := s.DaoManager.FamilySharedRecordDAO.GetSharedDataChecksumByFamilyIDAndUserID(ctx, familyMember.FamilyID, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if req.SharedDataChecksum == 0 {
		res.SharedDataUpdated = true
	} else {
		if sharedDataChecksum != req.SharedDataChecksum {
			res.SharedDataUpdated = true
		}
	}
	familyMembers, err := s.DaoManager.FamilyMemberDAO.GetByFamilyID(ctx, familyMember.FamilyID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var otherFamilyMembers []*pswds_api.CheckUpdatesResponse_OtherFamilyMember
	for _, fm := range familyMembers {
		if fm.UserID == loginInfo.UserID {
			continue
		}
		item := &pswds_api.CheckUpdatesResponse_OtherFamilyMember{
			UserID: fm.UserID,
		}
		var email string
		result, appError := s.FetchSlarkInfo(ctx, rpcCtx, item.UserID)
		if appError != nil {
			return nil, appError
		} else {
			email = result.Email
		}
		item.Email = email
		otherFamilyMembers = append(otherFamilyMembers, item)
	}
	res.OtherFamilyMembers = otherFamilyMembers
	return &res, nil
}

func (s *BackupService) uploadData_Backup(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.UploadDataRequest, loginInfo *slark_api.LoginResponse_Data, backup *Backup, daoManager *dao.Manager) error {
	if backup == nil {
		// 新建用户的backup记录
		if err := daoManager.BackupDAO.Create(ctx, &Backup{
			UpdatedAt:          req.UpdatedAt,
			UserID:             loginInfo.UserID,
			PasswordHash:       req.PasswordHash,
			UserPublicKey:      req.UserPublicKey,
			EncryptedFamilyKey: req.EncryptedFamilyKey,
		}); err != nil {
			return nil
		}
	} else {
		newBackup := &Backup{
			UpdatedAt:          req.UpdatedAt,
			PasswordHash:       req.PasswordHash,
			UserPublicKey:      req.UserPublicKey,
			EncryptedFamilyKey: req.EncryptedFamilyKey,
		}
		// 2. 如果passwordHash改变，说明用户修改了解锁密码
		if backup.PasswordHash != req.PasswordHash {
			// 2-1. 删除密保问题密文
			if err := daoManager.BackupDAO.DeleteSecurityQuestions(ctx, loginInfo.UserID); err != nil {
				return err
			}
			// 2-2. 删除可信联络人密文
			if err := daoManager.TrustedContactDAO.DeleteByUserID(ctx, loginInfo.UserID); err != nil {
				return err
			}
			// 2-3. 删除家庭邀请记录，并记录家庭消息
			familyInvitations, err := daoManager.FamilyInvitationDAO.GetAllByEmail(ctx, loginInfo.Email)
			if err != nil {
				return err
			}
			var messages []*FamilyMessage
			for _, familyInvitation := range familyInvitations {
				var invitedBy string
				result, appError := s.FetchSlarkInfo(ctx, rpcCtx, familyInvitation.InvitedBy)
				if appError != nil {
					return appError.Error
				} else {
					invitedBy = result.Email
				}
				messages = append(messages, &FamilyMessage{
					FamilyID:  familyInvitation.FamilyID,
					CreatedBy: loginInfo.Email,
					Target:    invitedBy,
					Operation: dao.FamilyMessageOperationCancelInvitation,
				})
			}
			if err := daoManager.FamilyInvitationDAO.DeleteByEmail(ctx, loginInfo.Email); err != nil {
				return err
			}
			if len(messages) > 0 {
				if err := daoManager.FamilyMessageDAO.Create(ctx, messages); err != nil {
					return err
				}
			}
			// 2-4. 删除家庭备份数据
			if err := daoManager.FamilyBackupDAO.DeleteByCreatedBy(ctx, loginInfo.UserID); err != nil {
				return err
			}
			if err := daoManager.FamilyRecoverDAO.DeleteByUserID(ctx, loginInfo.UserID); err != nil {
				return err
			}
		}
		if err := daoManager.BackupDAO.UpdateByUserID(ctx, loginInfo.UserID, newBackup); err != nil {
			return err
		}
	}
	return nil
}

func (s *BackupService) uploadData(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.UploadDataRequest, loginInfo *slark_api.LoginResponse_Data, backup *Backup) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	var err error
	// 1. slark第二密码检查处理
	if backup == nil {
		// （1）创建用户slark第二密码
		if err = slark.CreateSecondaryPassword(ctx, rpcCtx, req.PasswordHash); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		defer func() {
			// 操作回滚
			if err != nil {
				if err := slark.UpdateSecondaryPassword(ctx, rpcCtx, req.PasswordHash, ""); err != nil {
					logger.Error("internal error", zap.NamedError("appError", err))
					return
				}
			}
		}()
	} else {
		if backup.PasswordHash != req.PasswordHash {
			// （2）更新用户slark第二密码
			if err = slark.UpdateSecondaryPassword(ctx, rpcCtx, backup.PasswordHash, req.PasswordHash); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			defer func() {
				// 操作回滚
				if err != nil {
					if err := slark.UpdateSecondaryPassword(ctx, rpcCtx, req.PasswordHash, backup.PasswordHash); err != nil {
						logger.Error("internal error", zap.NamedError("appError", err))
						return
					}
				}
			}()
		}
	}
	// 2. 本地事务
	if err = s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		// 1. backup
		if err := s.uploadData_Backup(ctx, rpcCtx, req, loginInfo, backup, daoManager); err != nil {
			return err
		}
		// 2. password record
		for _, item := range req.PwdList {
			record, err := daoManager.PasswordRecordDAO.GetByUserIDAndDataID(ctx, loginInfo.UserID, item.DataID)
			if err != nil {
				return err
			}
			if record == nil {
				if err := daoManager.PasswordRecordDAO.Create(ctx, &PasswordRecord{
					DataID:    item.DataID,
					UpdatedAt: item.UpdatedAt,
					UserID:    loginInfo.UserID,
					Content:   item.Content,
					Version:   1,
				}); err != nil {
					return err
				}
			} else {
				if record.UpdatedAt < item.UpdatedAt {
					if err := daoManager.PasswordRecordDAO.UpdateByUserIDAndDataID(ctx, loginInfo.UserID, item.DataID, &PasswordRecord{
						UpdatedAt: item.UpdatedAt,
						Content:   item.Content,
					}); err != nil {
						return err
					}
				}
			}
		}
		// 3. non password record
		for _, item := range req.NprList {
			record, err := daoManager.NonPasswordRecordDAO.GetByUserIDAndDataID(ctx, loginInfo.UserID, item.DataID)
			if err != nil {
				return err
			}
			if record == nil {
				if err := daoManager.NonPasswordRecordDAO.Create(ctx, &NonPasswordRecord{
					DataID:    item.DataID,
					UpdatedAt: item.UpdatedAt,
					UserID:    loginInfo.UserID,
					Type:      item.Type,
					Content:   item.Content,
					Version:   1,
				}); err != nil {
					return err
				}
			} else {
				if record.UpdatedAt < item.UpdatedAt {
					if err := daoManager.NonPasswordRecordDAO.UpdateByUserIDAndDataID(ctx, loginInfo.UserID, item.DataID, &NonPasswordRecord{
						UpdatedAt: item.UpdatedAt,
						Content:   item.Content,
					}); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *BackupService) UploadData(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.UploadDataRequest) *gerror.AppError {
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
	backup, err := s.DaoManager.BackupDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// check updatedAt field
	if backup != nil && backup.UpdatedAt > req.UpdatedAt {
		err = errors.New("backend data pull ahead")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_BackendDataPullAhead")).WithCode(response.StatusCodeDataPullAhead)
	}
	return s.uploadData(ctx, rpcCtx, req, loginInfo, backup)
}

func (s *BackupService) validateBackupRecord(ctx context.Context, rpcCtx *rpc.Context, userID int64) (*Backup, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	backup, err := s.DaoManager.BackupDAO.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if backup == nil {
		err = errors.New("user has no backup")
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoBackup")).WithCode(response.StatusCodeNoBackup)
	}
	return backup, nil
}

func (s *BackupService) DownloadData(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.DownloadDataRequest) (*pswds_api.DownloadDataResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	// slark login info
	loginInfo, appError := s.ValidateLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	// 1. backup
	backup, appError := s.validateBackupRecord(ctx, rpcCtx, loginInfo.UserID)
	if appError != nil {
		return nil, appError
	}
	// check updatedAt field
	if backup.UpdatedAt < req.UpdatedAt {
		err := errors.New("backend data fall behind")
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_BackendDataFallBehind")).WithCode(response.StatusCodeDataFallBehind)
	}
	// 2. password records
	passwordRecords, err := s.DaoManager.PasswordRecordDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var passwords []string
	var pwdIDList []string
	for _, item := range passwordRecords {
		pwdIDList = append(pwdIDList, item.DataID)
		if item.UpdatedAt > req.UpdatedAt {
			passwords = append(passwords, item.Content)
		}
	}
	pwdList := "[]"
	if len(passwords) > 0 {
		data, err := json.Marshal(passwords)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		pwdList = string(data)
	}
	// 3. non password records
	nonPasswordRecords, err := s.DaoManager.NonPasswordRecordDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var nonPasswords []string
	var nprIDList []string
	for _, item := range nonPasswordRecords {
		nprIDList = append(nprIDList, item.DataID)
		if item.UpdatedAt > req.UpdatedAt {
			nonPasswords = append(nonPasswords, item.Content)
		}
	}
	nprList := "[]"
	if len(nonPasswords) > 0 {
		data, err := json.Marshal(nonPasswords)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		nprList = string(data)
	}
	res := pswds_api.DownloadDataResponse_Data{
		UpdatedAt:         backup.UpdatedAt,
		PasswordHash:      backup.PasswordHash,
		SecurityQuestions: backup.SecurityQuestions,
		PwdList:           pwdList,
		PwdIDList:         pwdIDList,
		NprList:           nprList,
		NprIDList:         nprIDList,
	}
	// 4. get family member
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		return &res, nil
	}
	// 5. get shared data
	records, err := s.DaoManager.FamilySharedRecordDAO.GetByFamilyIDAndUserID(ctx, familyMember.FamilyID, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var sharingList []*pswds_api.DownloadDataResponse_ListItem
	var sharedData []*pswds_api.DownloadDataResponse_SharedDataItem
	var allFamilyMembers []*FamilyMember
	for _, record := range records {
		var sharingMembers []int64
		if record.SharedToAll == dao.FamilySharedRecordSharedToAll {
			if len(allFamilyMembers) == 0 {
				allFamilyMembers, err = s.DaoManager.FamilyMemberDAO.GetByFamilyID(ctx, record.FamilyID)
				if err != nil {
					logger.Error("internal error", zap.NamedError("appError", err))
					return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
			}
			for _, member := range allFamilyMembers {
				sharingMembers = append(sharingMembers, member.UserID)
			}
		} else {
			if record.SharingMembers != "" {
				members := make([]int64, 0)
				if err := json.Unmarshal([]byte(record.SharingMembers), &members); err != nil {
					logger.Error("internal error", zap.NamedError("appError", err))
					return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
				sharingMembers = append(sharingMembers, members...)
			}
		}
		if len(sharingMembers) > 0 {
			// 所有与自己相关的分享中的数据
			sharingList = append(sharingList, &pswds_api.DownloadDataResponse_ListItem{
				DataID:      record.DataID,
				UpdatedAt:   record.UpdatedAt,
				FamilyID:    record.FamilyID,
				SharedBy:    record.SharedBy,
				Type:        record.Type,
				Content:     record.Content,
				Version:     record.Version,
				SharedToAll: record.SharedToAll == dao.FamilySharedRecordSharedToAll,
			})
			// 其中由自己分享出去的数据
			if record.SharedBy == loginInfo.UserID {
				var members []int64
				// 清理成员数据
				for _, item := range sharingMembers {
					if item != loginInfo.UserID {
						members = append(members, item)
					}
				}
				sharedData = append(sharedData, &pswds_api.DownloadDataResponse_SharedDataItem{
					DataID:         record.DataID,
					SharedAt:       record.UpdatedAt,
					SharingMembers: members,
				})
			}
		}
	}
	res.EncryptedFamilyKey = backup.EncryptedFamilyKey
	res.SharingList = sharingList
	res.SharedData = sharedData
	return &res, nil
}

func (s *BackupService) CheckUnlockPasswordBackups(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.CheckUnlockPasswordBackupsRequest) (*pswds_api.CheckUnlockPasswordBackupsResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	// 1. 检查用户是否注册
	registration, err := slark.CheckRegistration(ctx, rpcCtx, req.Email)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if registration == nil || registration.Id <= 0 {
		err = fmt.Errorf("the email [%s] has not registered", req.Email)
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_EmailNotRegistered")).WithCode(response.StatusCodeRegisterEmailNotExists)
	}
	// 2. check backup
	backup, appError := s.validateBackupRecord(ctx, rpcCtx, registration.Id)
	if appError != nil {
		return nil, appError
	}
	nullBackup := backup.SecurityQuestions == "" && backup.SecurityQuestionsCiphertext == "" // 没有任何备份
	// 3. security questions recover
	// 3-1. check security questions
	canRecover := backup.SecurityQuestions != "" && backup.SecurityQuestionsCiphertext != ""
	var lastRecoveredAt int64
	var nextRecoverTS int64
	// 3-2. check limit
	securityQuestionsRecover, err := s.DaoManager.UnlockPasswordRecoverDAO.GetSecurityQuestionsRecoverByUserID(ctx, registration.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if securityQuestionsRecover != nil {
		createdAt := time.Unix(securityQuestionsRecover.CreatedAt, 0)
		if time.Since(createdAt) < time.Hour*24 {
			canRecover = false
			lastRecoveredAt = securityQuestionsRecover.CreatedAt
			periodString := os.Getenv("RECOVER_LIMIT_PERIOD")
			if periodString == "" {
				err = errors.New("must set env variable for 'RECOVER_LIMIT_PERIOD'")
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			period, err := strconv.Atoi(periodString)
			if err != nil {
				err = errors.New("invalid value for 'RECOVER_LIMIT_PERIOD'")
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			nextRecoverTS = createdAt.Add(time.Hour * 24 * time.Duration(period)).Unix()
		} else {
			// delete the log
			if err := s.DaoManager.UnlockPasswordRecoverDAO.DeleteByID(ctx, securityQuestionsRecover.ID); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	}
	// 4. family recover
	// 4-1. check family backups
	familyBackups, err := s.DaoManager.FamilyBackupDAO.GetByCreator(ctx, registration.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	nullBackup = nullBackup && len(familyBackups) == 0
	canFamilyRecover := len(familyBackups) > 0
	var lastFamilyRecoveredAt int64
	var nextFamilyRecoverTS int64
	// 4-2. check limit
	familyRecoverLog, err := s.DaoManager.UnlockPasswordRecoverDAO.GetFamilyRecoverByUserID(ctx, registration.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyRecoverLog != nil {
		createdAt := time.Unix(familyRecoverLog.CreatedAt, 0)
		if time.Since(createdAt) < time.Hour*24 {
			canFamilyRecover = false
			lastFamilyRecoveredAt = familyRecoverLog.CreatedAt
			periodString := os.Getenv("RECOVER_LIMIT_PERIOD")
			if periodString == "" {
				err = errors.New("must set env variable for 'RECOVER_LIMIT_PERIOD'")
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			period, err := strconv.Atoi(periodString)
			if err != nil {
				err = errors.New("invalid value for 'RECOVER_LIMIT_PERIOD'")
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			nextFamilyRecoverTS = createdAt.Add(time.Hour * 24 * time.Duration(period)).Unix()
		} else {
			// delete the log
			if err := s.DaoManager.UnlockPasswordRecoverDAO.DeleteByID(ctx, familyRecoverLog.ID); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	}
	var backupMembers []*pswds_api.CheckUnlockPasswordBackupsResponse_Member
	if canFamilyRecover {
		familyBackups, err := s.DaoManager.FamilyBackupDAO.GetByCreator(ctx, registration.Id)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		for _, item := range familyBackups {
			backupMembers = append(backupMembers, &pswds_api.CheckUnlockPasswordBackupsResponse_Member{
				UserID: item.MemberID,
				Email:  item.Member,
			})
		}
	}
	return &pswds_api.CheckUnlockPasswordBackupsResponse_Data{
		NullBackup:            nullBackup,
		CanRecover:            canRecover,
		LastRecoveredAt:       lastRecoveredAt,
		NextRecoverTS:         nextRecoverTS,
		CanFamilyRecover:      canFamilyRecover,
		LastFamilyRecoveredAt: lastFamilyRecoveredAt,
		NextFamilyRecoverTS:   nextFamilyRecoverTS,
		BackupMembers:         backupMembers,
	}, nil
}
