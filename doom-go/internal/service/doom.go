package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	cache "github.com/go-pkgz/expirable-cache"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/nextsurfer/doom-go/api/response"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/slark/api"
	slark_grpc "github.com/nextsurfer/slark/pkg/grpc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type DoomService struct {
	*AaveService
	*AssetService
	*CronService
	*EthService
	*BitCoinService
	*MarketService
	*UniswapService
	*UserService

	Logger          *zap.Logger
	RedisClient     *redis.Client
	MongoDB         *mongo.Database
	ConnectorApiKey string
	ConnectorKeyID  string
	Validator       *validator.Validate
	ExpirableCache  cache.Cache
}

func NewDoomService(
	ctx context.Context,
	logger *zap.Logger,
	redisClient *redis.Client,
	mongoDB *mongo.Database,
	connectorApiKey string,
	connectorKeyID string,
	validator *validator.Validate,
) (*DoomService, error) {
	doomService := &DoomService{
		Logger:          logger,
		RedisClient:     redisClient,
		MongoDB:         mongoDB,
		ConnectorApiKey: connectorApiKey,
		ConnectorKeyID:  connectorKeyID,
		Validator:       validator,
	}
	var err error
	// aave service
	doomService.AaveService = NewAaveService(doomService)
	// asset service
	doomService.AssetService = NewAssetService(doomService)
	// cron service
	// doomService.CronService, err = NewCronService(ctx, doomService)
	// if err != nil {
	// 	return nil, err
	// }
	// bitcoin service
	doomService.BitCoinService = NewBitCoinService(doomService)
	// eth service
	doomService.EthService = NewEthService(doomService)
	// market service
	doomService.MarketService, err = NewMarketService(doomService, false)
	if err != nil {
		return nil, err
	}
	// uniswap service
	doomService.UniswapService = NewUniswapService(doomService)
	// user service
	doomService.UserService = NewUserService(doomService)
	// expirable cache
	doomService.ExpirableCache, err = cache.NewCache(cache.LRU(), cache.TTL(time.Minute*5))
	if err != nil {
		return nil, err
	}
	return doomService, nil
}

func NewSimpleDoomService(
	redisClient *redis.Client,
	mongoDB *mongo.Database,
	connectorApiKey string,
	connectorKeyID string,
) (*DoomService, error) {
	doomService := &DoomService{
		RedisClient:     redisClient,
		MongoDB:         mongoDB,
		ConnectorApiKey: connectorApiKey,
		ConnectorKeyID:  connectorKeyID,
	}
	var err error
	// aave service
	doomService.AaveService = NewAaveService(doomService)
	// asset service
	doomService.AssetService = NewAssetService(doomService)
	// bitcoin service
	doomService.BitCoinService = NewBitCoinService(doomService)
	// eth service
	doomService.EthService = NewEthService(doomService)
	// market service
	doomService.MarketService, err = NewMarketService(doomService, true)
	if err != nil {
		return nil, err
	}
	// uniswap service
	doomService.UniswapService = NewUniswapService(doomService)
	// user service
	doomService.UserService = NewUserService(doomService)
	return doomService, nil
}

func (s *DoomService) SessionLoginInfo(ctx context.Context, rpcCtx *rpc.Context) (*api.LoginResponse, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	loginInfo, err := slark_grpc.SessionLoginInfo(ctx, rpcCtx)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	} else {
		if loginInfo.Code != 0 {
			err = fmt.Errorf("session error: %v", rpcCtx)
			logger.Error("session error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(loginInfo.Message).WithCode(loginInfo.Code)
		}
	}
	return loginInfo, nil
}

func (s *DoomService) ValidateRequest(ctx context.Context, rpcCtx *rpc.Context, request interface{}) *gerror.AppError {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	if err := s.Validator.Struct(request); err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	return nil
}

type performanceInfoKey struct{}

var PerformanceInfoKey = performanceInfoKey{}

type PerformanceInfo struct {
	Mu                sync.Mutex
	StartedAt         time.Time
	Web3CallCount     int
	Web3CallDuration  time.Duration
	MongoCallCount    int
	MongoCallDuration time.Duration
}

func StatisticMongoCall(ctx context.Context, ts time.Time) {
	performanceInfo := ctx.Value(PerformanceInfoKey).(*PerformanceInfo)
	performanceInfo.Mu.Lock()
	performanceInfo.MongoCallCount++
	performanceInfo.MongoCallDuration += time.Since(ts)
	performanceInfo.Mu.Unlock()
}

func StatisticWeb3Call(ctx context.Context, ts time.Time) {
	performanceInfo := ctx.Value(PerformanceInfoKey).(*PerformanceInfo)
	performanceInfo.Mu.Lock()
	performanceInfo.Web3CallCount++
	performanceInfo.Web3CallDuration += time.Since(ts)
	performanceInfo.Mu.Unlock()
}

func (s *DoomService) DeferLogPerformanceInfo(ctx context.Context, rpcCtx *rpc.Context, method string) {
	performanceInfo := ctx.Value(PerformanceInfoKey).(*PerformanceInfo)
	rpcCtx.Logger.Debug("performance_info",
		zap.String("method", method),
		zap.String("request_duration", time.Since(performanceInfo.StartedAt).String()),
		zap.Int("web3_call_count", performanceInfo.Web3CallCount),
		zap.String("web3_call_duration", performanceInfo.Web3CallDuration.String()),
		zap.Int("mongo_call_count", performanceInfo.MongoCallCount),
		zap.String("mongo_call_duration", performanceInfo.MongoCallDuration.String()),
	)
}
