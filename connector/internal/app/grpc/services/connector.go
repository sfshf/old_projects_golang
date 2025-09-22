package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	connector_api "github.com/nextsurfer/connector/api"
	"github.com/nextsurfer/connector/api/response"
	"github.com/nextsurfer/connector/internal/pkg/dao"
	"github.com/nextsurfer/connector/internal/pkg/keystore"
	. "github.com/nextsurfer/connector/internal/pkg/model"
	"github.com/nextsurfer/connector/internal/pkg/redis"
	"github.com/nextsurfer/connector/internal/pkg/util"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

type ConnectorService struct {
	env         gutil.APPEnvType
	logger      *zap.Logger
	daoManager  *dao.Manager
	redisOption *redis.Option
}

func NewConnectorService(env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option) *ConnectorService {
	s := &ConnectorService{
		env:         env,
		logger:      logger,
		daoManager:  daoManager,
		redisOption: redisOption,
	}
	return s
}

func (s *ConnectorService) Info(rpcCtx *rpc.Context, message, method string, duration int64, apikey, keyID, dataID string) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.Int64("duration", duration),
	}
	if apikey != "" {
		fields = append(fields, zap.String("apikey", apikey))
	}
	if keyID != "" {
		fields = append(fields, zap.String("keyID", keyID))
	}
	if dataID != "" {
		fields = append(fields, zap.String("dataID", dataID))
	}
	rpcCtx.Logger.Info(message, fields...)
}

func (s *ConnectorService) Error(rpcCtx *rpc.Context, message, method string, duration int64, err error, apikey, keyID, dataID string) {
	fields := []zapcore.Field{
		zap.String("method", method),
		zap.Int64("duration", duration),
		zap.NamedError("appError", err),
	}
	if apikey != "" {
		fields = append(fields, zap.String("apikey", apikey))
	}
	if keyID != "" {
		fields = append(fields, zap.String("keyID", keyID))
	}
	if dataID != "" {
		fields = append(fields, zap.String("dataID", dataID))
	}
	rpcCtx.Logger.Error(message, fields...)
}

func (s *ConnectorService) CheckPermission(ctx context.Context, rpcCtx *rpc.Context, method, perm string, apiKey string, startTS time.Time, keyID, dataID string) (util.AppKey, *gerror.AppError) {
	appKey, valid := util.CheckPerm(apiKey, perm)
	if !valid {
		err := fmt.Errorf("invalid api key: %s", apiKey)
		s.Error(rpcCtx, "bad request", method, time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
		return appKey, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyIsInvalid")).WithCode(response.StatusCodeInvalidApiKey)
	}
	// return nil, if is admin
	if util.IsAdminApp(appKey) {
		return appKey, nil
	}
	// check data
	if keyID != "" && dataID != "" {
		relationAppDatum, err := s.daoManager.RelationAppDatumDAO.GetByAppWithKeyIDAndDataID(ctx, appKey.App, keyID, dataID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "DeleteData", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return appKey, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if relationAppDatum == nil {
			s.Error(rpcCtx, "bad request", "DeleteData", time.Since(startTS).Milliseconds(), err, apiKey, keyID, dataID)
			return appKey, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UnmatchedPermission")).WithCode(response.StatusCodeUnmatchedPermission)
		}
	}

	// check key
	if keyID != "" {
		relationAppKey, err := s.daoManager.RelationAppKeyDAO.GetByAppWithKeyID(ctx, appKey.App, keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", method, time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return appKey, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		if relationAppKey == nil {
			s.Error(rpcCtx, "bad request", method, time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return appKey, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UnmatchedPermission")).WithCode(response.StatusCodeUnmatchedPermission)
		}
	}
	// only check apikey
	return appKey, nil
}

func (s *ConnectorService) CreatePassword(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, keyID string) (*connector_api.CreatePasswordResponse_Data, *gerror.AppError) {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, "CreatePassword", util.PermWrite, apiKey, startTS, "", "")
	if appError != nil {
		return nil, appError
	}

	var err error
	var password string
	if err = s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		// first, create a relation_app_key record, if not exists
		relationAppKey, err := daoManager.RelationAppKeyDAO.GetByKeyID(ctx, keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "CreatePassword", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
			return err
		}
		if relationAppKey != nil {
			err = errors.New("keyID has existed")
			s.Error(rpcCtx, "bad request", "CreatePassword", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
			return err
		}
		relationAppKey = &RelationAppKey{
			App:   appKey.App,
			KeyID: keyID,
		}
		if err = daoManager.RelationAppKeyDAO.Create(ctx, relationAppKey); err != nil {
			return err
		}
		// second, post to keystore
		respData, err := keystore.CreatePassword(keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "CreatePassword", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
			return err
		}
		password = respData.Data.Password
		// third, update the relation_app_key record
		passwordHash, err := util.Keccak256Hex([]byte(respData.Data.Password))
		if err != nil {
			s.Error(rpcCtx, "internal error", "CreatePassword", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
			return err
		}
		if err = daoManager.RelationAppKeyDAO.UpdatePasswordHashByID(ctx, relationAppKey.ID, string(passwordHash)); err != nil {
			s.Error(rpcCtx, "internal error", "CreatePassword", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
			return err
		}
		return nil
	}); err != nil {
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &connector_api.CreatePasswordResponse_Data{
		Password: password,
	}, nil
}

func (s *ConnectorService) CheckPassword(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, keyID, passwordHash string) (*connector_api.CheckPasswordResponse_Data, *gerror.AppError) {
	// check permission
	_, appError := s.CheckPermission(ctx, rpcCtx, "CheckPassword", util.PermRead, apiKey, startTS, keyID, "")
	if appError != nil {
		return nil, appError
	}
	respData, err := keystore.CheckPassword(keyID, passwordHash)
	if err != nil {
		s.Error(rpcCtx, "internal error", "CheckPassword", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "CheckPassword", time.Since(startTS).Milliseconds(), apiKey, keyID, "")
	return &connector_api.CheckPasswordResponse_Data{
		Valid: respData.Data.Valid,
	}, nil
}

func (s *ConnectorService) DeletePassword(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, keyID string) *gerror.AppError {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, "DeletePassword", util.PermWrite, apiKey, startTS, keyID, "")
	if appError != nil {
		return appError
	}
	var err error
	if err = s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		// first, delete the relation_app_key record, if exists
		if err := daoManager.RelationAppKeyDAO.DeleteByAppWithKeyID(ctx, appKey.App, keyID); err != nil {
			s.Error(rpcCtx, "internal error", "DeletePassword", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		// second, post to keystore
		if _, err := keystore.DeletePassword(keyID); err != nil {
			s.Error(rpcCtx, "internal error", "DeletePassword", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		return nil
	}); err != nil {
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "DeletePassword", time.Since(startTS).Milliseconds(), apiKey, keyID, "")
	return nil
}

func (s *ConnectorService) CreatePrivateKey(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, keyID string) (*connector_api.CreatePrivateKeyResponse_Data, *gerror.AppError) {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, "CreatePrivateKey", util.PermWrite, apiKey, startTS, "", "")
	if appError != nil {
		return nil, appError
	}
	var err error
	var publicKey string
	if err = s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		// first, create a relation_app_key record, if not exists
		relationAppKey, err := daoManager.RelationAppKeyDAO.GetByKeyID(ctx, keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "CreatePrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		if relationAppKey != nil {
			err = errors.New("keyID has existed")
			s.Error(rpcCtx, "bad request", "CreatePrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		relationAppKey = &RelationAppKey{
			App:   appKey.App,
			KeyID: keyID,
		}
		if err = daoManager.RelationAppKeyDAO.Create(ctx, relationAppKey); err != nil {
			s.Error(rpcCtx, "internal error", "CreatePrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		// second, post to keystore
		respData, err := keystore.CreatePrivateKey(keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "CreatePrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		publicKey = respData.Data.PublicKey
		// third, update the relation_app_key record
		if err = daoManager.RelationAppKeyDAO.UpdatePasswordHashByID(ctx, relationAppKey.ID, publicKey); err != nil {
			s.Error(rpcCtx, "internal error", "CreatePrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		return nil
	}); err != nil {
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "CreatePrivateKey", time.Since(startTS).Milliseconds(), apiKey, keyID, "")
	return &connector_api.CreatePrivateKeyResponse_Data{
		PublicKey: publicKey,
	}, nil
}

func (s *ConnectorService) GetPublicKey(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, keyID string) (*connector_api.GetPublicKeyResponse_Data, *gerror.AppError) {
	// check permission
	_, appError := s.CheckPermission(ctx, rpcCtx, "GetPublicKey", util.PermRead, apiKey, startTS, keyID, "")
	if appError != nil {
		return nil, appError
	}
	respData, err := keystore.GetPublicKey(keyID)
	if err != nil {
		s.Error(rpcCtx, "internal error", "GetPublicKey", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "GetPublicKey", time.Since(startTS).Milliseconds(), apiKey, keyID, "")
	return &connector_api.GetPublicKeyResponse_Data{
		PublicKey: respData.Data.PublicKey,
	}, nil
}

func (s *ConnectorService) DeletePrivateKey(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, keyID string) *gerror.AppError {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, "GetPublicKey", util.PermWrite, apiKey, startTS, keyID, "")
	if appError != nil {
		return appError
	}
	var err error
	if err = s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		// first, delete the relation_app_key record, if exists
		if err := daoManager.RelationAppKeyDAO.DeleteByAppWithKeyID(ctx, appKey.App, keyID); err != nil {
			s.Error(rpcCtx, "internal error", "DeletePrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		// second, post to keystore
		if _, err := keystore.DeletePrivateKey(keyID); err != nil {
			s.Error(rpcCtx, "internal error", "DeletePrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		return nil
	}); err != nil {
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "DeletePrivateKey", time.Since(startTS).Milliseconds(), apiKey, keyID, "")
	return nil
}

func (s *ConnectorService) SaveData(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, keyID, dataID string, replaceCurrentItem bool, data string, plaintextHash string) *gerror.AppError {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, "SaveData", util.PermWrite, apiKey, startTS, keyID, "")
	if appError != nil {
		return appError
	}
	if err := s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		// first, create a relation_app_data record, if not exists
		relationAppDatum, err := daoManager.RelationAppDatumDAO.GetByAppWithKeyIDAndDataID(ctx, appKey.App, keyID, dataID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "SaveData", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		if relationAppDatum == nil {
			relationAppDatum = &RelationAppDatum{
				App:    appKey.App,
				KeyID:  keyID,
				DataID: dataID,
			}
			if err = daoManager.RelationAppDatumDAO.Create(ctx, relationAppDatum); err != nil {
				s.Error(rpcCtx, "internal error", "SaveData", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
				return err
			}
		}
		// second, post to keystore
		if _, err = keystore.SaveData(keyID, dataID, data, plaintextHash, replaceCurrentItem); err != nil {
			s.Error(rpcCtx, "internal error", "SaveData", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		return nil
	}); err != nil {
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "SaveData", time.Since(startTS).Milliseconds(), apiKey, keyID, dataID)
	return nil
}

func (s *ConnectorService) DeleteData(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, keyID, dataID string) *gerror.AppError {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, "DeleteData", util.PermWrite, apiKey, startTS, keyID, dataID)
	if appError != nil {
		return appError
	}
	var err error
	tx, daoManager := s.daoManager.Transaction()
	defer func() {
		if err != nil {
			if e := tx.Rollback().Error; e != nil {
				s.Error(rpcCtx, "internal error", "DeleteData", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			}
		}
	}()
	// first, delete the relation_app_data record, if exists
	if err = daoManager.RelationAppDatumDAO.DeleteByAppWithKeyIDAndDataID(ctx, appKey.App, keyID, dataID); err != nil {
		s.Error(rpcCtx, "internal error", "DeleteData", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// second, post to keystore
	_, err = keystore.DeleteData(keyID, dataID)
	if err != nil {
		s.Error(rpcCtx, "internal error", "DeleteData", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// last, commit the transaction
	if err = tx.Commit().Error; err != nil {
		s.Error(rpcCtx, "internal error", "DeleteData", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "DeleteData", time.Since(startTS).Milliseconds(), apiKey, keyID, dataID)
	return nil
}

func (s *ConnectorService) DecryptData(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, keyID, data string) (*connector_api.DecryptDataResponse_Data, *gerror.AppError) {
	// check permission
	_, appError := s.CheckPermission(ctx, rpcCtx, "DecryptData", util.PermWrite, apiKey, startTS, keyID, "")
	if appError != nil {
		return nil, appError
	}
	respData, err := keystore.DecryptData(keyID, data)
	if err != nil {
		s.Error(rpcCtx, "internal error", "DecryptData", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "DecryptData", time.Since(startTS).Milliseconds(), apiKey, keyID, "")
	return &connector_api.DecryptDataResponse_Data{
		DecrypedData: respData.Data.DecrypedData,
	}, nil
}

func (s *ConnectorService) GetData(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, keyID, dataID string) (*connector_api.GetDataResponse_Data, *gerror.AppError) {
	// check permission
	_, appError := s.CheckPermission(ctx, rpcCtx, "GetData", util.PermWrite, apiKey, startTS, keyID, dataID)
	if appError != nil {
		return nil, appError
	}
	respData, err := keystore.GetData(keyID, dataID)
	if err != nil {
		s.Error(rpcCtx, "internal error", "GetData", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "GetData", time.Since(startTS).Milliseconds(), apiKey, keyID, dataID)
	return &connector_api.GetDataResponse_Data{
		Data: respData.Data.Data,
	}, nil
}

func (s *ConnectorService) CheckKeyExisting(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, keyID string) (*connector_api.CheckKeyExistingResponse_Data, *gerror.AppError) {
	// check permission
	_, appError := s.CheckPermission(ctx, rpcCtx, "CheckKeyExisting", util.PermRead, apiKey, startTS, keyID, "")
	if appError != nil {
		return nil, appError
	}
	respData, err := keystore.CheckKeyExisting(keyID)
	if err != nil {
		s.Error(rpcCtx, "internal error", "CheckKeyExisting", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "CheckKeyExisting", time.Since(startTS).Milliseconds(), apiKey, keyID, "")
	return &connector_api.CheckKeyExistingResponse_Data{
		Existing: respData.Data.Existing,
	}, nil
}
