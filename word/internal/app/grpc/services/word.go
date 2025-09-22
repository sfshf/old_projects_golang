package services

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	word_api "github.com/nextsurfer/word/api"
	"github.com/nextsurfer/word/api/response"
	"github.com/nextsurfer/word/internal/pkg/dao"
	. "github.com/nextsurfer/word/internal/pkg/model"
	"github.com/nextsurfer/word/internal/pkg/redis"
	"github.com/nextsurfer/word/internal/pkg/util"
	"go.uber.org/zap"
)

// WordService : service is pure business
type WordService struct {
	logger      *zap.Logger
	daoManager  *dao.Manager
	redisOption *redis.Option
	s3Client    *s3.Client
	pollyClient *polly.Client
}

// NewWordService is factory function
func NewWordService(logger *zap.Logger, daoManager *dao.Manager, redisOption *redis.Option) *WordService {
	service := &WordService{
		logger:      logger,
		daoManager:  daoManager,
		redisOption: redisOption,
	}
	// aws
	cfg, awsErr := config.LoadDefaultConfig(context.TODO(), config.WithLogger(service))
	if awsErr != nil {
		logger.Panic("aws sdk LoadDefaultConfig failed: ", zap.NamedError("appError", awsErr))
	}
	// Create an Amazon S3 service client
	service.s3Client = s3.NewFromConfig(cfg)
	// polly service
	service.pollyClient = polly.NewFromConfig(cfg)
	return service
}

// aws logger
func (s *WordService) Logf(classification logging.Classification, format string, v ...interface{}) {
	log := fmt.Sprintf(format, v...)
	if classification == logging.Warn {
		zap.L().Warn("AWS Warning : ", zap.String("log", log))
	} else if classification == logging.Debug {
		zap.L().Debug("AWS Debug : ", zap.String("log", log))
	}
}

func (s *WordService) FavoriteDefinition(ctx context.Context, rpcCtx *rpc.Context, definitionID int64, userID int64) *gerror.AppError {
	// check old log
	favorite, err := s.daoManager.FavoriteDefinitionDAO.GetByUserIDAndDefinitionID(ctx, userID, definitionID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if favorite == nil {
		// do favorite, insert one record
		one := &FavoriteDefinition{
			UserID:       userID,
			DefinitionID: definitionID,
		}
		if err := s.daoManager.FavoriteDefinitionDAO.Create(ctx, one); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		return nil
	}
	if favorite.DeletedAt > 0 {
		// do favorite, recover the one
		if err := s.daoManager.FavoriteDefinitionDAO.RecoverByID(ctx, favorite.ID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	} else if favorite.DeletedAt == 0 {
		// do disfavor, soft-delete the record
		if err := s.daoManager.FavoriteDefinitionDAO.DeleteByID(ctx, favorite.ID); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return nil
}

func (s *WordService) FavoritedDefinitions(ctx context.Context, rpcCtx *rpc.Context, userID int64) (*word_api.FavoritedDefinitionsResponse_Data, *gerror.AppError) {
	favorites, err := s.daoManager.FavoriteDefinitionDAO.GetAllByUserID(ctx, userID)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var res []int64
	for _, favorite := range favorites {
		res = append(res, favorite.DefinitionID)
	}
	return &word_api.FavoritedDefinitionsResponse_Data{Definitions: res}, nil
}

func (s *WordService) ProgressBackupStatus(ctx context.Context, rpcCtx *rpc.Context, timestamp int64, version int32, userID int64) (*word_api.ProgressBackupStatusResponse_Data, *gerror.AppError) {
	backupLog, err := s.daoManager.ProgressBackupDAO.GetLatestByUserIDAndVersion(ctx, userID, version)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if backupLog == nil {
		return &word_api.ProgressBackupStatusResponse_Data{
			Code: 1,
		}, nil
	}
	// compare timestamp
	var statusCode int32                       // nothing to do
	cts := time.UnixMilli(timestamp)           // client timestamp
	sts := time.UnixMilli(backupLog.Timestamp) // server timestamp
	if cts.After(sts) {
		statusCode = 1 // need to upload
	} else if cts.Before(sts) {
		statusCode = 2 // need to download
	}
	return &word_api.ProgressBackupStatusResponse_Data{
		Code: statusCode,
	}, nil
}

func getHashedPath(version int32, userID, timestamp int64, data []byte) string {
	h := md5.New()
	h.Write(data)
	return strings.ToLower(fmt.Sprintf("%d/%d/%d.%x", version, userID, timestamp, h.Sum(nil)))
}

func (s *WordService) UploadProgressBackup(ctx context.Context, rpcCtx *rpc.Context, timestamp int64, version int32, data string, userID int64) *gerror.AppError {
	backupLog, err := s.daoManager.ProgressBackupDAO.GetLatestByUserIDAndVersion(ctx, userID, version)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	cts := time.UnixMilli(timestamp) // client timestamp
	if backupLog != nil {
		// validate timestamp
		sts := time.UnixMilli(backupLog.Timestamp)
		if !cts.After(sts) {
			err = fmt.Errorf("invalid client timestamp: %s is not after server timestamp %s", cts, sts)
			rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ParamErrMsg_TimestampIsInvalid")).WithCode(response.StatusCodeBadRequest)
		}
	}
	// AES16+CBC encrypt the backup data
	dataBytes, err := util.AES16CBCEncrypt([]byte(data), []byte(os.Getenv("STUDY_BACKUP_KEY")))
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// put data to aws s3
	resource := getHashedPath(version, userID, timestamp, dataBytes)
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("PROGRESS_BACKUP_BUCKET_NAME")),
		Key:    aws.String(resource),
		Body:   bytes.NewReader(dataBytes),
	})
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// save the log to db
	if err := s.daoManager.ProgressBackupDAO.Create(ctx, &ProgressBackup{
		Timestamp: timestamp,
		UserID:    userID,
		Version:   version,
		Resource:  resource,
	}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

func (s *WordService) DownloadProgressBackup(ctx context.Context, rpcCtx *rpc.Context, timestamp int64, version int32, userID int64) (*word_api.DownloadProgressBackupResponse_Data, *gerror.AppError) {
	backupLog, err := s.daoManager.ProgressBackupDAO.GetLatestByUserIDAndVersion(ctx, userID, version)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	if backupLog == nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("RequestErrMsg")).WithCode(response.StatusCodeBadRequest)
	}
	// validate timestamp
	if cts, sts := time.UnixMilli(timestamp), time.UnixMilli(backupLog.Timestamp); !cts.Before(sts) {
		err = fmt.Errorf("invalid client timestamp: %s is not before server timestamp %s", cts, sts)
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("LoginErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
	}
	// get data from aws s3
	getObjectOutput, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("PROGRESS_BACKUP_BUCKET_NAME")),
		Key:    aws.String(backupLog.Resource),
	})
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	content, err := io.ReadAll(getObjectOutput.Body)
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	content, err = util.AES16CBCDecrypt(content, []byte(os.Getenv("STUDY_BACKUP_KEY")))
	if err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return &word_api.DownloadProgressBackupResponse_Data{Content: string(content)}, nil
}
