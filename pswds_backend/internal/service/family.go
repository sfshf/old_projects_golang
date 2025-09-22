package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
	"github.com/nextsurfer/pswds_backend/internal/common/email"
	"github.com/nextsurfer/pswds_backend/internal/common/random"
	"github.com/nextsurfer/pswds_backend/internal/common/slark"
	"github.com/nextsurfer/pswds_backend/internal/dao"
	. "github.com/nextsurfer/pswds_backend/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FamilyService struct {
	*PswdsService
}

func NewFamilyService(ctx context.Context, pswdsService *PswdsService) *FamilyService {
	return &FamilyService{
		PswdsService: pswdsService,
	}
}

func (s *FamilyService) DownloadSharedData(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.DownloadSharedDataRequest) (*pswds_api.DownloadSharedDataResponse_Data, *gerror.AppError) {
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
	// 1. get family member
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		return nil, nil
	}
	// 2. get shared data
	records, err := s.DaoManager.FamilySharedRecordDAO.GetByFamilyIDAndUserID(ctx, familyMember.FamilyID, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var sharingList []*pswds_api.DownloadSharedDataResponse_ListItem
	var sharedData []*pswds_api.DownloadSharedDataResponse_SharedDataItem
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
			sharingList = append(sharingList, &pswds_api.DownloadSharedDataResponse_ListItem{
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
				sharedData = append(sharedData, &pswds_api.DownloadSharedDataResponse_SharedDataItem{
					DataID:         record.DataID,
					SharedAt:       record.UpdatedAt,
					SharingMembers: members,
				})
			}
		}
	}
	// 3. get encryptedFaimlyKey
	backup, err := s.DaoManager.BackupDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if backup == nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoBackup")).WithCode(response.StatusCodeNoBackup)
	}
	return &pswds_api.DownloadSharedDataResponse_Data{SharingList: sharingList, EncryptedFamilyKey: backup.EncryptedFamilyKey, SharedData: sharedData}, nil
}

func (s *FamilyService) CreateFamily(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.CreateFamilyRequest) *gerror.AppError {
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
	// 1. check if family already exists
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember != nil {
		err = errors.New("the user's family already exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceExists")).WithCode(response.StatusCodeResourceExists)
	}
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		familyID := random.NewUUIDString()
		// 2. create family
		if err := daoManager.FamilyDAO.Create(ctx, &Family{
			CreatedBy:   loginInfo.UserID,
			FamilyID:    familyID,
			Description: req.Description,
		}); err != nil {
			return err
		}
		// 3. create family member
		if err := daoManager.FamilyMemberDAO.Create(ctx, &FamilyMember{
			UserID:   loginInfo.UserID,
			FamilyID: familyID,
			IsAdmin:  dao.FamilyMemberIsAdmin,
		}); err != nil {
			return err
		}
		// 4. update backup
		if err := daoManager.BackupDAO.UpdateByUserID(ctx, loginInfo.UserID, &Backup{
			EncryptedFamilyKey: req.EncryptedFamilyKey,
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

func (s *FamilyService) GetFamilyInfo(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.GetFamilyInfoRequest) (*pswds_api.GetFamilyInfoResponse_Data, *gerror.AppError) {
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
	// 1. get from family member table
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		return &pswds_api.GetFamilyInfoResponse_Data{
			HasFamily: false,
		}, nil
	}
	// 2. get from family table
	family, err := s.DaoManager.FamilyDAO.GetByFamilyID(ctx, familyMember.FamilyID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if family == nil {
		err = fmt.Errorf("family %s not exists", familyMember.FamilyID)
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	// 3. get family members
	familyMembers, err := s.DaoManager.FamilyMemberDAO.GetWithUserPublicKeyByFamilyID(ctx, familyMember.FamilyID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var members []*pswds_api.GetFamilyInfoResponse_FamilyMember
	for _, item := range familyMembers {
		var email string
		result, appError := s.FetchSlarkInfo(ctx, rpcCtx, item.UserID)
		if appError != nil {
			return nil, appError
		} else {
			email = result.Email
		}
		members = append(members, &pswds_api.GetFamilyInfoResponse_FamilyMember{
			Id:            item.ID,
			UserID:        item.UserID,
			Email:         email,
			FamilyID:      item.FamilyID,
			JoinedAt:      item.CreatedAt,
			IsAdmin:       item.IsAdmin == dao.FamilyMemberIsAdmin,
			UserPublicKey: item.UserPublicKey,
		})
	}
	// 4. count shared data
	count, err := s.DaoManager.FamilySharedRecordDAO.CountByFamilyID(ctx, family.FamilyID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &pswds_api.GetFamilyInfoResponse_Data{
		HasFamily:     true,
		Description:   family.Description,
		FamilyMembers: members,
		SharedNumbers: count,
	}, nil
}

const (
	InvitationState_NoUser     = "no_user"
	InvitationState_HasFamily  = "has_family"
	InvitationState_HasInvited = "has_invited"
	InvitationState_Invitable  = "invitable"
)

func (s *FamilyService) CheckUserAvailable(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.CheckUserAvailableRequest) (*pswds_api.CheckUserAvailableResponse_Data, *gerror.AppError) {
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
	// check family admin
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		err = errors.New("the user's family not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	if familyMember.IsAdmin != dao.FamilyMemberIsAdmin {
		err = errors.New("the user is not an admin")
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_Unauthorized")).WithCode(response.StatusCodeUnauthorized)
	}
	var res pswds_api.CheckUserAvailableResponse_Data
	// 1. check email account
	var slarkID int64
	slarkInfo, appError := s.FetchSlarkInfoByEmail(ctx, rpcCtx, req.Email)
	if appError != nil {
		return nil, appError
	}
	if slarkInfo != nil {
		slarkID = slarkInfo.UserID
	}
	if slarkID == 0 {
		registration, err := slark.CheckRegistration(ctx, rpcCtx, req.Email)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if registration == nil || registration.Id <= 0 {
			res.State = InvitationState_NoUser
			return &res, nil
		}
		slarkID = registration.Id
		slarkInfo := &SlarkInfo{
			UserID:   registration.Id,
			Email:    req.Email,
			Nickname: registration.Nickname,
		}
		data, err := json.Marshal(slarkInfo)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if err := s.RedisClient.Set(ctx, fmt.Sprintf("%s%d", RedisKeyPrefixSlarkInfo, registration.Id), string(data), 20*time.Minute).Err(); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	// 2. check family member
	invitedFamilyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, slarkID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if invitedFamilyMember != nil {
		res.State = InvitationState_HasFamily
		return &res, nil
	}
	// 3. check family invitation
	invitation, err := s.DaoManager.FamilyInvitationDAO.GetByFamilyIDAndEmail(ctx, familyMember.FamilyID, req.Email)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if invitation != nil {
		res.State = InvitationState_HasInvited
		return &res, nil
	}
	res.State = InvitationState_Invitable
	// 4. get user public key
	invitedBackup, err := s.DaoManager.BackupDAO.GetByUserID(ctx, slarkID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if invitedBackup == nil {
		res.State = InvitationState_NoUser
		return &res, nil
	}
	res.UserPublicKey = invitedBackup.UserPublicKey
	return &res, nil
}

func (s *FamilyService) InviteFamilyMember(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.InviteFamilyMembersRequest) *gerror.AppError {
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
	// 1. check family admin
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		err = errors.New("the user's family not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	if familyMember.IsAdmin != dao.FamilyMemberIsAdmin {
		err = errors.New("the user is not an admin")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_Unauthorized")).WithCode(response.StatusCodeUnauthorized)
	}
	// 2. check family members' number
	count, err := s.DaoManager.FamilyMemberDAO.CountByFamilyID(ctx, familyMember.FamilyID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if count == 10 {
		err = errors.New("the family members' number has reached the limit")
		logger.Info("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceLimit")).WithCode(response.StatusCodeResourceLimit)
	}
	// 3. create invitation record
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if err := daoManager.FamilyInvitationDAO.Create(ctx, &FamilyInvitation{
			FamilyID:           familyMember.FamilyID,
			InvitedBy:          loginInfo.UserID,
			Email:              req.Email,
			EncryptedFamilyKey: req.EncryptedFamilyKey,
		}); err != nil {
			return err
		}
		if err := daoManager.FamilyMessageDAO.Create(ctx, &FamilyMessage{
			FamilyID:  familyMember.FamilyID,
			CreatedBy: loginInfo.Email,
			Target:    req.Email,
			Operation: dao.FamilyMessageOperationInviteMember,
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

func (s *FamilyService) CheckFamilyInvitation(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.CheckFamilyInvitationRequest) (*pswds_api.CheckFamilyInvitationResponse_Data, *gerror.AppError) {
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
	// 1. get invitation from db
	invitation, err := s.DaoManager.FamilyInvitationDAO.GetByEmail(ctx, loginInfo.Email)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if invitation == nil {
		return &pswds_api.CheckFamilyInvitationResponse_Data{
			HasInvitation: false,
		}, nil
	}
	var email string
	result, appError := s.FetchSlarkInfo(ctx, rpcCtx, invitation.InvitedBy)
	if appError != nil {
		return nil, appError
	} else {
		email = result.Email
	}
	return &pswds_api.CheckFamilyInvitationResponse_Data{
		HasInvitation:      true,
		Id:                 invitation.ID,
		InvitedBy:          email,
		InvitedAt:          invitation.CreatedAt,
		EncryptedFamilyKey: invitation.EncryptedFamilyKey,
	}, nil
}

func (s *FamilyService) ProcessFamilyInvitation(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.ProcessFamilyInvitationRequest) *gerror.AppError {
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
	// 1. get family invitation
	invitation, err := s.DaoManager.FamilyInvitationDAO.GetByID(ctx, req.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if invitation == nil {
		err = errors.New("the user's invitation not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	if invitation.Email != loginInfo.Email {
		err = fmt.Errorf("invalid session email")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	var invitedBy string
	result, appError := s.FetchSlarkInfo(ctx, rpcCtx, invitation.InvitedBy)
	if appError != nil {
		return appError
	} else {
		invitedBy = result.Email
	}
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if req.Accept {
			// 2-1. accept invitation
			if err := daoManager.FamilyMemberDAO.Create(ctx, &FamilyMember{
				UserID:   loginInfo.UserID,
				FamilyID: invitation.FamilyID,
				IsAdmin:  dao.FamilyMemberIsNotAdmin,
			}); err != nil {
				return err
			}
			if err := daoManager.BackupDAO.UpdateByUserID(ctx, loginInfo.UserID, &Backup{
				EncryptedFamilyKey: invitation.EncryptedFamilyKey,
			}); err != nil {
				return err
			}
			if err := daoManager.FamilyInvitationDAO.DeleteByID(ctx, invitation.ID); err != nil {
				return err
			}

			if err := daoManager.FamilyMessageDAO.Create(ctx, &FamilyMessage{
				FamilyID:  invitation.FamilyID,
				CreatedBy: loginInfo.Email,
				Target:    invitedBy,
				Operation: dao.FamilyMessageOperationAcceptInvitation,
			}); err != nil {
				return err
			}
		} else {
			// 2-2. reject invitation
			if err := daoManager.FamilyInvitationDAO.DeleteByID(ctx, invitation.ID); err != nil {
				return err
			}
			if err := daoManager.FamilyMessageDAO.Create(ctx, &FamilyMessage{
				FamilyID:  invitation.FamilyID,
				CreatedBy: loginInfo.Email,
				Target:    invitedBy,
				Operation: dao.FamilyMessageOperationRejectInvitation,
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

func (s *FamilyService) RemoveFamilyMember(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.RemoveFamilyMemberRequest) *gerror.AppError {
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
	// 1. check whether the user is an admin
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		err = errors.New("the user's family not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	if familyMember.IsAdmin != dao.FamilyMemberIsAdmin {
		err = errors.New("the user is not an admin")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_Unauthorized")).WithCode(response.StatusCodeUnauthorized)
	}
	// 2. check param userID
	otherFamilyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, req.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if otherFamilyMember == nil {
		err = errors.New("the other user's family not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	if otherFamilyMember.FamilyID != familyMember.FamilyID {
		err = errors.New("the users are not in one family")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_Unauthorized")).WithCode(response.StatusCodeUnauthorized)
	}
	if otherFamilyMember.IsAdmin == dao.FamilyMemberIsAdmin {
		err = errors.New("the other user is a family administrator")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_Unauthorized")).WithCode(response.StatusCodeUnauthorized)
	}
	var target string
	result, appError := s.FetchSlarkInfo(ctx, rpcCtx, req.UserID)
	if appError != nil {
		return appError
	} else {
		target = result.Email
	}
	// 3. delete
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		// (1) the family member
		if err := daoManager.FamilyMemberDAO.DeleteByFamilyIDAndUserID(ctx, familyMember.FamilyID, req.UserID); err != nil {
			return err
		}
		// (2) sharing data
		if err := daoManager.FamilySharedRecordDAO.DeleteMemberByFamilyIDAndUserID(ctx, familyMember.FamilyID, req.UserID); err != nil {
			return err
		}
		// (3) family recover
		if err := daoManager.FamilyRecoverDAO.DeleteByUserID(ctx, req.UserID); err != nil {
			return err
		}
		// (4) family backup
		if err := daoManager.FamilyBackupDAO.DeleteByUserID(ctx, req.UserID); err != nil {
			return err
		}
		// (5) update backup record
		if err := daoManager.BackupDAO.UpdateFamilyKeyByUserID(ctx, req.UserID, ""); err != nil {
			return err
		}
		if err := daoManager.FamilyMessageDAO.Create(ctx, &FamilyMessage{
			FamilyID:  familyMember.FamilyID,
			CreatedBy: loginInfo.Email,
			Target:    target,
			Operation: dao.FamilyMessageOperationRemoveMember,
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

func (s *FamilyService) LeaveFamily(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.LeaveFamilyRequest) *gerror.AppError {
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
	// 1. get family member
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember != nil && familyMember.IsAdmin != dao.FamilyMemberIsAdmin {
		if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
			daoManager := dao.ManagerWithDB(tx)
			// (1) family member
			if err := daoManager.FamilyMemberDAO.DeleteByFamilyIDAndUserID(ctx, familyMember.FamilyID, loginInfo.UserID); err != nil {
				return err
			}
			// (2) shared records
			if err := daoManager.FamilySharedRecordDAO.DeleteMemberByFamilyIDAndUserID(ctx, familyMember.FamilyID, loginInfo.UserID); err != nil {
				return err
			}
			// (3) family recover
			if err := daoManager.FamilyRecoverDAO.DeleteByUserID(ctx, loginInfo.UserID); err != nil {
				return err
			}
			// (4) family backup
			if err := daoManager.FamilyBackupDAO.DeleteByUserID(ctx, loginInfo.UserID); err != nil {
				return err
			}
			// (5) update backup record
			if err := daoManager.BackupDAO.UpdateFamilyKeyByUserID(ctx, loginInfo.UserID, ""); err != nil {
				return err
			}
			if err := daoManager.FamilyMessageDAO.Create(ctx, &FamilyMessage{
				FamilyID:  familyMember.FamilyID,
				CreatedBy: loginInfo.Email,
				Target:    loginInfo.Email,
				Operation: dao.FamilyMessageOperationLeaveFamily,
			}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return nil
}

func (s *FamilyService) ShareDataToFamily(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.ShareDataToFamilyRequest) *gerror.AppError {
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
	// 1. get family member
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		err = errors.New("the user's family not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	count, err := s.DaoManager.FamilySharedRecordDAO.CountByFamilyID(ctx, familyMember.FamilyID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if count == 100 {
		err = errors.New("the family shared datas' number has reached the limit")
		logger.Info("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceLimit")).WithCode(response.StatusCodeResourceLimit)
	}
	// 3. check whether the data exists
	if req.Type == "password" {
		record, err := s.DaoManager.PasswordRecordDAO.GetByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if record == nil {
			err = fmt.Errorf("the data [id=%s] not exists", req.DataID)
			logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
		}
	} else {
		record, err := s.DaoManager.NonPasswordRecordDAO.GetByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if record == nil {
			err = fmt.Errorf("the data [id=%s] not exists", req.DataID)
			logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
		}
	}
	// 4. share the data
	record, err := s.DaoManager.FamilySharedRecordDAO.GetByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// create
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		sharedRecord := &FamilySharedRecord{
			Content: req.Content,
		}
		// shared record member
		members := make([]int64, 0)
		if len(req.SharingMembers) > 0 {
			sharedRecord.SharedToAll = dao.FamilySharedRecordShared // shared to memebers
			members = req.SharingMembers
		} else {
			sharedRecord.SharedToAll = dao.FamilySharedRecordSharedToAll // shared to all members
			familyMembers, err := s.DaoManager.FamilyMemberDAO.GetByFamilyID(ctx, familyMember.FamilyID)
			if err != nil {
				return err
			}
			for _, item := range familyMembers {
				if item.UserID != loginInfo.UserID {
					members = append(members, item.UserID)
				}
			}
		}
		sharingMembers, err := json.Marshal(members)
		if err != nil {
			return err
		}
		sharedRecord.SharingMembers = string(sharingMembers)
		if record != nil {
			// update
			if err := daoManager.FamilySharedRecordDAO.UpdateByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID, sharedRecord); err != nil {
				return err
			}
		} else {
			// shared record
			sharedRecord.DataID = req.DataID
			sharedRecord.FamilyID = familyMember.FamilyID
			sharedRecord.SharedBy = loginInfo.UserID
			sharedRecord.Type = req.Type
			sharedRecord.Content = req.Content
			sharedRecord.Version = 1
			if err := daoManager.FamilySharedRecordDAO.Create(ctx, sharedRecord); err != nil {
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

func (s *FamilyService) ManageSharingData(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.ManageSharingDataRequest) *gerror.AppError {
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
	// 1. get family member
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		return nil
	}
	// 2. check shared record
	record, err := s.DaoManager.FamilySharedRecordDAO.GetByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if record == nil {
		err = fmt.Errorf("the shared data [id=%s] not exists", req.DataID)
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	if record.SharedBy != loginInfo.UserID {
		err = errors.New("the user is not the data owner")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_Unauthorized")).WithCode(response.StatusCodeUnauthorized)
	}
	// 2. delete shared data
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if req.Stop {
			if err := daoManager.FamilySharedRecordDAO.DeleteByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID); err != nil {
				return err
			}
		} else {
			sharedRecord := &FamilySharedRecord{
				SharedToAll: record.SharedToAll,
			}
			members := make([]int64, 0)
			if len(req.SharingMembers) > 0 {
				sharedRecord.SharedToAll = dao.FamilySharedRecordShared
				members = req.SharingMembers
			} else {
				sharedRecord.SharedToAll = dao.FamilySharedRecordSharedToAll
				familyMembers, err := s.DaoManager.FamilyMemberDAO.GetByFamilyID(ctx, familyMember.FamilyID)
				if err != nil {
					return err
				}
				for _, item := range familyMembers {
					if item.UserID != loginInfo.UserID {
						members = append(members, item.UserID)
					}
				}
			}
			sharedRecord.UpdatedAt = time.Now().Unix()
			sharingMembers, err := json.Marshal(members)
			if err != nil {
				return err
			}
			sharedRecord.SharingMembers = string(sharingMembers)
			if err := daoManager.FamilySharedRecordDAO.UpdateByUserIDAndDataID(ctx, loginInfo.UserID, req.DataID, sharedRecord); err != nil {
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

func (s *FamilyService) HandleAdminAuthority(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.HandleAdminAuthorityRequest) *gerror.AppError {
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
	// 1. get family member
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		err = errors.New("the user's family not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	if familyMember.IsAdmin != dao.FamilyMemberIsAdmin {
		err = errors.New("the user is not an admin")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_Unauthorized")).WithCode(response.StatusCodeUnauthorized)
	}
	// 2. check param userID
	otherFamilyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, req.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if otherFamilyMember == nil {
		err = errors.New("the other user's family not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	if otherFamilyMember.FamilyID != familyMember.FamilyID {
		err = errors.New("the users are not in one family")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_Unauthorized")).WithCode(response.StatusCodeUnauthorized)
	}
	family, err := s.DaoManager.FamilyDAO.GetByFamilyID(ctx, otherFamilyMember.FamilyID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if family.CreatedBy == req.UserID {
		err = errors.New("the other user is the creator of the family")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_Unauthorized")).WithCode(response.StatusCodeUnauthorized)
	}
	var target string
	result, appError := s.FetchSlarkInfo(ctx, rpcCtx, otherFamilyMember.UserID)
	if appError != nil {
		return appError
	} else {
		target = result.Email
	}
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if otherFamilyMember.IsAdmin != dao.FamilyMemberIsAdmin {
			if err := daoManager.FamilyMemberDAO.UpdateAdminByUserID(ctx, req.UserID, dao.FamilyMemberIsAdmin); err != nil {
				return err
			}
			if err := daoManager.FamilyMessageDAO.Create(ctx, &FamilyMessage{
				FamilyID:  familyMember.FamilyID,
				CreatedBy: loginInfo.Email,
				Target:    target,
				Operation: dao.FamilyMessageOperationAuthorizeAdmin,
			}); err != nil {
				return err
			}
		} else {
			if err := daoManager.FamilyMemberDAO.UpdateAdminByUserID(ctx, req.UserID, dao.FamilyMemberIsNotAdmin); err != nil {
				return err
			}
			if err := daoManager.FamilyMessageDAO.Create(ctx, &FamilyMessage{
				FamilyID:  familyMember.FamilyID,
				CreatedBy: loginInfo.Email,
				Target:    target,
				Operation: dao.FamilyMessageOperationUnauthorizeAdmin,
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

func (s *FamilyService) RemoveFamily(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.RemoveFamilyRequest) *gerror.AppError {
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
	// 1. get family
	family, err := s.DaoManager.FamilyDAO.GetByCreator(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if family == nil {
		err = errors.New("the user's family not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	// 2. check family member numbers
	count, err := s.DaoManager.FamilyMemberDAO.CountByFamilyID(ctx, family.FamilyID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if count != 1 {
		err = errors.New("the family members' number does not meet requirements")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_Forbidden")).WithCode(response.StatusCodeForbidden)
	}
	// 3. delete data
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		// (1) delete family
		if err := daoManager.FamilyDAO.Delete(ctx, family.FamilyID); err != nil {
			return err
		}
		// (2) delete family member
		if err := daoManager.FamilyMemberDAO.DeleteByFamilyID(ctx, family.FamilyID); err != nil {
			return err
		}
		// (3) delete family sharing data
		if err := daoManager.FamilySharedRecordDAO.DeleteByFamilyID(ctx, family.FamilyID); err != nil {
			return err
		}
		// (4) family invitation
		if err := daoManager.FamilyInvitationDAO.DeleteByFamilyID(ctx, family.FamilyID); err != nil {
			return err
		}
		// (5) delete encrypted family key
		if err := daoManager.BackupDAO.UpdateFamilyKeyByUserID(ctx, loginInfo.UserID, ""); err != nil {
			return err
		}
		// (6) family messages
		if err := daoManager.FamilyMessageDAO.DeleteByFamilyID(ctx, family.FamilyID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *FamilyService) GetFamilyMessages(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.GetFamilyMessagesRequest) (*pswds_api.GetFamilyMessagesResponse_Data, *gerror.AppError) {
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
	// 1. get from family member table
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		return nil, nil
	}
	// 3. get family messages
	messages, err := s.DaoManager.FamilyMessageDAO.GetByFamilyID(ctx, familyMember.FamilyID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var list []*pswds_api.GetFamilyMessagesResponse_Message
	for _, item := range messages {
		list = append(list, &pswds_api.GetFamilyMessagesResponse_Message{
			Id:        item.ID,
			CreatedAt: item.CreatedAt,
			Creator:   item.CreatedBy,
			Target:    item.Target,
			Operation: int64(item.Operation),
		})
	}
	// 4. delete expired messages
	if err := s.DaoManager.FamilyMessageDAO.DeleteExpiredByFamilyID(ctx, familyMember.FamilyID); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &pswds_api.GetFamilyMessagesResponse_Data{
		List: list,
	}, nil
}

func (s *FamilyService) GetFamilyBackups(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.GetFamilyBackupsRequest) (*pswds_api.GetFamilyBackupsResponse_Data, *gerror.AppError) {
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
	list, err := s.DaoManager.FamilyBackupDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var self []*pswds_api.GetFamilyBackupsResponse_Backup
	var family []*pswds_api.GetFamilyBackupsResponse_Backup
	for _, item := range list {
		if item.CreatedBy == loginInfo.UserID {
			self = append(self, &pswds_api.GetFamilyBackupsResponse_Backup{
				Id:        item.ID,
				UserID:    item.MemberID,
				Email:     item.Member,
				CreatedAt: item.CreatedAt,
			})
		} else {
			var email string
			result, appError := s.FetchSlarkInfo(ctx, rpcCtx, item.CreatedBy)
			if appError != nil {
				return nil, appError
			} else {
				email = result.Email
			}
			family = append(family, &pswds_api.GetFamilyBackupsResponse_Backup{
				Id:        item.ID,
				UserID:    item.CreatedBy,
				Email:     email,
				CreatedAt: item.CreatedAt,
			})
		}
	}
	return &pswds_api.GetFamilyBackupsResponse_Data{
		Self:   self,
		Family: family,
	}, nil
}

func (s *FamilyService) SetFamilyBackup(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.SetFamilyBackupRequest) *gerror.AppError {
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
	// 1. check family
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		err = errors.New("the user's family not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	// 2. create/update backup
	var familyBackups []FamilyBackup
	for _, item := range req.FamilyBackups {
		familyBackups = append(familyBackups, FamilyBackup{
			FamilyID:   familyMember.FamilyID,
			CreatedBy:  loginInfo.UserID,
			MemberID:   item.UserID,
			Member:     item.Email,
			Ciphertext: item.Ciphertext,
		})
	}
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		// delete old backups
		if err := daoManager.FamilyBackupDAO.DeleteByCreatedBy(ctx, loginInfo.UserID); err != nil {
			return err
		}
		// delete old recovers
		if err := daoManager.FamilyRecoverDAO.DeleteByUserID(ctx, loginInfo.UserID); err != nil {
			return err
		}
		// do new operations
		if req.Set {
			// create new backups
			if len(familyBackups) > 0 {
				if err := daoManager.FamilyBackupDAO.Create(ctx, familyBackups); err != nil {
					return err
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

func (s *FamilyService) RequestFamilyRecover(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.RequestFamilyRecoverRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// 1. 先检测邮箱是否注册， 没注册报错
	registration, err := slark.CheckRegistration(ctx, rpcCtx, req.Email)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if registration == nil || registration.Id <= 0 {
		err = fmt.Errorf("the email [%s] has not registered", req.Email)
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_EmailNotRegistered")).WithCode(response.StatusCodeRegisterEmailNotExists)
	}
	// 2. check request limit
	familyRecoverLog, err := s.DaoManager.UnlockPasswordRecoverDAO.GetFamilyRecoverByUserID(ctx, registration.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyRecoverLog != nil {
		if time.Since(time.Unix(familyRecoverLog.CreatedAt, 0)) < time.Hour*24 {
			err = errors.New("the family recover has reached the limit")
			logger.Info("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceLimit")).WithCode(response.StatusCodeResourceLimit)
		} else {
			// delete the log
			if err := s.DaoManager.UnlockPasswordRecoverDAO.DeleteByID(ctx, familyRecoverLog.ID); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	}
	// 2. check family
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, registration.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		err = errors.New("the user's family not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	// 3. check family backup
	familyBackup, err := s.DaoManager.FamilyBackupDAO.GetByCreatedByAndMemberIDWithFamilyID(ctx, familyMember.FamilyID, registration.Id, req.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyBackup == nil {
		err = errors.New("the user's family backup not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	recoverUUID := uuid.NewString()
	// 4. send email
	if err := email.SendEmail_SelfRecover(ctx, req.Email, recoverUUID); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// 5. create recover record
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if err := daoManager.FamilyRecoverDAO.Create(ctx, &FamilyRecover{
			UUID:      recoverUUID,
			BackupID:  familyBackup.ID,
			CreatedBy: registration.Id,
			TargetID:  req.UserID,
			Operation: dao.FamilyRecoverOperationBySelf,
		}); err != nil {
			return err
		}
		if err := daoManager.UnlockPasswordRecoverDAO.Create(ctx, &UnlockPasswordRecover{
			CreatedBy: registration.Id,
			Type:      dao.UnlockPasswordRecoverTypeFamilyRecover,
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

func (s *FamilyService) HelpFamilyRecover(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.HelpFamilyRecoverRequest) *gerror.AppError {
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
	// 1. check request limit
	// 要检查 24小时限流，（一天内只能帮助一个家人，24小时后可以帮助另一个家人）
	// 以及 是否有正在进行的 被动恢复请求
	familyRecoverLog, err := s.DaoManager.UnlockPasswordRecoverDAO.GetHelpFamilyRecoverByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyRecoverLog != nil {
		if time.Since(time.Unix(familyRecoverLog.CreatedAt, 0)) < time.Hour*24 {
			err = errors.New("the family recover has reached the limit")
			logger.Info("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceLimit")).WithCode(response.StatusCodeResourceLimit)
		} else {
			// delete the log
			if err := s.DaoManager.UnlockPasswordRecoverDAO.DeleteByID(ctx, familyRecoverLog.ID); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	}
	// check target recover
	familyRecover, err := s.DaoManager.FamilyRecoverDAO.GetFamilyRecoverByCreator(ctx, loginInfo.UserID, req.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyRecover != nil {
		// 周期内处于正常状态的找回帮助：帮助没有被拒绝，且未超出7天
		if familyRecover.CheckedAt == 0 && time.Since(time.Unix(familyRecover.CreatedAt, 0)) < time.Hour*24*7 {
			err = errors.New("the family recover has reached the limit")
			logger.Info("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceLimit")).WithCode(response.StatusCodeResourceLimit)
		} else {
			// delete the recover
			if err := s.DaoManager.FamilyRecoverDAO.DeleteByID(ctx, familyRecover.ID); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	}
	// 2. check family
	familyMember, err := s.DaoManager.FamilyMemberDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyMember == nil {
		err = errors.New("the user's family not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	// 3. check family backup
	familyBackup, err := s.DaoManager.FamilyBackupDAO.GetByCreatedByAndMemberIDWithFamilyID(ctx, familyMember.FamilyID, req.UserID, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyBackup == nil {
		err = errors.New("the user's family backup not exists")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	recoverUUID := uuid.NewString()
	// 4. send email
	if err := email.SendEmail_FamilyRecover(ctx, req.Email, loginInfo.Email, recoverUUID); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// 4. create recover record
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if err := daoManager.FamilyRecoverDAO.Create(ctx, &FamilyRecover{
			UUID:      recoverUUID,
			BackupID:  familyBackup.ID,
			CreatedBy: loginInfo.UserID,
			TargetID:  req.UserID,
			Operation: dao.FamilyRecoverOperationFromFamily,
		}); err != nil {
			return err
		}
		if err := daoManager.UnlockPasswordRecoverDAO.Create(ctx, &UnlockPasswordRecover{
			CreatedBy: loginInfo.UserID,
			Type:      dao.UnlockPasswordRecoverTypeHelpFamilyRecover,
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

func (s *FamilyService) RejectFamilyRecover(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.RejectFamilyRecoverRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	familyRecover, err := s.DaoManager.FamilyRecoverDAO.GetByUUID(ctx, req.Uuid)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if familyRecover == nil ||
		familyRecover.CheckedAt > 0 ||
		familyRecover.Operation != dao.FamilyRecoverOperationFromFamily {
		return nil
	}
	if err := s.DaoManager.FamilyRecoverDAO.CheckedByID(ctx, familyRecover.ID); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *FamilyService) GetFamilyBackupRecovers(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.GetFamilyBackupRecoversRequest) (*pswds_api.GetFamilyBackupRecoversResponse_Data, *gerror.AppError) {
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
	// 1. delete expired data
	if err := s.DaoManager.FamilyRecoverDAO.DeleteExpiredByUserID(ctx, loginInfo.UserID, req.UserID); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// 2. get list
	list, err := s.DaoManager.FamilyRecoverDAO.GetWithCiphertextByUserID(ctx, loginInfo.UserID, req.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var self []*pswds_api.GetFamilyBackupRecoversResponse_Recover   // 我（被动）帮助某人找回
	var family []*pswds_api.GetFamilyBackupRecoversResponse_Recover // 某人（主动）找我帮忙
	periodString := os.Getenv("FAMILY_RECOVER_PROBATION_PERIOD")
	if periodString == "" {
		err = errors.New("must set env variable for 'FAMILY_RECOVER_PROBATION_PERIOD'")
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	period, err := strconv.Atoi(periodString)
	if err != nil {
		err = errors.New("invalid value for 'FAMILY_RECOVER_PROBATION_PERIOD'")
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for _, item := range list {
		if item.CreatedBy == loginInfo.UserID {
			var email string
			result, appError := s.FetchSlarkInfo(ctx, rpcCtx, item.TargetID)
			if appError != nil {
				return nil, appError
			} else {
				email = result.Email
			}
			one := &pswds_api.GetFamilyBackupRecoversResponse_Recover{
				Id:        item.ID,
				CreatedAt: item.CreatedAt,
				UserID:    item.TargetID,
				Email:     email,
				CheckedAt: item.CheckedAt,
			}
			createdAt := time.Unix(item.CreatedAt, 0)
			if time.Since(createdAt) > time.Hour*24*time.Duration(period) && item.CheckedAt == 0 && time.Since(createdAt) <= time.Hour*24*7 { // 超过3天，未被拒绝，且不超过一周
				one.Ciphertext = item.Ciphertext
			}
			self = append(self, one)
		} else {
			var email string
			result, appError := s.FetchSlarkInfo(ctx, rpcCtx, item.CreatedBy)
			if appError != nil {
				return nil, appError
			} else {
				email = result.Email
			}
			one := &pswds_api.GetFamilyBackupRecoversResponse_Recover{
				Id:        item.ID,
				CreatedAt: item.CreatedAt,
				UserID:    item.CreatedBy,
				Email:     email,
				CheckedAt: item.CheckedAt,
			}
			createdAt := time.Unix(item.CreatedAt, 0)
			if item.CheckedAt > 0 && time.Since(createdAt) <= time.Hour*24*1 { // 被确认，且不超过一天
				one.Ciphertext = item.Ciphertext
			}
			family = append(family, one)
		}
	}
	return &pswds_api.GetFamilyBackupRecoversResponse_Data{
		Self:   self,
		Family: family,
	}, nil
}

func (s *FamilyService) ConfirmFamilyRecover(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.ConfirmFamilyRecoverRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	selfRecover, err := s.DaoManager.FamilyRecoverDAO.GetByUUID(ctx, req.Uuid)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if selfRecover == nil ||
		selfRecover.CheckedAt > 0 ||
		selfRecover.Operation != dao.FamilyRecoverOperationBySelf {
		return nil
	}
	if err := s.DaoManager.FamilyRecoverDAO.CheckedByID(ctx, selfRecover.ID); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}
