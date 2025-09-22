package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
	doom_api "github.com/nextsurfer/doom-go/api"
	"github.com/nextsurfer/doom-go/api/response"
	"github.com/nextsurfer/doom-go/internal/common/config"
	"github.com/nextsurfer/doom-go/internal/common/eth"
	. "github.com/nextsurfer/doom-go/internal/model"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type AssetService struct {
	*DoomService
}

func NewAssetService(DoomService *DoomService) *AssetService {
	s := &AssetService{
		DoomService: DoomService,
	}
	return s
}

type TokenAsset struct {
	Symbol string
	Amount float64
	Price  string
	Value  float64
}

type DappAsset struct {
	Name         string
	TokenAddress string
	IsDebt       bool
	TotalValue   float64
	Holdings     []TokenAsset
}

type DappReward struct {
	Token  string
	Amount float64
	Price  string
	Value  float64
}

func (s *AssetService) ethTokens(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetAssetsRequest, client *ethclient.Client, res *doom_api.GetAssetsResponse_Data) (*UpsertUserErc20TokensResult, float64, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	upsertUserErc20TokensResult, appError := s.EthService.UpsertUserErc20Tokens(ctx, rpcCtx, client, req.Address, true)
	if appError != nil {
		return nil, 0, appError
	}
	if upsertUserErc20TokensResult.ToBlock > 0 {
		res.ToBlock = structpb.NewStringValue("0x" + strconv.FormatInt(upsertUserErc20TokensResult.ToBlock, 16))
	}
	var tokens []*doom_api.GetAssetsResponse_Token
	var unknownTokens []*doom_api.GetAssetsResponse_UnknownToken
	var tokenAssetsTotalValue float64
	for _, token := range upsertUserErc20TokensResult.Tokens {
		if token.Type == TokenTypeERC20 {
			one := &doom_api.GetAssetsResponse_Token{
				Balance: token.Balance,
				Address: token.Address,
				Name:    token.Name,
				Symbol:  token.Symbol,
			}
			if token.Price != "" {
				one.Balance = token.BalanceValue
				one.Price = structpb.NewStringValue(token.Price)
			}
			if token.Value != "" {
				one.Value = structpb.NewStringValue(token.Value)
				val, err := strconv.ParseFloat(token.Value, 64)
				if err != nil {
					logger.Error("internal error", zap.NamedError("appError", err))
					return nil, 0, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
				tokenAssetsTotalValue += val
			}
			tokens = append(tokens, one)
		} else if token.Type == TokenTypeError {
			unknownTokens = append(unknownTokens, &doom_api.GetAssetsResponse_UnknownToken{
				Balance: token.Balance,
				Address: token.Address,
			})
		}
	}
	res.Tokens = tokens
	res.UnknownTokens = unknownTokens
	return upsertUserErc20TokensResult, tokenAssetsTotalValue, nil
}

func (s *AssetService) ethBalance(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetAssetsRequest, client *ethclient.Client, res *doom_api.GetAssetsResponse_Data) (float64, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	contractAddress := common.HexToAddress(req.Address)
	ts := time.Now()
	balanceValue, err := client.BalanceAt(ctx, contractAddress, nil)
	StatisticWeb3Call(ctx, ts)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return 0, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	ethPrice, _ := s.MarketService.GetLatestSpotPrice(ctx, rpcCtx, &doom_api.GetLatestSpotPriceRequest{
		Symbol:   "ETH",
		BaseCoin: "USDT",
	})
	if ethPrice == nil || ethPrice.Price == "" {
		err = errors.New("eth price not found")
		logger.Error("internal error", zap.NamedError("appError", err))
		return 0, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	price, err := strconv.ParseFloat(ethPrice.Price, 64)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return 0, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var ethBalance doom_api.GetAssetsResponse_Balance
	balance := eth.Uint256ToFloat64(uint256.MustFromBig(balanceValue), 18)
	ethBalance.Balance = eth.FormatFloat64(balance)
	ethBalance.Price = ethPrice.Price
	ethValue := balance * price
	ethBalance.Value = eth.FormatFloat64(ethValue)
	res.Balance = &ethBalance
	return ethValue, nil
}

func (s *AssetService) aaveV3Assets(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetAssetsRequest, client *ethclient.Client, upsertUserErc20TokensResult *UpsertUserErc20TokensResult) (float64, *gerror.AppError) {
	dappAssets, appError := s.AaveService.GetAaveV3Assets(ctx, rpcCtx, client, req.Address)
	if appError != nil {
		return 0, appError
	}
	var aaveV3TotalValue float64
	for _, dappAsset := range dappAssets {
		for _, item := range dappAsset.Holdings {
			if item.Symbol != "" {
				price, _ := s.MarketService.GetLatestSpotPrice(ctx, rpcCtx, &doom_api.GetLatestSpotPriceRequest{
					Symbol:   item.Symbol,
					BaseCoin: "USDT",
				})
				if price == nil || price.Price == "" {
					continue
				}
				priceValue, err := strconv.ParseFloat(price.Price, 64)
				if err != nil {
					return 0, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
				value := priceValue * item.Amount
				if dappAsset.IsDebt {
					aaveV3TotalValue -= value
				} else {
					aaveV3TotalValue += value
				}
			}
		}
	}
	return aaveV3TotalValue, nil
}

func (s *AssetService) uniswapAssets(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetAssetsRequest, client *ethclient.Client, uniswapType string, upsertUserErc20TokensResult *UpsertUserErc20TokensResult) (float64, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	var err error
	var dappAssets []*DappAsset
	var dappRewards []*DappReward
	var appError *gerror.AppError
	switch uniswapType {
	case UniswapTokenTypeV2:
		dappAssets, err = s.UniswapService.GetUniswapV2Assets(ctx, rpcCtx, client, req.Address, upsertUserErc20TokensResult)
		if err != nil {
			logger.Error("bad request", zap.NamedError("appError", err))
			return 0, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
		}
	case UniswapTokenTypeV3:
		dappAssets, dappRewards, appError = s.UniswapService.GetUniswapV3Assets(ctx, rpcCtx, client, req.Address)
		if appError != nil {
			return 0, appError
		}
	}
	var uniswapTotalValue float64
	for _, dappAsset := range dappAssets {
		if dappAsset.IsDebt {
			uniswapTotalValue -= dappAsset.TotalValue
		} else {
			uniswapTotalValue += dappAsset.TotalValue
		}
	}
	for _, dappReward := range dappRewards {
		uniswapTotalValue += dappReward.Value
	}
	return uniswapTotalValue, nil
}

func (s *AssetService) GetAssets(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetAssetsRequest) (*doom_api.GetAssetsResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	chain, err := config.ChainByName(req.Chain)
	if err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
	}
	res := &doom_api.GetAssetsResponse_Data{}
	var tokenAssetsTotalValue float64
	// 1. tokens
	client, err := ethclient.DialContext(ctx, chain.Address)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	defer client.Close()
	upsertUserErc20TokensResult, totalValue, appError := s.ethTokens(ctx, rpcCtx, req, client, res)
	if appError != nil {
		return nil, appError
	}
	tokenAssetsTotalValue += totalValue
	// 2. eth balance
	ethValue, appError := s.ethBalance(ctx, rpcCtx, req, client, res)
	if appError != nil {
		return nil, appError
	}
	tokenAssetsTotalValue += ethValue
	// 3. dapp assets
	// aave_v3
	aaveV3TotalValue, appError := s.aaveV3Assets(ctx, rpcCtx, req, client, upsertUserErc20TokensResult)
	if appError != nil {
		return nil, appError
	}
	// uniswap_v2
	uniswapV2TotalValue, appError := s.uniswapAssets(ctx, rpcCtx, req, client, UniswapTokenTypeV2, upsertUserErc20TokensResult)
	if appError != nil {
		return nil, appError
	}
	// uniswap_v3
	uniswapV3TotalValue, appError := s.uniswapAssets(ctx, rpcCtx, req, client, UniswapTokenTypeV3, upsertUserErc20TokensResult)
	if appError != nil {
		return nil, appError
	}
	res.DappAssets = []*doom_api.GetAssetsResponse_DappAsset{
		{
			App:   "aave_v3",
			Value: eth.FormatFloat64(aaveV3TotalValue),
		},
		{
			App:   "uniswap_v2",
			Value: eth.FormatFloat64(uniswapV2TotalValue * 2),
		},
		{
			App:   "uniswap_v3",
			Value: eth.FormatFloat64(uniswapV3TotalValue),
		},
	}
	res.TotalValue = eth.FormatFloat64(tokenAssetsTotalValue + aaveV3TotalValue + uniswapV2TotalValue*2 + uniswapV3TotalValue)
	return res, nil
}

func (s *AssetService) handleGetDappAssetsResponseData(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetDappAssetsRequest, dappAssets []*DappAsset, dappRewards []*DappReward) (*doom_api.GetDappAssetsResponse_Data, *gerror.AppError) {
	var assets []*doom_api.GetDappAssetsResponse_Asset
	var debts []*doom_api.GetDappAssetsResponse_Asset
	var rewards []*doom_api.GetDappAssetsResponse_Reward
	var totalValue float64
	for _, dappAsset := range dappAssets {
		var assetTotalValue float64
		var holdings []*doom_api.GetDappAssetsResponse_Holding
		for _, item := range dappAsset.Holdings {
			holding := &doom_api.GetDappAssetsResponse_Holding{
				Token:  item.Symbol,
				Amount: eth.FormatFloat64(item.Amount),
			}
			if req.App == "aave_v3" && item.Symbol != "" {
				price, _ := s.MarketService.GetLatestSpotPrice(ctx, rpcCtx, &doom_api.GetLatestSpotPriceRequest{
					Symbol:   item.Symbol,
					BaseCoin: "USDT",
				})
				if price == nil || price.Price == "" {
					continue
				}
				priceValue, err := strconv.ParseFloat(price.Price, 64)
				if err != nil {
					return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
				value := priceValue * item.Amount
				assetTotalValue += value
				holding.Price = eth.FormatFloat64(priceValue)
				holding.Value = eth.FormatFloat64(value)
			} else {
				holding.Price = item.Price
				assetTotalValue += item.Value
				holding.Value = eth.FormatFloat64(item.Value)
			}
			holdings = append(holdings, holding)
		}
		one := &doom_api.GetDappAssetsResponse_Asset{
			Name:         dappAsset.Name,
			TokenAddress: dappAsset.TokenAddress,
			Holdings:     holdings,
		}
		if req.App == "uniswap_v2" {
			assetTotalValue *= 2
			one.TotalValue = eth.FormatFloat64(assetTotalValue)
		} else {
			one.TotalValue = eth.FormatFloat64(assetTotalValue)
		}
		if dappAsset.IsDebt {
			totalValue -= assetTotalValue
			debts = append(debts, one)
		} else {
			totalValue += assetTotalValue
			assets = append(assets, one)
		}
	}
	for _, dappReward := range dappRewards {
		reward := &doom_api.GetDappAssetsResponse_Reward{
			Token:  dappReward.Token,
			Amount: eth.FormatFloat64(dappReward.Amount),
		}
		if req.App == "aave_v3" && dappReward.Token != "" {
			price, _ := s.MarketService.GetLatestSpotPrice(ctx, rpcCtx, &doom_api.GetLatestSpotPriceRequest{
				Symbol:   dappReward.Token,
				BaseCoin: "USDT",
			})
			if price == nil || price.Price == "" {
				continue
			}
			priceValue, err := strconv.ParseFloat(price.Price, 64)
			if err != nil {
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			value := priceValue * dappReward.Amount
			reward.Price = eth.FormatFloat64(priceValue)
			reward.Value = eth.FormatFloat64(value)
			totalValue += value
		} else {
			reward.Price = dappReward.Price
			reward.Value = eth.FormatFloat64(dappReward.Value)
			totalValue += dappReward.Value
		}
		rewards = append(rewards, reward)
	}
	return &doom_api.GetDappAssetsResponse_Data{
		Assets:     assets,
		Debts:      debts,
		Rewards:    rewards,
		TotalValue: eth.FormatFloat64(totalValue),
	}, nil
}

func (s *AssetService) GetDappAssets(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetDappAssetsRequest) (*doom_api.GetDappAssetsResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	var err error
	chain, err := config.ChainByName(req.Chain)
	if err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
	}
	client, err := ethclient.DialContext(ctx, chain.Address)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	defer client.Close()
	var dappAssets []*DappAsset
	var dappRewards []*DappReward
	var appError *gerror.AppError
	switch req.App {
	case "aave_v3":
		dappAssets, appError = s.AaveService.GetAaveV3Assets(ctx, rpcCtx, client, req.Address)
		if appError != nil {
			return nil, appError
		}
	case "uniswap_v2":
		dappAssets, err = s.UniswapService.GetUniswapV2Assets(ctx, rpcCtx, client, req.Address, nil)
		if err != nil {
			logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
		}
	case "uniswap_v3":
		dappAssets, dappRewards, appError = s.UniswapService.GetUniswapV3Assets(ctx, rpcCtx, client, req.Address)
		if appError != nil {
			return nil, appError
		}
	}
	return s.handleGetDappAssetsResponseData(ctx, rpcCtx, req, dappAssets, dappRewards)
}

func (s *AssetService) GetDapps(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetDappsRequest) (*doom_api.GetDappsResponse_Data, *gerror.AppError) {
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	return &doom_api.GetDappsResponse_Data{List: []string{"aave_v3", "uniswap_v2", "uniswap_v3"}}, nil
}

func (s *AssetService) GetTokenBalances(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetTokenBalancesRequest) (*doom_api.GetTokenBalancesResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	res := &doom_api.GetTokenBalancesResponse_Data{}
	switch req.Chain {
	case "eth":
		chain, err := config.ChainByName(req.Chain)
		if err != nil {
			logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
		}
		client, err := ethclient.DialContext(ctx, chain.Address)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		defer client.Close()
		handleEthChainResult, appError := s.EthService.UpsertUserErc20Tokens(ctx, rpcCtx, client, req.Address, true)
		if appError != nil {
			return nil, appError
		}
		if handleEthChainResult.ToBlock > 0 {
			res.ToBlock = structpb.NewStringValue("0x" + strconv.FormatInt(handleEthChainResult.ToBlock, 16))
		}
		var tokens []*doom_api.GetTokenBalancesResponse_Token
		var unknownTokens []*doom_api.GetTokenBalancesResponse_UnknownToken
		for _, token := range handleEthChainResult.Tokens {
			if token.Type == TokenTypeERC20 {
				one := &doom_api.GetTokenBalancesResponse_Token{
					Balance: token.Balance,
					Address: token.Address,
					Name:    token.Name,
					Symbol:  token.Symbol,
				}
				if token.Price != "" {
					one.Balance = token.BalanceValue
					one.Price = structpb.NewStringValue(token.Price)
				}
				if token.Value != "" {
					one.Value = structpb.NewStringValue(token.Value)
				}
				tokens = append(tokens, one)
			} else if token.Type == TokenTypeError {
				unknownTokens = append(unknownTokens, &doom_api.GetTokenBalancesResponse_UnknownToken{
					Balance: token.Balance,
					Address: token.Address,
				})
			}
		}
		res.Tokens = tokens
		res.UnknownTokens = unknownTokens
	}
	return res, nil
}

func (s *AssetService) GetBalance(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetBalanceRequest) (*doom_api.GetBalanceResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	res := &doom_api.GetBalanceResponse_Data{}
	switch req.Chain {
	case "eth":
		chain, err := config.ChainByName(req.Chain)
		if err != nil {
			logger.Error("bad request", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
		}
		client, err := ethclient.DialContext(ctx, chain.Address)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		defer client.Close()
		contractAddress := common.HexToAddress(req.Address)
		balanceValue, err := client.BalanceAt(ctx, contractAddress, nil)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		ethPrice, _ := s.MarketService.GetLatestSpotPrice(ctx, rpcCtx, &doom_api.GetLatestSpotPriceRequest{
			Symbol:   "ETH",
			BaseCoin: "USDT",
		})
		if ethPrice == nil || ethPrice.Price == "" {
			err = errors.New("eth price not found")
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		price, err := strconv.ParseFloat(ethPrice.Price, 64)
		if err != nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
		balance := eth.Uint256ToFloat64(uint256.MustFromBig(balanceValue), 18)
		res.Balance = eth.FormatFloat64(balance)
		res.Price = ethPrice.Price
		res.Value = eth.FormatFloat64(balance * price)
	}
	return res, nil
}
