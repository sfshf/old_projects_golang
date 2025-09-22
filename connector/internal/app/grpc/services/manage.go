package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

type ConnectorConsoleService struct {
	env         gutil.APPEnvType
	logger      *zap.Logger
	daoManager  *dao.Manager
	redisOption *redis.Option
}

func NewConnectorConsoleService(env gutil.APPEnvType, logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option) *ConnectorConsoleService {
	s := &ConnectorConsoleService{
		env:         env,
		logger:      logger,
		daoManager:  daoManager,
		redisOption: redisOption,
	}
	return s
}

func (s *ConnectorConsoleService) Info(rpcCtx *rpc.Context, message, method string, duration int64, apikey, app, keyID string) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.Int64("duration", duration),
	}
	if apikey != "" {
		fields = append(fields, zap.String("apikey", apikey))
	}
	if app != "" {
		fields = append(fields, zap.String("app", app))
	}
	if keyID != "" {
		fields = append(fields, zap.String("keyID", keyID))
	}
	rpcCtx.Logger.Info(message, fields...)
}

func (s *ConnectorConsoleService) Error(rpcCtx *rpc.Context, message, method string, duration int64, err error, apikey, app, keyID string) {
	fields := []zapcore.Field{
		zap.String("method", method),
		zap.Int64("duration", duration),
		zap.NamedError("appError", err),
	}
	if apikey != "" {
		fields = append(fields, zap.String("apikey", apikey))
	}
	if app != "" {
		fields = append(fields, zap.String("app", app))
	}
	if keyID != "" {
		fields = append(fields, zap.String("keyID", keyID))
	}
	rpcCtx.Logger.Error(message, fields...)
}

func (s *ConnectorConsoleService) CheckPermission(ctx context.Context, rpcCtx *rpc.Context, mustAdmin bool, method, perm string, apiKey string, startTS time.Time, app, keyID string) (util.AppKey, *gerror.AppError) {
	appKey, valid := util.CheckPerm(apiKey, perm)
	if !valid {
		err := fmt.Errorf("invalid api key: %s", apiKey)
		s.Error(rpcCtx, "bad request", method, time.Since(startTS).Milliseconds(), err, apiKey, app, keyID)
		return appKey, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyIsInvalid")).WithCode(response.StatusCodeInvalidApiKey)
	}
	// return nil, if is admin
	if util.IsAdminApp(appKey) {
		return appKey, nil
	} else if mustAdmin {
		err := errors.New("must be admin account")
		return appKey, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UnmatchedPermission")).WithCode(response.StatusCodeUnmatchedPermission)
	}
	// check app
	if app != "" && appKey.App != app {
		err := fmt.Errorf("app key's app name [%s] not equal to app parameter [%s]", appKey.App, app)
		s.Error(rpcCtx, "bad request", method, time.Since(startTS).Milliseconds(), err, apiKey, app, keyID)
		return appKey, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UnmatchedPermission")).WithCode(response.StatusCodeUnmatchedPermission)
	}
	// only check apikey
	return appKey, nil
}

func (s *ConnectorConsoleService) PlatformLog(ctx context.Context, rpcCtx *rpc.Context, method string, appKey util.AppKey, startTS time.Time, err error, app, keyID, objJson string) {
	status := "fail"
	if err == nil {
		status = "success"
	}
	if e := s.daoManager.ManagePlatformLogDAO.Create(ctx, &ManagePlatformLog{
		IP:         rpcCtx.IP,
		APIKeyName: appKey.Name,
		Status:     status,
		Object:     objJson,
		Operation:  method,
		App:        app,
		KeyID:      keyID,
	}); e != nil {
		s.Error(rpcCtx, "internal error", method, time.Since(startTS).Milliseconds(), err, appKey.ApiKey, "", "")
	}
}

func (s *ConnectorConsoleService) GetAllApps(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey string) (*connector_api.GetAllAppsResponse_Data, *gerror.AppError) {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "GetAllApps", util.PermRead, apiKey, startTS, "", "")
	if appError != nil {
		return nil, appError
	}
	// filter by app, if not admin app
	allApps, err := s.daoManager.AppConfigDAO.GetAllApps(ctx)
	if err != nil {
		s.Error(rpcCtx, "internal error", "GetAllApps", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var res []string
	if util.IsAdminApp(appKey) {
		res = allApps
	} else {
		for _, app := range allApps {
			if app == appKey.App {
				res = append(res, app)
			}
		}
	}
	return &connector_api.GetAllAppsResponse_Data{
		List: res,
	}, nil
}

func (s *ConnectorConsoleService) GetAllKeyIDs(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey string) (*connector_api.GetAllKeyIDsResponse_Data, *gerror.AppError) {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "GetAllKeyIDs", util.PermRead, apiKey, startTS, "", "")
	if appError != nil {
		return nil, appError
	}
	// filter by app, if not admin app
	allKeyIDs, err := s.daoManager.RelationAppKeyDAO.GetAllKeyIDs(ctx)
	if err != nil {
		s.Error(rpcCtx, "internal error", "GetAllKeyIDs", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var res []string
	if util.IsAdminApp(appKey) {
		res = allKeyIDs
	} else {
		for _, keyID := range allKeyIDs {
			if keyID == appKey.KeyID {
				res = append(res, keyID)
			}
		}
	}
	return &connector_api.GetAllKeyIDsResponse_Data{
		List: res,
	}, nil
}

func (s *ConnectorConsoleService) ValidateApiKey(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, app, apiKey, role string) (*connector_api.ValidateApiKeyResponse_Data, *gerror.AppError) {
	valid := false
	relationAppKey, err := s.daoManager.RelationAppKeyDAO.GetByAppWithPasswordHash(ctx, app, apiKey)
	if err != nil {
		s.Error(rpcCtx, "internal error", "ValidateApiKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if relationAppKey == nil {
		err = fmt.Errorf("relation app key [%s] record not exists", apiKey)
		s.Error(rpcCtx, "bad request", "ValidateApiKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyIsInvalid")).WithCode(response.StatusCodeInvalidApiKey)
	}
	apiKeyObj, err := s.daoManager.ApiKeyDAO.GetByAppWithKeyID(ctx, app, relationAppKey.KeyID)
	if err != nil {
		s.Error(rpcCtx, "internal error", "ValidateApiKey", time.Since(startTS).Milliseconds(), err, apiKey, app, relationAppKey.KeyID)
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if apiKeyObj == nil {
		err = fmt.Errorf("api key [%s] record not exists", apiKey)
		s.Error(rpcCtx, "bad request", "ValidateApiKey", time.Since(startTS).Milliseconds(), err, apiKey, app, relationAppKey.KeyID)
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyIsInvalid")).WithCode(response.StatusCodeInvalidApiKey)
	}
	valid = apiKeyObj.Permission == role || apiKeyObj.Permission == util.PermWrite
	if !valid {
		s.Info(rpcCtx, "ApiKey is invalid", "ValidateApiKey", time.Since(startTS).Milliseconds(), apiKey, app, relationAppKey.KeyID)
	}
	return &connector_api.ValidateApiKeyResponse_Data{
		Valid: valid,
	}, nil
}

func (s *ConnectorConsoleService) ListApiKey(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey string) (*connector_api.ListApiKeyResponse_Data, *gerror.AppError) {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "ListApiKey", util.PermRead, apiKey, startTS, "", "")
	if appError != nil {
		return nil, appError
	}
	// filter by app, if not admin app
	var err error
	var apiKeys []*APIKey
	if util.IsAdminApp(appKey) {
		apiKeys, err = s.daoManager.ApiKeyDAO.GetAll(ctx)
		if err != nil {
			s.Error(rpcCtx, "internal error", "ListApiKey", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	} else {
		apiKeys, err = s.daoManager.ApiKeyDAO.GetListByApp(ctx, appKey.App)
		if err != nil {
			s.Error(rpcCtx, "internal error", "ListApiKey", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	var list []*connector_api.ListApiKeyResponse_ListItem
	for _, item := range apiKeys {
		list = append(list, &connector_api.ListApiKeyResponse_ListItem{
			Id:         item.ID,
			CreatedAt:  item.CreatedAt.UnixMilli(),
			Permission: item.Permission,
			KeyID:      item.KeyID,
			App:        item.App,
			Name:       item.Name,
		})
	}

	return &connector_api.ListApiKeyResponse_Data{List: list}, nil
}

func (s *ConnectorConsoleService) AddApiKey(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, app, keyID, name, permission string) *gerror.AppError {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "AddApiKey", util.PermWrite, apiKey, startTS, app, "")
	if appError != nil {
		return appError
	}
	var err error
	// manage platform log
	defer func() {
		s.PlatformLog(ctx, rpcCtx, "AddApiKey", appKey, startTS, err, app, keyID, fmt.Sprintf(`{"app": "%s", "keyID": "%s", "name": "%s", "permission": "%s"}`, app, keyID, name, permission))
	}()
	// first, check keyID whether exists
	relationAppKey, err := s.daoManager.RelationAppKeyDAO.GetByKeyID(ctx, keyID)
	if err != nil {
		s.Error(rpcCtx, "internal error", "AddApiKey", time.Since(startTS).Milliseconds(), err, apiKey, app, keyID)
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if relationAppKey == nil {
		s.Error(rpcCtx, "bad request", "AddApiKey", time.Since(startTS).Milliseconds(), err, apiKey, app, keyID)
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_KeyIDNotExists")).WithCode(response.StatusCodeKeyIDNotExists)
	}
	// second, create a api_key record, if not exists
	apiKeyObj, err := s.daoManager.ApiKeyDAO.GetByName(ctx, name)
	if err != nil {
		s.Error(rpcCtx, "internal error", "AddApiKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if apiKeyObj != nil {
		s.Error(rpcCtx, "bad request", "AddApiKey", time.Since(startTS).Milliseconds(), err, apiKey, app, keyID)
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_ApiKeyNameExists")).WithCode(response.StatusCodeApiKeyNameExists)
	}
	if err = s.daoManager.ApiKeyDAO.Create(ctx, &APIKey{
		App:        app,
		KeyID:      keyID,
		Name:       name,
		Permission: permission,
	}); err != nil {
		s.Error(rpcCtx, "internal error", "AddApiKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// last, add the app key
	util.AddAppKey(app, keyID, name, permission, relationAppKey.PasswordHash)
	s.Info(rpcCtx, "success", "AddApiKey", time.Since(startTS).Milliseconds(), apiKey, app, keyID)
	return nil
}

func (s *ConnectorConsoleService) RemoveApiKey(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey string, id int64, apiKeyObj *APIKey) *gerror.AppError {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "RemoveApiKey", util.PermWrite, apiKey, startTS, apiKeyObj.App, "")
	if appError != nil {
		return appError
	}
	var err error
	// manage platform log
	defer func() {
		s.PlatformLog(ctx, rpcCtx, "RemoveApiKey", appKey, startTS, err, apiKeyObj.App, apiKeyObj.KeyID, fmt.Sprintf(`{"id": "%d"}`, id))
	}()
	// first, delete api_key record
	if err := s.daoManager.ApiKeyDAO.DeleteByID(ctx, id); err != nil {
		s.Error(rpcCtx, "internal error", "RemoveApiKey", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// last, remove the app key
	util.RemoveAppKey(apiKeyObj.Name)
	s.Info(rpcCtx, "success", "RemoveApiKey", time.Since(startTS).Milliseconds(), apiKey, "", "")
	return nil
}

func (s *ConnectorConsoleService) ListConfigs(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey string) (*connector_api.ListConfigsResponse_Data, *gerror.AppError) {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "ListConfigs", util.PermRead, apiKey, startTS, "", "")
	if appError != nil {
		return nil, appError
	}
	var err error
	var appConfigs []*AppConfig
	// filter by app, if not admin app
	if util.IsAdminApp(appKey) {
		appConfigs, err = s.daoManager.AppConfigDAO.GetAll(ctx)
		if err != nil {
			s.Error(rpcCtx, "internal error", "ListConfigs", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	} else {
		appConfigs, err = s.daoManager.AppConfigDAO.GetListByApp(ctx, appKey.App)
		if err != nil {
			s.Error(rpcCtx, "internal error", "ListConfigs", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	var list []*connector_api.ListConfigsResponse_AppConfig
	for _, appConfig := range appConfigs {
		list = append(list, &connector_api.ListConfigsResponse_AppConfig{
			App:    appConfig.App,
			Config: appConfig.Config,
		})
	}

	return &connector_api.ListConfigsResponse_Data{ConfigList: list}, nil
}

func (s *ConnectorConsoleService) CreateConfig(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, app, config string) *gerror.AppError {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, true, "CreateConfig", util.PermWrite, apiKey, startTS, app, "")
	if appError != nil {
		return appError
	}
	var err error
	// manage platform log
	defer func() {
		s.PlatformLog(ctx, rpcCtx, "CreateConfig", appKey, startTS, err, app, "", fmt.Sprintf(`{"app": "%s", "config": "%s"}`, app, config))
	}()
	if err = s.daoManager.AppConfigDAO.Create(ctx, &AppConfig{
		App:    app,
		Config: config,
	}); err != nil {
		s.Error(rpcCtx, "internal error", "CreateConfig", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *ConnectorConsoleService) UpdateConfig(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, app, config string) *gerror.AppError {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "UpdateConfig", util.PermWrite, apiKey, startTS, app, "")
	if appError != nil {
		return appError
	}
	var err error
	// manage platform log
	defer func() {
		s.PlatformLog(ctx, rpcCtx, "UpdateConfig", appKey, startTS, err, app, "", fmt.Sprintf(`{"app": "%s", "config": "%s"}`, app, config))
	}()
	if err = s.daoManager.AppConfigDAO.UpdateByApp(ctx, app, config); err != nil {
		s.Error(rpcCtx, "internal error", "UpdateConfig", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *ConnectorConsoleService) DeleteConfig(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, app string) *gerror.AppError {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, true, "DeleteConfig", util.PermWrite, apiKey, startTS, app, "")
	if appError != nil {
		return appError
	}
	var err error
	// manage platform log
	defer func() {
		s.PlatformLog(ctx, rpcCtx, "DeleteConfig", appKey, startTS, err, app, "", fmt.Sprintf(`{"app": "%s"}`, app))
	}()
	if err = s.daoManager.AppConfigDAO.DeleteByApp(ctx, app); err != nil {
		s.Error(rpcCtx, "internal error", "DeleteConfig", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *ConnectorConsoleService) ListPassword(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, app string) (*connector_api.ListPasswordResponse_Data, *gerror.AppError) {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "ListPassword", util.PermRead, apiKey, startTS, "", "")
	if appError != nil {
		return nil, appError
	}
	var err error
	var relationAppKeys []*RelationAppKey
	if !util.IsAdminApp(appKey) {
		if app = strings.TrimSpace(app); app == "" {
			app = appKey.App
		}
		if appKey.App != app {
			err := fmt.Errorf("invalid api key: %s", apiKey)
			s.Error(rpcCtx, "bad request", "ListPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_NoWritePermission")).WithCode(response.StatusCodeUnmatchedPermission)
		}
		relationAppKeys, err = s.daoManager.RelationAppKeyDAO.GetListByApp(ctx, app)
		if err != nil {
			s.Error(rpcCtx, "internal error", "ListPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	} else {
		relationAppKeys, err = s.daoManager.RelationAppKeyDAO.GetListWithApp(ctx, app)
		if err != nil {
			s.Error(rpcCtx, "internal error", "ListPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	var list []*connector_api.ListPasswordResponse_ListItem
	for _, relation := range relationAppKeys {
		if strings.HasPrefix(relation.KeyID, "pswd_") {
			list = append(list, &connector_api.ListPasswordResponse_ListItem{
				App:          relation.App,
				KeyID:        relation.KeyID,
				PasswordHash: relation.PasswordHash,
				CreatedAt:    relation.CreatedAt.UnixMilli(),
			})
		}
	}

	return &connector_api.ListPasswordResponse_Data{List: list}, nil
}

func (s *ConnectorConsoleService) AddPassword(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, app, keyID string) (*connector_api.AddPasswordResponse_Data, *gerror.AppError) {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "AddPassword", util.PermWrite, apiKey, startTS, app, "")
	if appError != nil {
		return nil, appError
	}
	var err error
	// manage platform log
	defer func() {
		s.PlatformLog(ctx, rpcCtx, "AddPassword", appKey, startTS, err, app, keyID, fmt.Sprintf(`{"app": "%s", "keyID": "%s"}`, app, keyID))
	}()
	var relationAppKey *RelationAppKey
	var passwordHash []byte
	if err = s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		// first, create a relation_app_key record, if not exists
		relationAppKey, err = daoManager.RelationAppKeyDAO.GetByKeyID(ctx, keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "AddPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return err
		}
		if relationAppKey != nil {
			err = errors.New("keyID has existed")
			s.Error(rpcCtx, "bad request", "AddPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return err
		}
		relationAppKey = &RelationAppKey{
			App:   app,
			KeyID: keyID,
		}
		// check key existing
		respData, err := keystore.CheckKeyExisting(keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "CheckKeyExisting", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		if !respData.Data.Existing {
			if err = daoManager.RelationAppKeyDAO.Create(ctx, relationAppKey); err != nil {
				s.Error(rpcCtx, "internal error", "AddPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
			// second, post to keystore
			respData, err := keystore.CreatePassword(keyID)
			if err != nil {
				s.Error(rpcCtx, "internal error", "AddPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
			// third, update the relation_app_key record
			passwordHash, err = util.Keccak256Hex([]byte(respData.Data.Password))
			if err != nil {
				s.Error(rpcCtx, "internal error", "AddPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
			if err := daoManager.RelationAppKeyDAO.UpdatePasswordHashByID(ctx, relationAppKey.ID, string(passwordHash)); err != nil {
				s.Error(rpcCtx, "internal error", "AddPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
		} else {
			respData, err := keystore.GetPassword(keyID)
			if err != nil {
				s.Error(rpcCtx, "internal error", "AddPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
			passwordHash, err := util.Keccak256Hex([]byte(respData.Data.Password))
			if err != nil {
				s.Error(rpcCtx, "internal error", "AddPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
			relationAppKey.PasswordHash = string(passwordHash)
			if err = daoManager.RelationAppKeyDAO.Create(ctx, relationAppKey); err != nil {
				s.Error(rpcCtx, "internal error", "AddPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "AddPassword", time.Since(startTS).Milliseconds(), apiKey, app, keyID)
	return &connector_api.AddPasswordResponse_Data{
		KeyID:        keyID,
		PasswordHash: string(passwordHash),
		CreatedAt:    relationAppKey.CreatedAt.UnixMilli(),
	}, nil
}

func (s *ConnectorConsoleService) RemovePassword(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, app, keyID string) *gerror.AppError {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "RemovePassword", util.PermWrite, apiKey, startTS, app, "")
	if appError != nil {
		return appError
	}
	var err error
	// manage platform log
	defer func() {
		s.PlatformLog(ctx, rpcCtx, "RemovePassword", appKey, startTS, err, app, keyID, fmt.Sprintf(`{"app": "%s", "keyID": "%s"}`, app, keyID))
	}()
	if err = s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		// first, delete the relation_app_key record, if exists
		err := daoManager.RelationAppKeyDAO.DeleteByAppWithKeyID(ctx, app, keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "RemovePassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return err
		}
		// second, post to keystore
		_, err = keystore.DeletePassword(keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "RemovePassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return err
		}
		return nil
	}); err != nil {
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "RemovePassword", time.Since(startTS).Milliseconds(), apiKey, app, keyID)
	return nil
}

func (s *ConnectorConsoleService) FetchPassword(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, app, keyID string) (*connector_api.FetchPasswordResponse_Data, *gerror.AppError) {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "FetchPassword", util.PermWrite, apiKey, startTS, app, "")
	if appError != nil {
		return nil, appError
	}
	var err error
	// manage platform log
	defer func() {
		s.PlatformLog(ctx, rpcCtx, "FetchPassword", appKey, startTS, err, app, keyID, fmt.Sprintf(`{"app": "%s", "keyID": "%s"}`, app, keyID))
	}()
	respData, err := keystore.GetPassword(keyID)
	if err != nil {
		s.Error(rpcCtx, "internal error", "FetchPassword", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &connector_api.FetchPasswordResponse_Data{
		Password: respData.Data.Password,
	}, nil
}

func (s *ConnectorConsoleService) ListPrivateKey(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, app string) (*connector_api.ListPrivateKeyResponse_Data, *gerror.AppError) {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "FetchPassword", util.PermRead, apiKey, startTS, "", "")
	if appError != nil {
		return nil, appError
	}
	var err error
	var relationAppKeys []*RelationAppKey
	if !util.IsAdminApp(appKey) {
		if app = strings.TrimSpace(app); app == "" {
			app = appKey.App
		}
		if appKey.App != app {
			err := fmt.Errorf("invalid api key: %s", apiKey)
			s.Error(rpcCtx, "bad request", "ListPrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_NoWritePermission")).WithCode(response.StatusCodeUnmatchedPermission)
		}
		relationAppKeys, err = s.daoManager.RelationAppKeyDAO.GetListByApp(ctx, app)
		if err != nil {
			s.Error(rpcCtx, "internal error", "ListPrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	} else {
		relationAppKeys, err = s.daoManager.RelationAppKeyDAO.GetListWithApp(ctx, app)
		if err != nil {
			s.Error(rpcCtx, "internal error", "ListPrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	var list []*connector_api.ListPrivateKeyResponse_ListItem
	for _, relation := range relationAppKeys {
		if !strings.HasPrefix(relation.KeyID, "pswd_") {
			list = append(list, &connector_api.ListPrivateKeyResponse_ListItem{
				App:       relation.App,
				KeyID:     relation.KeyID,
				CreatedAt: relation.CreatedAt.UnixMilli(),
				PublicKey: relation.PasswordHash,
			})
		}
	}
	return &connector_api.ListPrivateKeyResponse_Data{List: list}, nil
}

func (s *ConnectorConsoleService) AddPrivateKey(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, app, keyID string) (*connector_api.AddPrivateKeyResponse_Data, *gerror.AppError) {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "AddPrivateKey", util.PermWrite, apiKey, startTS, app, "")
	if appError != nil {
		return nil, appError
	}
	var err error
	// manage platform log
	defer func() {
		s.PlatformLog(ctx, rpcCtx, "AddPrivateKey", appKey, startTS, err, app, keyID, fmt.Sprintf(`{"app": "%s", "keyID": "%s"}`, app, keyID))
	}()
	var relationAppKey *RelationAppKey
	var publicKey string
	if err = s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		// first, create a relation_app_key record, if not exists
		relationAppKey, err = daoManager.RelationAppKeyDAO.GetByKeyID(ctx, keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "AddPrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return err
		}
		if relationAppKey != nil {
			err = errors.New("keyID has existed")
			s.Error(rpcCtx, "bad request", "AddPrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return err
		}
		relationAppKey = &RelationAppKey{
			App:   app,
			KeyID: keyID,
		}
		// check key existing
		respData, err := keystore.CheckKeyExisting(keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "CheckKeyExisting", time.Since(startTS).Milliseconds(), err, apiKey, keyID, "")
			return err
		}
		if !respData.Data.Existing {
			if err = daoManager.RelationAppKeyDAO.Create(ctx, relationAppKey); err != nil {
				s.Error(rpcCtx, "internal error", "AddPrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
			// second, post to keystore
			respData, err := keystore.CreatePrivateKey(keyID)
			if err != nil {
				s.Error(rpcCtx, "internal error", "AddPrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
			publicKey = respData.Data.PublicKey
			// third, update the relation_app_key record
			if err = daoManager.RelationAppKeyDAO.UpdatePasswordHashByID(ctx, relationAppKey.ID, publicKey); err != nil {
				s.Error(rpcCtx, "internal error", "AddPrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
		} else {
			respData, err := keystore.GetPublicKey(keyID)
			if err != nil {
				s.Error(rpcCtx, "internal error", "AddPrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
			relationAppKey.PasswordHash = respData.Data.PublicKey
			if err = daoManager.RelationAppKeyDAO.Create(ctx, relationAppKey); err != nil {
				s.Error(rpcCtx, "internal error", "AddPrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "AddPrivateKey", time.Since(startTS).Milliseconds(), apiKey, app, keyID)
	return &connector_api.AddPrivateKeyResponse_Data{
		KeyID:     keyID,
		CreatedAt: relationAppKey.CreatedAt.UnixMilli(),
		PublicKey: publicKey,
	}, nil
}

func (s *ConnectorConsoleService) RemovePrivateKey(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey, app, keyID string) *gerror.AppError {
	// check permission
	appKey, appError := s.CheckPermission(ctx, rpcCtx, false, "RemovePrivateKey", util.PermWrite, apiKey, startTS, app, "")
	if appError != nil {
		return appError
	}
	var err error
	// manage platform log
	defer func() {
		s.PlatformLog(ctx, rpcCtx, "RemovePrivateKey", appKey, startTS, err, app, keyID, fmt.Sprintf(`{"app": "%s", "keyID": "%s"}`, app, keyID))
	}()
	if err = s.daoManager.DB.Transaction(func(tx *gorm.DB) error {
		daoManager := dao.NewManagerWithDB(tx)
		// first, delete the relation_app_key record, if exists
		err := daoManager.RelationAppKeyDAO.DeleteByAppWithKeyID(ctx, app, keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "RemovePrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return err
		}
		// second, post to keystore
		_, err = keystore.DeletePrivateKey(keyID)
		if err != nil {
			s.Error(rpcCtx, "internal error", "RemovePrivateKey", time.Since(startTS).Milliseconds(), err, apiKey, app, "")
			return err
		}
		return nil
	}); err != nil {
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "RemovePrivateKey", time.Since(startTS).Milliseconds(), apiKey, app, keyID)
	return nil
}

func (s *ConnectorConsoleService) GetLogs(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey string, pageNumber, pageSize int32) (*connector_api.GetLogsResponse_Data, *gerror.AppError) {
	// check permission
	_, appError := s.CheckPermission(ctx, rpcCtx, true, "GetLogs", util.PermRead, apiKey, startTS, "", "")
	if appError != nil {
		return nil, appError
	}
	respData, err := keystore.GetLogs(int(pageNumber), int(pageSize))
	if err != nil {
		s.Error(rpcCtx, "internal error", "GetLogs", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	s.Info(rpcCtx, "success", "GetLogs", time.Since(startTS).Milliseconds(), apiKey, "", "")
	return &connector_api.GetLogsResponse_Data{
		Total: int32(respData.Data.Total),
		List:  respData.Data.List,
	}, nil
}

func (s *ConnectorConsoleService) GetMonitorInfos(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey string) (*connector_api.GetMonitorInfosResponse_Data, *gerror.AppError) {
	// check permission
	_, appError := s.CheckPermission(ctx, rpcCtx, true, "GetMonitorInfos", util.PermRead, apiKey, startTS, "", "")
	if appError != nil {
		return nil, appError
	}
	respData, err := keystore.GetMonitorInfos()
	if err != nil {
		s.Error(rpcCtx, "internal error", "GetMonitorInfos", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var res []*connector_api.GetMonitorInfosResponse_InfoItem
	for _, info := range respData.Data.Infos {
		res = append(res, &connector_api.GetMonitorInfosResponse_InfoItem{
			Name:  info.Name,
			Value: info.Value,
		})
	}
	s.Info(rpcCtx, "success", "GetMonitorInfos", time.Since(startTS).Milliseconds(), apiKey, "", "")
	return &connector_api.GetMonitorInfosResponse_Data{
		Infos: res,
	}, nil
}

func (s *ConnectorConsoleService) GetManagePlatformLogs(ctx context.Context, rpcCtx *rpc.Context, startTS time.Time, apiKey string, pageSize, pageNumber int32) (*connector_api.GetManagePlatformLogsResponse_Data, *gerror.AppError) {
	// check permission
	_, appError := s.CheckPermission(ctx, rpcCtx, true, "GetManagePlatformLogs", util.PermRead, apiKey, startTS, "", "")
	if appError != nil {
		return nil, appError
	}
	total, err := s.daoManager.ManagePlatformLogDAO.TotalAll(ctx)
	if err != nil {
		s.Error(rpcCtx, "internal error", "GetManagePlatformLogs", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	logs, err := s.daoManager.ManagePlatformLogDAO.GetPagination(ctx, int(pageSize)*int(pageNumber), int(pageSize))
	if err != nil {
		s.Error(rpcCtx, "internal error", "GetManagePlatformLogs", time.Since(startTS).Milliseconds(), err, apiKey, "", "")
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var list []*connector_api.GetManagePlatformLogsResponse_ListItem
	for _, item := range logs {
		list = append(list, &connector_api.GetManagePlatformLogsResponse_ListItem{
			CreatedAt:  item.CreatedAt.UnixMilli(),
			Ip:         item.IP,
			Status:     item.Status,
			Object:     item.Object,
			Operation:  item.Operation,
			KeyID:      item.KeyID,
			App:        item.App,
			ApiKeyName: item.APIKeyName,
		})
	}
	s.Info(rpcCtx, "success", "GetManagePlatformLogs", time.Since(startTS).Milliseconds(), apiKey, "", "")
	return &connector_api.GetManagePlatformLogsResponse_Data{
		Total: total,
		List:  list,
	}, nil
}
