package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	connector_grpc "github.com/nextsurfer/connector/pkg/grpc"
	doom_api "github.com/nextsurfer/doom-go/api"
	"github.com/nextsurfer/doom-go/api/response"
	"github.com/nextsurfer/doom-go/internal/common/config"
	. "github.com/nextsurfer/doom-go/internal/model"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type UserService struct {
	*DoomService
}

func NewUserService(DoomService *DoomService) *UserService {
	return &UserService{
		DoomService: DoomService,
	}
}

func (s *UserService) CreateSecurityQuestions(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.CreateSecurityQuestionsRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// save data
	dataID := fmt.Sprintf("%d-sq-%s", loginInfo.Data.UserID, req.Title)
	if err := connector_grpc.SaveData(ctx, rpcCtx, s.ConnectorApiKey, s.ConnectorKeyID, dataID, req.PlainText, req.CipherText, true); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	ts := time.Now().UnixMilli()
	if _, err := s.MongoDB.Collection(CollectionName_SecurityQuestion).
		InsertOne(ctx, &SecurityQuestion{
			CreatedAt:   ts,
			UpdatedAt:   ts,
			UserID:      loginInfo.Data.UserID,
			DataID:      dataID,
			Title:       req.Title,
			Description: req.Description,
		}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *UserService) ListSecurityQuestions(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.ListSecurityQuestionsRequest) (*doom_api.ListSecurityQuestionsResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	var res []*doom_api.ListSecurityQuestionsResponse_ListItem
	cursor, err := s.MongoDB.Collection(CollectionName_SecurityQuestion).
		Find(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}})
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for cursor.Next(ctx) {
		var securityQuestion SecurityQuestion
		if err := cursor.Decode(&securityQuestion); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		res = append(res, &doom_api.ListSecurityQuestionsResponse_ListItem{
			Title:       securityQuestion.Title,
			DataID:      securityQuestion.DataID,
			Description: securityQuestion.Description,
			Date:        time.UnixMilli(securityQuestion.CreatedAt).Format("2006-01-02"),
		})
	}
	if err := cursor.Err(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &doom_api.ListSecurityQuestionsResponse_Data{List: res}, nil
}

func (s *UserService) GetSecurityQuestions(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetSecurityQuestionsRequest) (*doom_api.GetSecurityQuestionsResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	dataID := fmt.Sprintf("%d-sq-%s", loginInfo.Data.UserID, req.Title)

	var record SecurityQuestion
	if err := s.MongoDB.Collection(CollectionName_SecurityQuestion).
		FindOne(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}, {Key: "dataID", Value: dataID}}).
		Decode(&record); err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoSecurityQuestions")).WithCode(response.StatusCodeBadRequest)
		}
	}
	plaintext, err := connector_grpc.GetData(ctx, rpcCtx, s.ConnectorApiKey, s.ConnectorKeyID, record.DataID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &doom_api.GetSecurityQuestionsResponse_Data{PlainText: plaintext}, nil
}

func (s *UserService) DeleteSecurityQuestions(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.DeleteSecurityQuestionsRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	dataID := fmt.Sprintf("%d-sq-%s", loginInfo.Data.UserID, req.Title)

	coll := s.MongoDB.Collection(CollectionName_SecurityQuestion)
	var record SecurityQuestion
	if err := coll.FindOne(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}, {Key: "dataID", Value: dataID}}).
		Decode(&record); err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			logger.Warn("valid request with invalid params", zap.NamedError("appError", err))
			return nil
		}
	}
	if _, err := coll.UpdateByID(ctx, record.ID, bson.D{{Key: "$set", Value: bson.D{{Key: "deletedAt", Value: time.Now().UnixMilli()}}}}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	// delete data
	if err := connector_grpc.DeleteData(ctx, rpcCtx, s.ConnectorApiKey, s.ConnectorKeyID, record.DataID); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *UserService) CreateData(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.CreateDataRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// save data
	dataID := fmt.Sprintf("%d-data-%s", loginInfo.Data.UserID, req.Title)
	if err := connector_grpc.SaveData(ctx, rpcCtx, s.ConnectorApiKey, s.ConnectorKeyID, dataID, req.PlainText, req.CipherText, true); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	ts := time.Now().UnixMilli()
	if _, err := s.MongoDB.Collection(CollectionName_Datum).InsertOne(ctx, &Datum{
		CreatedAt: ts,
		UpdatedAt: ts,
		UserID:    loginInfo.Data.UserID,
		DataID:    dataID,
		Title:     req.Title,
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *UserService) ListData(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.ListDataRequest) (*doom_api.ListDataResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var res []*doom_api.ListDataResponse_ListItem
	cursor, err := s.MongoDB.Collection(CollectionName_Datum).Find(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}})
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for cursor.Next(ctx) {
		var data Datum
		if err := cursor.Decode(&data); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		res = append(res, &doom_api.ListDataResponse_ListItem{
			Title:  data.Title,
			DataID: data.DataID,
			Date:   time.UnixMilli(data.CreatedAt).Format("2006-01-02"),
		})
	}
	if err := cursor.Err(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &doom_api.ListDataResponse_Data{List: res}, nil
}

func (s *UserService) GetData(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetDataRequest) (*doom_api.GetDataResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	dataID := fmt.Sprintf("%d-data-%s", loginInfo.Data.UserID, req.Title)

	var record Datum
	if err := s.MongoDB.Collection(CollectionName_Datum).
		FindOne(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}, {Key: "dataID", Value: dataID}}).Decode(&record); err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			err = fmt.Errorf("data [userID=%d, dataID=%s] not exists", loginInfo.Data.UserID, dataID)
			logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoDatum")).WithCode(response.StatusCodeBadRequest)
		}
	}
	plaintext, err := connector_grpc.GetData(ctx, rpcCtx, s.ConnectorApiKey, s.ConnectorKeyID, record.DataID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &doom_api.GetDataResponse_Data{PlainText: plaintext}, nil
}

func (s *UserService) DeleteData(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.DeleteDataRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	dataID := fmt.Sprintf("%d-data-%s", loginInfo.Data.UserID, req.Title)

	coll := s.MongoDB.Collection(CollectionName_Datum)
	var record Datum
	if err := coll.FindOne(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}, {Key: "dataID", Value: dataID}}).Decode(&record); err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			logger.Warn("valid request with invalid params", zap.NamedError("appError", err))
			return nil
		}
	}
	if _, err := coll.UpdateByID(ctx, record.ID, bson.D{{Key: "$set", Value: bson.D{{Key: "deletedAt", Value: time.Now().UnixMilli()}}}}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	// delete data
	if err := connector_grpc.DeleteData(ctx, rpcCtx, s.ConnectorApiKey, s.ConnectorKeyID, record.DataID); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *UserService) FavoriteToken(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.FavoriteTokenRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	if req.Symbol = strings.ToUpper(req.Symbol); config.InReputableTokens(req.Symbol) == nil {
		err := fmt.Errorf("invalid token: %v", req.Symbol)
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalClientErrMsg_WrongRequestParametersErrMsg")).WithCode(response.StatusCodeUnsupportedCryptocurrency)
	}
	coll := s.MongoDB.Collection(CollectionName_FavoriteToken)
	// check old log
	var favorite FavoriteToken
	if err := coll.FindOne(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}, {Key: "symbol", Value: req.Symbol}}).Decode(&favorite); err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			// do favorite, insert one record
			ts := time.Now().UnixMilli()
			one := &FavoriteToken{
				CreatedAt: ts,
				UpdatedAt: ts,
				UserID:    loginInfo.Data.UserID,
				Symbol:    req.Symbol,
			}
			if _, err := coll.InsertOne(ctx, one); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			return nil
		}
	}
	if favorite.DeletedAt > 0 {
		// do favorite, recover the one
		if _, err := coll.UpdateByID(ctx, favorite.ID, bson.D{{Key: "$unset", Value: bson.D{{Key: "deletedAt", Value: ""}}}}); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	} else if favorite.DeletedAt == 0 {
		// do disfavor, soft-delete the record
		if _, err := coll.UpdateByID(ctx, favorite.ID, bson.D{{Key: "$set", Value: bson.D{{Key: "deletedAt", Value: time.Now().UnixMilli()}}}}); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return nil
}

func (s *UserService) GetFavoritedTokens(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetFavoritedTokensRequest) (*doom_api.GetFavoritedTokensResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var res []string
	cursor, err := s.MongoDB.Collection(CollectionName_FavoriteToken).Find(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}})
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for cursor.Next(ctx) {
		var favorite FavoriteToken
		if err := cursor.Decode(&favorite); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		res = append(res, favorite.Symbol)
	}
	if err := cursor.Err(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &doom_api.GetFavoritedTokensResponse_Data{List: res}, nil
}

func (s *UserService) GetFavoritedLatestSpotPrices(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetFavoritedLatestSpotPricesRequest) (*doom_api.GetFavoritedLatestSpotPricesResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var symbols []string
	cursor, err := s.MongoDB.Collection(CollectionName_FavoriteToken).Find(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}})
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for cursor.Next(ctx) {
		var favorite FavoriteToken
		if err := cursor.Decode(&favorite); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		symbols = append(symbols, favorite.Symbol)
	}
	if err := cursor.Err(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	data, appError := s.MarketService.GetLatestSpotPrices(ctx, rpcCtx, &doom_api.GetLatestSpotPricesRequest{
		Symbols:  symbols,
		BaseCoin: req.BaseCoin,
	})
	if appError != nil {
		return nil, appError
	}
	return &doom_api.GetFavoritedLatestSpotPricesResponse_Data{List: data.List}, nil
}

// internal interface --------------------------------------------------------------------------------

func (s *UserService) CreatePrivateKeyBackup(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.CreatePrivateKeyBackupRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	// save data
	dataID := fmt.Sprintf("%d-pkbp-%s", loginInfo.Data.UserID, req.Title)
	if err := connector_grpc.SaveData(ctx, rpcCtx, s.ConnectorApiKey, s.ConnectorKeyID, dataID, req.PlainText, req.CipherText, true); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	ts := time.Now().UnixMilli()
	if _, err := s.MongoDB.Collection(CollectionName_PrivateKeyBackup).InsertOne(ctx, &PrivateKeyBackup{
		CreatedAt: ts,
		UpdatedAt: ts,
		UserID:    loginInfo.Data.UserID,
		DataID:    dataID,
		Title:     req.Title,
	}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *UserService) ListPrivateKeyBackup(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.ListPrivateKeyBackupRequest) (*doom_api.ListPrivateKeyBackupResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var res []*doom_api.ListPrivateKeyBackupResponse_ListItem
	cursor, err := s.MongoDB.Collection(CollectionName_PrivateKeyBackup).Find(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}})
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for cursor.Next(ctx) {
		var backup PrivateKeyBackup
		if err := cursor.Decode(&backup); err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		res = append(res, &doom_api.ListPrivateKeyBackupResponse_ListItem{
			Title:  backup.Title,
			DataID: backup.DataID,
			Date:   time.UnixMilli(backup.CreatedAt).Format("2006-01-02"),
		})
	}
	if err := cursor.Err(); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &doom_api.ListPrivateKeyBackupResponse_Data{List: res}, nil
}

func (s *UserService) GetPrivateKeyBackup(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetPrivateKeyBackupRequest) (*doom_api.GetPrivateKeyBackupResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return nil, appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	dataID := fmt.Sprintf("%d-pkbp-%s", loginInfo.Data.UserID, req.Title)

	var record PrivateKeyBackup
	if err := s.MongoDB.Collection(CollectionName_PrivateKeyBackup).
		FindOne(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}, {Key: "dataID", Value: dataID}}).
		Decode(&record); err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_UserHasNoPrivateKeyBackup")).WithCode(response.StatusCodeBadRequest)
		}
	}
	plaintext, err := connector_grpc.GetData(ctx, rpcCtx, s.ConnectorApiKey, s.ConnectorKeyID, record.DataID)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &doom_api.GetPrivateKeyBackupResponse_Data{PlainText: plaintext}, nil
}

func (s *UserService) DeletePrivateKeyBackup(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.DeletePrivateKeyBackupRequest) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// get slark login info
	loginInfo, appError := s.SessionLoginInfo(ctx, rpcCtx)
	if appError != nil {
		return appError
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	dataID := fmt.Sprintf("%d-pkbp-%s", loginInfo.Data.UserID, req.Title)

	coll := s.MongoDB.Collection(CollectionName_PrivateKeyBackup)
	var record PrivateKeyBackup
	if err := coll.FindOne(ctx, bson.D{{Key: "userID", Value: loginInfo.Data.UserID}, {Key: "dataID", Value: dataID}}).
		Decode(&record); err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			logger.Warn("valid request with invalid params", zap.NamedError("appError", err))
			return nil
		}
	}
	if _, err := coll.UpdateByID(ctx, record.ID, bson.D{{Key: "$set", Value: bson.D{{Key: "deletedAt", Value: time.Now().UnixMilli()}}}}); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}

	// delete data
	if err := connector_grpc.DeleteData(ctx, rpcCtx, s.ConnectorApiKey, s.ConnectorKeyID, record.DataID); err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}
