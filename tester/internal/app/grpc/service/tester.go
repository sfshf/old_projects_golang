package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	redisv8 "github.com/go-redis/redis/v8"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	tester_api "github.com/nextsurfer/tester/api"
	"github.com/nextsurfer/tester/api/response"
	"github.com/nextsurfer/tester/internal/pkg/dao"
	. "github.com/nextsurfer/tester/internal/pkg/model"
	tester_mongo "github.com/nextsurfer/tester/internal/pkg/mongo"
	"github.com/nextsurfer/tester/internal/pkg/redis"
	"github.com/nextsurfer/tester/internal/pkg/simplehttp"
	"github.com/nextsurfer/tester/internal/pkg/slark"
	"github.com/wxpusher/wxpusher-sdk-go"
	"github.com/wxpusher/wxpusher-sdk-go/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

type TesterService struct {
	Logger      *zap.Logger
	MongoDB     *mongo.Database
	DaoManager  *dao.Manager
	RedisOption *redis.Option
	S3Client    *s3.Client
}

func NewTesterService(ctx context.Context, logger *zap.Logger, mongoDB *mongo.Database, daoManager *dao.Manager, redisOption *redis.Option) (*TesterService, error) {
	TesterService := &TesterService{
		Logger:      logger,
		MongoDB:     mongoDB,
		DaoManager:  daoManager,
		RedisOption: redisOption,
	}
	// aws
	cfg, awsErr := config.LoadDefaultConfig(ctx, config.WithLogger(TesterService))
	if awsErr != nil {
		return nil, awsErr
	}
	// Create an Amazon S3 service client
	TesterService.S3Client = s3.NewFromConfig(cfg)
	// InitMessageNotificationConfig
	if err := TesterService.InitMessageNotificationConfig(ctx); err != nil {
		return nil, err
	}
	return TesterService, nil
}

func (s *TesterService) Logf(classification logging.Classification, format string, v ...interface{}) {
	log := fmt.Sprintf(format, v...)
	if classification == logging.Warn {
		zap.L().Warn("AWS Warning : ", zap.String("log", log))
	} else if classification == logging.Debug {
		zap.L().Debug("AWS Debug : ", zap.String("log", log))
	}
}

func (s *TesterService) UpdateAPITestcase(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.UpdateAPITestcaseRequest) *gerror.AppError {
	// check data format
	var apiTestcases []tester_mongo.ApiTestcase
	if err := json.Unmarshal([]byte(req.Data), &apiTestcases); err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	coll := s.MongoDB.Collection(tester_mongo.CollectionName_AppApiTestcases)
	if _, err := coll.ReplaceOne(
		ctx,
		bson.D{{Key: "app", Value: req.App}},
		tester_mongo.AppApiTestcases{App: req.App, ApiTestCases: apiTestcases},
		options.Replace().SetUpsert(true),
	); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *TesterService) GetApps(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetAppsRequest) (*tester_api.GetAppsResponse_Data, *gerror.AppError) {
	var list []string
	coll := s.MongoDB.Collection(tester_mongo.CollectionName_AppApiTestcases)
	cursor, err := coll.Find(ctx, bson.D{}, options.Find().SetProjection(bson.D{{Key: "app", Value: 1}, {Key: "apiTestCases", Value: -1}}))
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for cursor.Next(ctx) {
		var one tester_mongo.AppApiTestcases
		if err := cursor.Decode(&one); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		list = append(list, one.App)
	}
	if err := cursor.Err(); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &tester_api.GetAppsResponse_Data{
		List: list,
	}, nil
}

func (s *TesterService) GetAPITestcases(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetAPITestcasesRequest) (*tester_api.GetAPITestcasesResponse_Data, *gerror.AppError) {
	coll := s.MongoDB.Collection(tester_mongo.CollectionName_AppApiTestcases)
	var appApiTestcases tester_mongo.AppApiTestcases
	if err := coll.FindOne(ctx, bson.D{{Key: "app", Value: req.App}}).Decode(&appApiTestcases); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			return nil, nil
		}
	}
	var list []*tester_api.GetAPITestcasesResponse_ApiTestCase
	for _, item := range appApiTestcases.ApiTestCases {
		one := &tester_api.GetAPITestcasesResponse_ApiTestCase{
			Name: item.Name,
			Path: item.Path,
		}
		if item.Body != "" {
			one.Body = structpb.NewStringValue(item.Body)
		}
		list = append(list, one)
	}
	return &tester_api.GetAPITestcasesResponse_Data{List: list}, nil
}

func (s *TesterService) GetMysqlInfo(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetMysqlInfoRequest) (*tester_api.GetMysqlInfoResponse_Data, *gerror.AppError) {
	var list []*tester_api.GetMysqlInfoResponse_DatabaseInfo
	if err := s.DaoManager.DB.Raw(`SELECT table_schema AS "database", SUM(data_length + index_length) / 1024 / 1024 AS "size" FROM information_schema.TABLES GROUP BY table_schema;`).Scan(&list).Error; err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &tester_api.GetMysqlInfoResponse_Data{List: list}, nil
}

func (s *TesterService) GetMongoInfo(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetMongoInfoRequest) (*tester_api.GetMongoInfoResponse_Data, *gerror.AppError) {
	result, err := s.MongoDB.Client().ListDatabases(ctx, bson.D{})
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var list []*tester_api.GetMongoInfoResponse_DatabaseInfo
	for _, item := range result.Databases {
		list = append(list, &tester_api.GetMongoInfoResponse_DatabaseInfo{
			Database: item.Name,
			Size:     float64(item.SizeOnDisk) / 1024 / 1024,
		})
	}
	return &tester_api.GetMongoInfoResponse_Data{List: list}, nil
}

func (s *TesterService) GetBtcDiff(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetBtcDiffRequest) (*tester_api.GetBtcDiffResponse_Data, *gerror.AppError) {
	recordKey := "BTC/USDT"
	coll := s.MongoDB.Collection(tester_mongo.CollectionName_DiffCoinbaseBinance)
	var record tester_mongo.DiffCoinbaseBinance
	if err := coll.FindOne(ctx, bson.D{{Key: "key", Value: recordKey}}).Decode(&record); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return &tester_api.GetBtcDiffResponse_Data{
		PriceCB:      record.PriceCB,
		PriceBN:      record.PriceBN,
		PriceDiff:    record.PriceDiff,
		DiffPercent:  record.DiffPercent,
		ErrorMessage: record.ErrorMessage,
		UpdatedAt:    record.UpdatedAt,
	}, nil
}

func (s *TesterService) GetTxsMempoolInfo(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetTxsMempoolInfoRequest) (*tester_api.GetTxsMempoolInfoResponse_Data, *gerror.AppError) {
	address := "bc1qngydl7hmgdtmuqjmtsyj3pcwszv0yn5mj6kz4c"
	coll := s.MongoDB.Collection(tester_mongo.CollectionName_BtcTxsMempools)
	var record tester_mongo.BtcTxsMempools
	if err := coll.FindOne(ctx, bson.D{{Key: "key", Value: address}}).Decode(&record); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	var lastTS int64
	var msg string
	if record.Type == tester_mongo.MempoolType_Error {
		msg = record.Data
	} else {
		if time.Since(time.UnixMilli(record.UpdatedAt)) > 1*time.Minute {
			msg = "deadline error: monitor server maybe crash"
		} else {
			msg = record.Data
			lastTS = record.UpdatedAt
		}
	}
	return &tester_api.GetTxsMempoolInfoResponse_Data{LastTS: lastTS, Msg: msg}, nil
}

func (s *TesterService) GetSlarkRegistrationCaptchas(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetSlarkRegistrationCaptchasRequest) (*tester_api.GetSlarkRegistrationCaptchasResponse_Data, *gerror.AppError) {
	slarkDB := s.MongoDB.Client().Database("slark")
	coll := slarkDB.Collection(tester_mongo.CollectionName_RegistrationCaptcha)
	var list []*tester_api.GetSlarkRegistrationCaptchasResponse_EmailCaptcha
	var registrationCaptcha tester_mongo.RegistrationCaptcha
	if err := coll.FindOne(ctx, bson.D{}).Decode(&registrationCaptcha); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			return nil, nil
		}
	}
	for _, item := range registrationCaptcha.EmailCaptchas {
		list = append(list, &tester_api.GetSlarkRegistrationCaptchasResponse_EmailCaptcha{
			Email:     item.Email,
			Captcha:   item.Captcha,
			CreatedAt: item.CreatedAt,
		})
	}
	return &tester_api.GetSlarkRegistrationCaptchasResponse_Data{List: list}, nil
}

func (s *TesterService) GetSlarkLoginCaptchas(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetSlarkLoginCaptchasRequest) (*tester_api.GetSlarkLoginCaptchasResponse_Data, *gerror.AppError) {
	slarkDB := s.MongoDB.Client().Database("slark")
	coll := slarkDB.Collection(tester_mongo.CollectionName_LoginCaptcha)
	var list []*tester_api.GetSlarkLoginCaptchasResponse_EmailCaptcha
	var loginCaptcha tester_mongo.LoginCaptcha
	if err := coll.FindOne(ctx, bson.D{}).Decode(&loginCaptcha); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			return nil, nil
		}
	}
	for _, item := range loginCaptcha.EmailCaptchas {
		list = append(list, &tester_api.GetSlarkLoginCaptchasResponse_EmailCaptcha{
			Email:     item.Email,
			Captcha:   item.Captcha,
			CreatedAt: item.CreatedAt,
		})
	}
	return &tester_api.GetSlarkLoginCaptchasResponse_Data{List: list}, nil
}

func (s *TesterService) UploadApp(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.UploadAppRequest) *gerror.AppError {
	coll := s.MongoDB.Collection(tester_mongo.CollectionName_UploadApp)
	var uploadApp tester_mongo.UploadApp
	if err := coll.FindOne(ctx, bson.D{{Key: "appName", Value: req.AppName}}).Decode(&uploadApp); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			// new one
			uploadApp.AppName = req.AppName
			uploadApp.AppVersions = []tester_mongo.AppVersion{{Version: 1, Download: "https://d2y6ia7j6nkf8t.cloudfront.net/upload/" + req.AppHashName}}
			if _, err := coll.InsertOne(ctx, uploadApp); err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			return nil
		}
	}
	// update
	appLen := len(uploadApp.AppVersions)
	if appLen == 0 {
		uploadApp.AppVersions = append(uploadApp.AppVersions, tester_mongo.AppVersion{Version: 1, Download: "https://d2y6ia7j6nkf8t.cloudfront.net/upload/" + req.AppHashName})
	} else {
		uploadApp.AppVersions = append(uploadApp.AppVersions, tester_mongo.AppVersion{Version: uploadApp.AppVersions[appLen-1].Version + 1, Download: "https://d2y6ia7j6nkf8t.cloudfront.net/upload/" + req.AppHashName})
		if appLen == 3 {
			hashName0 := strings.TrimPrefix(uploadApp.AppVersions[0].Download, "https://d2y6ia7j6nkf8t.cloudfront.net/upload/")
			// delete s3
			if _, err := s.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(os.Getenv("UPLOAD_BUCKET_NAME")),
				Key:    aws.String("upload/" + hashName0),
			}); err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			uploadApp.AppVersions = uploadApp.AppVersions[1:]
		}
	}
	if _, err := coll.UpdateByID(ctx, uploadApp.ID, bson.D{{Key: "$set", Value: bson.D{{Key: "appVersions", Value: uploadApp.AppVersions}}}}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *TesterService) GetUploadedApps(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetUploadedAppsRequest) (*tester_api.GetUploadedAppsResponse_Data, *gerror.AppError) {
	coll := s.MongoDB.Collection(tester_mongo.CollectionName_UploadApp)
	var list []*tester_api.GetUploadedAppsResponse_UploadedApp
	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for cursor.Next(ctx) {
		var one tester_mongo.UploadApp
		if err := cursor.Decode(&one); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		item := &tester_api.GetUploadedAppsResponse_UploadedApp{
			Id:      one.ID.Hex(),
			AppName: one.AppName,
		}
		for _, ver := range one.AppVersions {
			item.AppVersions = append(item.AppVersions, &tester_api.GetUploadedAppsResponse_AppVersion{
				Version:  int32(ver.Version),
				Download: ver.Download,
			})
		}
		list = append(list, item)
	}
	if err := cursor.Err(); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &tester_api.GetUploadedAppsResponse_Data{List: list}, nil
}

const (
	RedisKey_MessageNotificationConfig = "Tester::MessageNotificationConfig"
)

type MessageNotificationConfig struct {
	UseTelegram  bool     `json:"useTelegram"`
	UseWxpusher  bool     `json:"useWxpusher"`
	WxpusherUIDs []string `json:"wxpusherUIDs"`
	UseEmail     bool     `json:"useEmail"`
	Emails       []string `json:"emails"`
}

func (s *TesterService) InitMessageNotificationConfig(ctx context.Context) error {
	coll := s.MongoDB.Collection(tester_mongo.CollectionName_MessageNotificationConfig)
	var config tester_mongo.MessageNotificationConfig
	if err := coll.FindOne(ctx, bson.D{}).Decode(&config); err != nil {
		if err != mongo.ErrNoDocuments {
			return err
		} else {
			// new one
			config.UseTelegram = true
			config.UseWxpusher = true
			config.WxpusherUIDs = []string{"UID_EiQixp190EoBXVgP7jd9XYxUWgph"}
			config.UseEmail = true
			config.Emails = []string{"luoxianmingg@gmail.com"}
			if _, err := coll.InsertOne(ctx, config); err != nil {
				return err
			}
		}
	}
	// reset redis cache
	configJson, err := json.Marshal(&MessageNotificationConfig{
		UseTelegram:  config.UseTelegram,
		UseWxpusher:  config.UseWxpusher,
		WxpusherUIDs: config.WxpusherUIDs,
		UseEmail:     config.UseEmail,
		Emails:       config.Emails,
	})
	if err != nil {
		return err
	}
	if err := s.RedisOption.Client.Set(ctx, RedisKey_MessageNotificationConfig, configJson, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (s *TesterService) GetMessageNotificationConfig(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetMessageNotificationConfigRequest) (*tester_api.GetMessageNotificationConfigResponse_Data, *gerror.AppError) {
	var config tester_api.GetMessageNotificationConfigResponse_Data
	// get config from redis
	configJson, err := s.RedisOption.Client.Get(ctx, RedisKey_MessageNotificationConfig).Result()
	if err != nil {
		if err != redisv8.Nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			// get config from mongo
			coll := s.MongoDB.Collection(tester_mongo.CollectionName_MessageNotificationConfig)
			var configFromDB tester_mongo.MessageNotificationConfig
			if err := coll.FindOne(ctx, bson.D{}).Decode(&configFromDB); err != nil {
				if err != mongo.ErrNoDocuments {
					rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
					return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				} else {
					return nil, nil
				}
			}
			config.UseTelegram = configFromDB.UseTelegram
			config.UseWxpusher = configFromDB.UseWxpusher
			config.WxpusherUIDs = configFromDB.WxpusherUIDs
			config.UseEmail = configFromDB.UseEmail
			config.Emails = configFromDB.Emails
		}
	} else {
		if err := json.Unmarshal([]byte(configJson), &config); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return &config, nil
}

func (s *TesterService) UpdateMessageNotificationConfig(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.UpdateMessageNotificationConfigRequest) *gerror.AppError {
	coll := s.MongoDB.Collection(tester_mongo.CollectionName_MessageNotificationConfig)
	// filter arguments
	if len(req.WxpusherUIDs) > 3 {
		req.WxpusherUIDs = req.WxpusherUIDs[:3]
	}
	if len(req.Emails) > 2 {
		req.Emails = req.Emails[:2]
	}
	var config tester_mongo.MessageNotificationConfig
	if err := coll.FindOne(ctx, bson.D{}).Decode(&config); err != nil {
		if err != mongo.ErrNoDocuments {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			// new one
			if _, err := coll.InsertOne(ctx, tester_mongo.MessageNotificationConfig{
				UseTelegram:  req.UseTelegram,
				UseWxpusher:  req.UseWxpusher,
				WxpusherUIDs: req.WxpusherUIDs,
				UseEmail:     req.UseEmail,
				Emails:       req.Emails,
			}); err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	} else {
		// update config
		if _, err := coll.UpdateByID(
			ctx,
			config.ID,
			bson.D{{Key: "$set", Value: bson.D{
				{Key: "useTelegram", Value: req.UseTelegram},
				{Key: "useWxpusher", Value: req.UseTelegram},
				{Key: "wxpusherUIDs", Value: req.WxpusherUIDs},
				{Key: "useEmail", Value: req.UseEmail},
				{Key: "emails", Value: req.Emails},
			}}},
		); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	// reset redis cache
	configJson, err := json.Marshal(&MessageNotificationConfig{
		UseTelegram:  req.UseTelegram,
		UseWxpusher:  req.UseWxpusher,
		WxpusherUIDs: req.WxpusherUIDs,
		UseEmail:     req.UseEmail,
		Emails:       req.Emails,
	})
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if err := s.RedisOption.Client.Set(ctx, RedisKey_MessageNotificationConfig, configJson, 0).Err(); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *TesterService) SendMessageByTelegram(message string) error {
	resp, err := simplehttp.PostJsonRequest(
		`https://api.telegram.org/bot1678806156:AAE8cWdlygrGCHWmHElQHNJ0ZjOv1IRQGeg/sendMessage`,
		map[string]string{
			"Accept":     "*/*",
			"Host":       "api.telegram.org",
			"User-Agent": `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.2 Safari/605.1.15`,
		},
		struct {
			ChatID string `json:"chat_id"`
			Text   string `json:"text"`
		}{
			ChatID: "1417969737",
			Text:   message,
		},
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("!!! Telegram Request Error: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("!!! Telegram Request Error: %s", simplehttp.ErrResponseStatusCodeNotEqualTo200)
	}
	return nil
}

func (s *TesterService) SendMessageByWxpusher(content, summary string, uids []string) error {
	if len(uids) == 0 {
		return nil
	}
	if _, err := wxpusher.SendMessage(&model.Message{
		AppToken:    "AT_4MTVNvvjupPNKFSxcg6k5ezDxhh25rBa",
		Content:     content,
		Summary:     summary,
		ContentType: 2,
		UIds:        uids,
	}); err != nil {
		return err
	}
	return nil
}

var (
	_emailServerHost              = "box.n1xt.net"
	_emailServerPort              = 465
	_emailUsername                = "noreply@n1xt.net"
	_emailPassword                = "xernyh-hyktyg13"
	_emailFrom                    = "noreply@n1xt.net"
	_MessageEmailHtmlTemplateText = `<!DOCTYPE html>
	<html>
		<head>
			<title>Retrieve PSWDS Unlock Password</title>
			<style type="text/css">
				:root {
					box-sizing: border-box;
				}
				*, ::before, ::after {
					box-sizing: inherit;
				}
				body {
					font-family: Arial, Helvetica, sans-serif;
					margin: 0;
				}
				.container {
					max-width: 680px;
					margin: 0 auto;
					text-align: center;
				}
			</style>
		</head>
		<body>
			<header>
			</header>
			<div class="container">
				<p>{{.Message}}</p>
				<p>NextSurfer 账号团队敬上</p>
			</div>
		</body>
	</html>`
)

func (s *TesterService) SendEmail(email, message string) error {
	// generate email message
	msg := gomail.NewMessage()
	msg.SetHeader("To", email)
	msg.SetHeader("From", _emailFrom)
	msg.SetHeader("Subject", "NextSurfer Message Email")
	// generate html text, and setted to email body
	data := struct {
		Message string
	}{
		Message: message,
	}
	var buf bytes.Buffer
	emailHtmlTemplate, err := template.New("MessageEmailHtmlTemplate").Parse(_MessageEmailHtmlTemplateText)
	if err != nil {
		return err
	}
	if err := emailHtmlTemplate.Execute(&buf, data); err != nil {
		return err
	}
	msg.SetBody("text/html", buf.String())
	dialer := gomail.NewDialer(_emailServerHost, _emailServerPort, _emailUsername, _emailPassword)
	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}

func (s *TesterService) SendMessage(ctx context.Context, rpcCtx *rpc.Context, config *MessageNotificationConfig, message string) error {
	log := tester_mongo.MessageNotificationLog{
		Message:   message,
		CreatedAt: time.Now().Unix(),
	}
	if config.UseTelegram {
		log.Mode |= tester_mongo.Ltelegram
		go func() {
			s.SendMessageByTelegram(message)
		}()
	}
	if config.UseWxpusher {
		log.Mode |= tester_mongo.Lwxpusher
		go func() {
			s.SendMessageByWxpusher(message, "NextSurfer Message", config.WxpusherUIDs)
		}()
	}
	if config.UseEmail {
		log.Mode |= tester_mongo.Lemail
		for _, email := range config.Emails {
			go func(email string) {
				s.SendEmail(email, message)
			}(email)
		}
	}
	// add log
	s.MongoDB.Collection(tester_mongo.CollectionName_MessageNotificationLog).InsertOne(ctx, log)
	return nil
}

func (s *TesterService) SendMessageNotification(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.SendMessageNotificationRequest) *gerror.AppError {
	var config MessageNotificationConfig
	// get config from redis
	configJson, err := s.RedisOption.Client.Get(ctx, RedisKey_MessageNotificationConfig).Result()
	if err != nil {
		if err != redisv8.Nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			// get config from mongo
			coll := s.MongoDB.Collection(tester_mongo.CollectionName_MessageNotificationConfig)
			var configFromDB tester_mongo.MessageNotificationConfig
			if err := coll.FindOne(ctx, bson.D{}).Decode(&configFromDB); err != nil {
				rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
				return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			config.UseTelegram = configFromDB.UseTelegram
			config.UseWxpusher = configFromDB.UseWxpusher
			config.WxpusherUIDs = configFromDB.WxpusherUIDs
			config.UseEmail = configFromDB.UseEmail
			config.Emails = configFromDB.Emails
		}
	} else {
		if err := json.Unmarshal([]byte(configJson), &config); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	// send message
	if err := s.SendMessage(ctx, rpcCtx, &config, "Test Send Message"); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *TesterService) GetMessageNotificationLogs(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetMessageNotificationLogsRequest) (*tester_api.GetMessageNotificationLogsResponse_Data, *gerror.AppError) {
	var list []*tester_api.GetMessageNotificationLogsResponse_Log
	coll := s.MongoDB.Collection(tester_mongo.CollectionName_MessageNotificationLog)
	cursor, err := coll.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(10))
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	for cursor.Next(ctx) {
		var one tester_mongo.MessageNotificationLog
		if err := cursor.Decode(&one); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		list = append(list, &tester_api.GetMessageNotificationLogsResponse_Log{
			Id:        one.ID.Hex(),
			Mode:      one.Mode,
			Message:   one.Message,
			CreatedAt: one.CreatedAt,
		})
	}
	if err := cursor.Err(); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &tester_api.GetMessageNotificationLogsResponse_Data{List: list}, nil
}

func (s *TesterService) GetPrivacyEmailAccounts(ctx context.Context, rpcCtx *rpc.Context, req *tester_api.GetPrivacyEmailAccountsRequest) (*tester_api.GetPrivacyEmailAccountsResponse_Data, *gerror.AppError) {
	var err error
	var accounts []*PrivacyEmailAccount
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		tx.Exec("USE pswds;") // switch to pswds db
		daoManager := dao.NewPswdsManagerWithDB(tx)
		accounts, err = daoManager.PrivacyEmailAccountDAO.GetAll(ctx)
		if err != nil {
			return err
		}
		tx.Exec("USE oracle;")
		return nil
	}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var list []*tester_api.GetPrivacyEmailAccountsResponse_Account
	for _, item := range accounts {
		userInfo, _ := slark.GetUserInfo(ctx, rpcCtx, item.UserID)
		list = append(list, &tester_api.GetPrivacyEmailAccountsResponse_Account{
			Id:           item.ID,
			UserID:       item.UserID,
			UserEmail:    userInfo.Email,
			EmailAccount: item.EmailAccount,
			Password:     item.Password,
		})
	}
	return &tester_api.GetPrivacyEmailAccountsResponse_Data{List: list}, nil
}
