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
	"github.com/nextsurfer/pswds_backend/internal/dao"
	. "github.com/nextsurfer/pswds_backend/internal/model"
	"go.uber.org/zap"
)

type SecurityQuestionsService struct {
	*PswdsService
}

func NewSecurityQuestionsService(ctx context.Context, pswdsService *PswdsService) (*SecurityQuestionsService, error) {
	s := &SecurityQuestionsService{
		PswdsService: pswdsService,
	}
	return s, nil
}

type SecurityQuestions struct {
	Question1 string `json:"question1"`
	Question2 string `json:"question2"`
	Question3 string `json:"question3"`
}

func (s *SecurityQuestionsService) UploadSecurityQuestions(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.UploadSecurityQuestionsRequest) *gerror.AppError {
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
	// validate security questions format
	var securityQuestions SecurityQuestions
	if err := json.Unmarshal([]byte(req.SecurityQuestions), &securityQuestions); err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	backup, err := s.DaoManager.BackupDAO.GetByUserID(ctx, loginInfo.UserID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if backup == nil {
		err = errors.New("user has no backup")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoBackup")).WithCode(response.StatusCodeNoBackup)
	}
	// check updatedAt field
	if backup.UpdatedAt > req.UpdatedAt {
		err = errors.New("backend data pull ahead")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_BackendDataPullAhead")).WithCode(response.StatusCodeDataPullAhead)
	}
	if err := s.DaoManager.BackupDAO.UpdateByUserID(ctx, loginInfo.UserID, &Backup{
		UpdatedAt:                   req.UpdatedAt,
		SecurityQuestions:           req.SecurityQuestions,
		SecurityQuestionsCiphertext: req.SecurityQuestionsCiphertext,
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *SecurityQuestionsService) RecoverUnlockPassword(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.RecoverUnlockPasswordRequest) *gerror.AppError {
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
	// 2. check request limit
	securityQuestionsRecover, err := s.DaoManager.UnlockPasswordRecoverDAO.GetSecurityQuestionsRecoverByUserID(ctx, registration.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if securityQuestionsRecover != nil {
		if time.Since(time.Unix(securityQuestionsRecover.CreatedAt, 0)) < time.Hour*24 {
			err = errors.New("the security questions recover has reached the limit")
			logger.Info("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceLimit")).WithCode(response.StatusCodeResourceLimit)
		} else {
			// delete the log
			if err := s.DaoManager.UnlockPasswordRecoverDAO.DeleteByID(ctx, securityQuestionsRecover.ID); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	}
	// 3. 检测邮箱用户是否有 密码数据， 没有报错
	backup, err := s.DaoManager.BackupDAO.GetByUserID(ctx, registration.Id)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if backup == nil {
		err = errors.New("user has no backup")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoBackup")).WithCode(response.StatusCodeNoBackup)
	}
	// 4. 检测是否有 密保问题， 没有 提示用户没有设置密保
	if backup.SecurityQuestions == "" || backup.SecurityQuestionsCiphertext == "" {
		err = errors.New("user has not set security questions")
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_HasNotSetSecurityQuestions")).WithCode(response.StatusCodeNotSetSecurityQuestions)
	}
	// 5. 有密保问题， 则 通过邮件发给用户 密保问题密文
	if err := email.SendEmail_RecoverUnlockPassword(ctx, req.Email, backup.SecurityQuestionsCiphertext); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// 6. 插入找回记录
	if err := s.DaoManager.UnlockPasswordRecoverDAO.Create(ctx, &UnlockPasswordRecover{
		CreatedBy: registration.Id,
		Type:      dao.UnlockPasswordRecoverTypeSecurityQuestions,
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}
