package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
	"github.com/nextsurfer/pswds_backend/internal/common/email"
	"github.com/nextsurfer/pswds_backend/internal/common/slark"
	. "github.com/nextsurfer/pswds_backend/internal/model"
	"go.uber.org/zap"
)

type TrustedContactService struct {
	*PswdsService
}

func NewTrustedContactService(ctx context.Context, pswdsService *PswdsService) (*TrustedContactService, error) {
	s := &TrustedContactService{
		PswdsService: pswdsService,
	}
	return s, nil
}

func (s *TrustedContactService) CreateTrustedContact(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.CreateTrustedContactRequest) *gerror.AppError {
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
	if err := s.DaoManager.TrustedContactDAO.Create(ctx, &TrustedContact{
		UserID:           loginInfo.UserID,
		ContactEmail:     req.ContactEmail,
		BackupCiphertext: req.BackupCiphertext,
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *TrustedContactService) GetTrustedContacts(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.GetTrustedContactsRequest) (*pswds_api.GetTrustedContactsResponse_Data, *gerror.AppError) {
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
	result, err := s.DaoManager.TrustedContactDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var list []*pswds_api.GetTrustedContactsResponse_TrustedContact
	for _, item := range result {
		list = append(list, &pswds_api.GetTrustedContactsResponse_TrustedContact{
			Id:           item.ID,
			ContactEmail: item.ContactEmail,
		})
	}
	return &pswds_api.GetTrustedContactsResponse_Data{
		List: list,
	}, nil
}

func (s *TrustedContactService) DeleteTrustedContact(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.DeleteTrustedContactRequest) *gerror.AppError {
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
	if err := s.DaoManager.TrustedContactDAO.DeleteByID(ctx, loginInfo.UserID, req.Id); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *TrustedContactService) GetBackupCiphertext(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.GetBackupCiphertextRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// 1. 先检测邮箱是否注册， 没注册报错
	var targetID int64
	slarkInfo, appError := s.FetchSlarkInfoByEmail(ctx, rpcCtx, req.Email)
	if appError != nil {
		return appError
	}
	if slarkInfo != nil {
		targetID = slarkInfo.UserID
	}
	if targetID == 0 {
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
		targetID = registration.Id
		slarkInfo := &SlarkInfo{
			UserID:   registration.Id,
			Email:    req.Email,
			Nickname: registration.Nickname,
		}
		data, err := json.Marshal(slarkInfo)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if err := s.RedisClient.Set(ctx, fmt.Sprintf("%s%d", RedisKeyPrefixSlarkInfo, registration.Id), string(data), 20*time.Minute).Err(); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	// 2. 找可信人记录
	result, err := s.DaoManager.TrustedContactDAO.GetByUserIDAndContactEmail(ctx, targetID, req.ContactEmail)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if result == nil {
		err = errors.New("empty trusted contact record")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	// 3. 发送邮件
	if err := email.SendEmail_TrustedContactBackupCiphertext(ctx, req.ContactEmail, result.BackupCiphertext); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}
