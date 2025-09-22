package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	doom_api "github.com/nextsurfer/doom-go/api"
	"github.com/nextsurfer/doom-go/api/response"
	"github.com/nextsurfer/doom-go/internal/common/simplejson"
	"github.com/nextsurfer/doom-go/internal/service"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type DoomServer struct {
	doom_api.UnimplementedDoomServiceServer
	*service.DoomService
	env             gutil.APPEnvType
	logger          *zap.Logger
	localizeManager *localize.Manager
	validator       *validator.Validate
}

func NewDoomServer(ctx context.Context, env gutil.APPEnvType, logger *zap.Logger, redisClient *redis.Client, localizeManager *localize.Manager, connectorApiKey, connectorKeyID string, validator *validator.Validate, mongoDB *mongo.Database) (*DoomServer, error) {
	s := &DoomServer{
		env:             env,
		logger:          logger,
		localizeManager: localizeManager,
		validator:       validator,
	}
	// market service
	doomService, err := service.NewDoomService(ctx, logger, redisClient, mongoDB, connectorApiKey, connectorKeyID, validator)
	if err != nil {
		return nil, err
	}
	s.DoomService = doomService
	return s, nil
}

func (s *DoomServer) isEnvTest() bool {
	return s.env == gutil.AppEnvDEV || s.env == gutil.AppEnvPPE
}

func (s *DoomServer) logRequestData(rpcCtx *rpc.Context, method string, request interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(request)
		rpcCtx.Logger.Info(method, zap.String("request_data", data))
	}
}

func (s *DoomServer) deferLogResponseData(rpcCtx *rpc.Context, method string, response interface{}) {
	if s.isEnvTest() {
		data, _ := simplejson.DigestToJson(response)
		rpcCtx.Logger.Info(method, zap.String("response_data", data))
	}
}

type OracleFieldType struct {
	Duration int64 `json:"duration"` // millisecond
}

func (s *DoomServer) oracleField(startTS time.Time) string {
	data, _ := json.Marshal(OracleFieldType{Duration: int64(time.Since(startTS) / time.Millisecond)})
	return string(data)
}

// service methods --------------------------------------------------------------------------------------------

func (s *DoomServer) CreateSecurityQuestions(ctx context.Context, req *doom_api.CreateSecurityQuestionsRequest) (*doom_api.CreateSecurityQuestionsResponse, error) {
	startTS := time.Now()
	var resp doom_api.CreateSecurityQuestionsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreateSecurityQuestions", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreateSecurityQuestions", &resp) }()
	appError := s.UserService.CreateSecurityQuestions(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) ListSecurityQuestions(ctx context.Context, req *doom_api.ListSecurityQuestionsRequest) (*doom_api.ListSecurityQuestionsResponse, error) {
	startTS := time.Now()
	var resp doom_api.ListSecurityQuestionsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ListSecurityQuestions", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ListSecurityQuestions", &resp) }()
	data, appError := s.UserService.ListSecurityQuestions(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetSecurityQuestions(ctx context.Context, req *doom_api.GetSecurityQuestionsRequest) (*doom_api.GetSecurityQuestionsResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetSecurityQuestionsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetSecurityQuestions", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetSecurityQuestions", &resp) }()
	data, appError := s.UserService.GetSecurityQuestions(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) DeleteSecurityQuestions(ctx context.Context, req *doom_api.DeleteSecurityQuestionsRequest) (*doom_api.DeleteSecurityQuestionsResponse, error) {
	startTS := time.Now()
	var resp doom_api.DeleteSecurityQuestionsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeleteSecurityQuestions", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeleteSecurityQuestions", &resp) }()
	appError := s.UserService.DeleteSecurityQuestions(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) CreateData(ctx context.Context, req *doom_api.CreateDataRequest) (*doom_api.CreateDataResponse, error) {
	startTS := time.Now()
	var resp doom_api.CreateDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreateData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreateData", &resp) }()
	appError := s.UserService.CreateData(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) ListData(ctx context.Context, req *doom_api.ListDataRequest) (*doom_api.ListDataResponse, error) {
	startTS := time.Now()
	var resp doom_api.ListDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ListData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ListData", &resp) }()
	data, appError := s.UserService.ListData(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetData(ctx context.Context, req *doom_api.GetDataRequest) (*doom_api.GetDataResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetData", &resp) }()
	data, appError := s.UserService.GetData(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) DeleteData(ctx context.Context, req *doom_api.DeleteDataRequest) (*doom_api.DeleteDataResponse, error) {
	startTS := time.Now()
	var resp doom_api.DeleteDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeleteData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeleteData", &resp) }()
	appError := s.UserService.DeleteData(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetLatestSpotPrice(ctx context.Context, req *doom_api.GetLatestSpotPriceRequest) (*doom_api.GetLatestSpotPriceResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetLatestSpotPriceResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetLatestSpotPrice", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetLatestSpotPrice", &resp) }()
	data, appError := s.MarketService.GetLatestSpotPrice(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetLatestSpotPrices(ctx context.Context, req *doom_api.GetLatestSpotPricesRequest) (*doom_api.GetLatestSpotPricesResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetLatestSpotPricesResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetLatestSpotPrices", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetLatestSpotPrices", &resp) }()
	data, appError := s.MarketService.GetLatestSpotPrices(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetAssets(ctx context.Context, req *doom_api.GetAssetsRequest) (*doom_api.GetAssetsResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetAssetsResponse
	ctx = context.WithValue(ctx, service.PerformanceInfoKey, &service.PerformanceInfo{
		StartedAt: time.Now(),
	})
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetAssets", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetAssets", &resp) }()
	defer func() {
		if s.isEnvTest() {
			s.DeferLogPerformanceInfo(ctx, rpcCtx, "GetAssets")
		}
	}()
	data, appError := s.AssetService.GetAssets(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetUIKlines(ctx context.Context, req *doom_api.GetUIKlinesRequest) (*doom_api.GetUIKlinesResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetUIKlinesResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetUIKlines", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetUIKlines", &resp) }()
	data, appError := s.MarketService.GetUIKlines(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetDappAssets(ctx context.Context, req *doom_api.GetDappAssetsRequest) (*doom_api.GetDappAssetsResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetDappAssetsResponse
	ctx = context.WithValue(ctx, service.PerformanceInfoKey, &service.PerformanceInfo{
		StartedAt: time.Now(),
	})
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetDappAssets", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetDappAssets", &resp) }()
	defer func() {
		if s.isEnvTest() {
			s.DeferLogPerformanceInfo(ctx, rpcCtx, "GetDappAssets")
		}
	}()
	data, appError := s.AssetService.GetDappAssets(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetTokens(ctx context.Context, req *doom_api.GetTokensRequest) (*doom_api.GetTokensResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetTokensResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetTokens", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetTokens", &resp) }()
	data, appError := s.MarketService.GetTokens(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetDapps(ctx context.Context, req *doom_api.GetDappsRequest) (*doom_api.GetDappsResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetDappsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetDapps", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetDapps", &resp) }()
	data, appError := s.AssetService.GetDapps(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) FavoriteToken(ctx context.Context, req *doom_api.FavoriteTokenRequest) (*doom_api.FavoriteTokenResponse, error) {
	startTS := time.Now()
	var resp doom_api.FavoriteTokenResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "FavoriteToken", req)
	defer func() { s.deferLogResponseData(rpcCtx, "FavoriteToken", &resp) }()
	appError := s.UserService.FavoriteToken(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetFavoritedTokens(ctx context.Context, req *doom_api.GetFavoritedTokensRequest) (*doom_api.GetFavoritedTokensResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetFavoritedTokensResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetFavoritedTokens", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetFavoritedTokens", &resp) }()
	data, appError := s.UserService.GetFavoritedTokens(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetFavoritedLatestSpotPrices(ctx context.Context, req *doom_api.GetFavoritedLatestSpotPricesRequest) (*doom_api.GetFavoritedLatestSpotPricesResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetFavoritedLatestSpotPricesResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetFavoritedLatestSpotPrices", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetFavoritedLatestSpotPrices", &resp) }()
	data, appError := s.UserService.GetFavoritedLatestSpotPrices(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetTokenBalances(ctx context.Context, req *doom_api.GetTokenBalancesRequest) (*doom_api.GetTokenBalancesResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetTokenBalancesResponse
	ctx = context.WithValue(ctx, service.PerformanceInfoKey, &service.PerformanceInfo{
		StartedAt: time.Now(),
	})
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetTokenBalances", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetTokenBalances", &resp) }()
	defer func() {
		if s.isEnvTest() {
			s.DeferLogPerformanceInfo(ctx, rpcCtx, "GetTokenBalances")
		}
	}()
	data, appError := s.AssetService.GetTokenBalances(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetBalance(ctx context.Context, req *doom_api.GetBalanceRequest) (*doom_api.GetBalanceResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetBalanceResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetBalance", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetBalance", &resp) }()
	data, appError := s.AssetService.GetBalance(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetABI(ctx context.Context, req *doom_api.GetABIRequest) (*doom_api.GetABIResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetABIResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetABI", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetABI", &resp) }()
	data, appError := s.EthService.GetABI(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetTokenApprovals(ctx context.Context, req *doom_api.GetTokenApprovalsRequest) (*doom_api.GetTokenApprovalsResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetTokenApprovalsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetTokenApprovals", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetTokenApprovals", &resp) }()
	data, appError := s.EthService.GetTokenApprovals(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetAddress(ctx context.Context, req *doom_api.GetAddressRequest) (*doom_api.GetAddressResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetAddressResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetAddress", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetAddress", &resp) }()
	data, appError := s.BitCoinService.GetAddress(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetAddressTransactions(ctx context.Context, req *doom_api.GetAddressTransactionsRequest) (*doom_api.GetAddressTransactionsResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetAddressTransactionsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetAddressTransactions", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetAddressTransactions", &resp) }()
	data, appError := s.BitCoinService.GetAddressTransactions(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetAddressTransactionsMempool(ctx context.Context, req *doom_api.GetAddressTransactionsMempoolRequest) (*doom_api.GetAddressTransactionsMempoolResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetAddressTransactionsMempoolResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetAddressTransactionsMempool", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetAddressTransactionsMempool", &resp) }()
	data, appError := s.BitCoinService.GetAddressTransactionsMempool(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetGasFee(ctx context.Context, req *doom_api.GetGasFeeRequest) (*doom_api.GetGasFeeResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetGasFeeResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetGasFee", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetGasFee", &resp) }()
	data, appError := s.EthService.GetGasFee(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetEstimationOfConfirmationTime(ctx context.Context, req *doom_api.GetEstimationOfConfirmationTimeRequest) (*doom_api.GetEstimationOfConfirmationTimeResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetEstimationOfConfirmationTimeResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetEstimationOfConfirmationTime", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetEstimationOfConfirmationTime", &resp) }()
	data, appError := s.EthService.GetEstimationOfConfirmationTime(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

// internal interface --------------------------------------------------------------------------------

func (s *DoomServer) CreatePrivateKeyBackup(ctx context.Context, req *doom_api.CreatePrivateKeyBackupRequest) (*doom_api.CreatePrivateKeyBackupResponse, error) {
	startTS := time.Now()
	var resp doom_api.CreatePrivateKeyBackupResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreatePrivateKeyBackup", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreatePrivateKeyBackup", &resp) }()
	appError := s.UserService.CreatePrivateKeyBackup(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) ListPrivateKeyBackup(ctx context.Context, req *doom_api.ListPrivateKeyBackupRequest) (*doom_api.ListPrivateKeyBackupResponse, error) {
	startTS := time.Now()
	var resp doom_api.ListPrivateKeyBackupResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ListPrivateKeyBackup", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ListPrivateKeyBackup", &resp) }()
	data, appError := s.UserService.ListPrivateKeyBackup(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) GetPrivateKeyBackup(ctx context.Context, req *doom_api.GetPrivateKeyBackupRequest) (*doom_api.GetPrivateKeyBackupResponse, error) {
	startTS := time.Now()
	var resp doom_api.GetPrivateKeyBackupResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetPrivateKeyBackup", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetPrivateKeyBackup", &resp) }()
	data, appError := s.UserService.GetPrivateKeyBackup(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	resp.Data = data
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *DoomServer) DeletePrivateKeyBackup(ctx context.Context, req *doom_api.DeletePrivateKeyBackupRequest) (*doom_api.DeletePrivateKeyBackupResponse, error) {
	startTS := time.Now()
	var resp doom_api.DeletePrivateKeyBackupResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DeletePrivateKeyBackup", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DeletePrivateKeyBackup", &resp) }()
	appError := s.UserService.DeletePrivateKeyBackup(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}
