package service

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/holiman/uint256"
	"github.com/nextsurfer/doom-go/api/response"
	"github.com/nextsurfer/doom-go/internal/common/config"
	"github.com/nextsurfer/doom-go/internal/common/eth"
	ethabi "github.com/nextsurfer/doom-go/internal/common/eth/abi"
	. "github.com/nextsurfer/doom-go/internal/model"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type AaveService struct {
	*DoomService
}

func NewAaveService(DoomService *DoomService) *AaveService {
	return &AaveService{
		DoomService: DoomService,
	}
}

func (s *AaveService) GetAaveV3Assets(ctx context.Context, rpcCtx *rpc.Context, client *ethclient.Client, userAddress string) ([]*DappAsset, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	uiPoolDataProviderV3ContractABI, err := ethabi.GetABI(ethabi.AaveEthereumMainnetUiPoolDataProviderV3ContractABI)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	ts := time.Now()
	userReserves, err := AaveV3_GetUserReservesData(ctx, client, *uiPoolDataProviderV3ContractABI, userAddress)
	StatisticWeb3Call(ctx, ts)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	erc20ABI, err := ethabi.GetABI(ethabi.ERC20ABI)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	poolDataProviderContractABI, err := ethabi.GetABI(ethabi.AaveEthereumMainnetPoolDataProviderContractABI)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var dappAssets []*DappAsset
	var reservesData []ReservesData
	var appError *gerror.AppError
	for _, userReserve := range userReserves {
		if userReserve.ScaledATokenBalance.Cmp(eth.Uint256Zero) == 1 {
			if len(reservesData) == 0 {
				reservesData, appError = s.getReservesData(ctx, rpcCtx, client, uiPoolDataProviderV3ContractABI)
				if appError != nil {
					return nil, appError
				}
			}
			dappAsset, err := s.dappAsset(ctx, client, erc20ABI, poolDataProviderContractABI, userReserve.UnderlyingAsset.Hex(), userReserve.ScaledATokenBalance, false, reservesData)
			if err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			dappAssets = append(dappAssets, dappAsset)
		} else {
			if userReserve.ScaledVariableDebt.Cmp(eth.Uint256Zero) == 1 {
				if len(reservesData) == 0 {
					reservesData, appError = s.getReservesData(ctx, rpcCtx, client, uiPoolDataProviderV3ContractABI)
					if appError != nil {
						return nil, appError
					}
				}
				dappAsset, err := s.dappAsset(ctx, client, erc20ABI, poolDataProviderContractABI, userReserve.UnderlyingAsset.Hex(), userReserve.ScaledVariableDebt, true, reservesData)
				if err != nil {
					logger.Error("internal error", zap.NamedError("appError", err))
					return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
				dappAssets = append(dappAssets, dappAsset)
			}
		}
	}
	return dappAssets, nil
}

func (s *AaveService) getReservesData(ctx context.Context, rpcCtx *rpc.Context, client *ethclient.Client, uiPoolDataProviderV3ContractABI *abi.ABI) ([]ReservesData, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	var reservesData []ReservesData
	// first, from redis
	raw, err := s.RedisClient.Get(ctx, RedisPrefix_UiPoolDataProviderV3_GetReservesData).Result()
	if err != nil {
		if err != redis.Nil {
			logger.Error("internal error", zap.NamedError("appError", err))
			return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		} else {
			// second, from web3
			list, err := AaveV3_GetReservesData(ctx, client, *uiPoolDataProviderV3ContractABI)
			if err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
			if len(list) > 0 {
				data, err := json.Marshal(list)
				if err != nil {
					logger.Error("internal error", zap.NamedError("appError", err))
					return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
				if err := s.RedisClient.Set(ctx, RedisPrefix_UiPoolDataProviderV3_GetReservesData, string(data), 30*time.Minute).Err(); err != nil {
					logger.Error("internal error", zap.NamedError("appError", err))
					return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
				}
			}
			reservesData = list
		}
	} else {
		if len(raw) > 0 {
			if err := json.Unmarshal([]byte(raw), &reservesData); err != nil {
				logger.Error("internal error", zap.NamedError("appError", err))
				return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
			}
		}
	}
	return reservesData, nil
}

// TODO compute balance
func (s *AaveService) dappAsset(ctx context.Context, client *ethclient.Client, erc20ABI, poolDataProviderContractABI *abi.ABI, assetAddress string, balance *uint256.Int, isDebt bool, reservesData []ReservesData) (*DappAsset, error) {
	// get reserve tokens addresses
	ts := time.Now()
	addresses, err := AaveV3_GetReserveTokensAddresses(ctx, client, *poolDataProviderContractABI, assetAddress)
	StatisticWeb3Call(ctx, ts)
	if err != nil {
		return nil, err
	}
	var address string
	if isDebt {
		address = eth.MixedcaseAddress(addresses[1])
	} else {
		address = eth.MixedcaseAddress(addresses[0])
	}
	var index *uint256.Int
	for _, item := range reservesData {
		if strings.EqualFold(item.UnderlyingAsset.Hex(), assetAddress) {
			if isDebt {
				index = item.VariableBorrowIndex
			} else {
				index = item.LiquidityIndex
			}
			break
		}
	}
	if index == nil {
		return nil, errors.New("nil liquidity index")
	}
	// asset symbol
	var symbol string
	if token := config.InReputableTokens(assetAddress); token != nil {
		symbol = token.Symbol
	}
	// address value
	coll := s.MongoDB.Collection(CollectionName_ERC20Tokens)
	var m Erc20Tokens
	ts = time.Now()
	err = coll.FindOne(ctx, bson.D{{Key: "key", Value: address}}).Decode(&m)
	StatisticMongoCall(ctx, ts)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return nil, err
		} else {
			ts := time.Now()
			name, err := eth.ERC20_Name(ctx, client, *erc20ABI, address)
			StatisticWeb3Call(ctx, ts)
			if err != nil {
				return nil, err
			}
			m.Value.Name = name
			ts = time.Now()
			decimals, err := eth.ERC20_Decimals(ctx, client, *erc20ABI, address)
			StatisticWeb3Call(ctx, ts)
			if err != nil {
				return nil, err
			}
			m.Value.Decimals = decimals
		}
	}
	balance = balance.Mul(balance, index)
	balance = balance.Div(balance, eth.Uint256E27)
	res := DappAsset{
		Name:         m.Value.Name,
		TokenAddress: address,
		IsDebt:       isDebt,
		Holdings: []TokenAsset{
			{
				Symbol: symbol,
				Amount: eth.Uint256ToFloat64(balance, m.Value.Decimals),
			},
		},
	}
	return &res, nil
}

// web3 functions ---------------------------------------------------------------------------------------------------------------------------------------

const (
	AaveEthereumMainnetPoolDataProviderContractAddress      = "0x41393e5e337606dc3821075Af65AeE84D7688CBD"
	AaveEthereumMainnetUiPoolDataProviderV3ContractAddress  = "0x194324C9Af7f56E22F1614dD82E18621cb9238E7"
	AaveEthereumMainnetPoolAddressesProviderContractAddress = "0x2f39d218133AFaB8F2B819B1066c7E434Ad94E9e"
)

type UserReservesData struct {
	UnderlyingAsset                 common.Address
	ScaledATokenBalance             *uint256.Int
	UsageAsCollateralEnabledOnUser  bool
	StableBorrowRate                *uint256.Int
	ScaledVariableDebt              *uint256.Int
	PrincipalStableDebt             *uint256.Int
	StableBorrowLastUpdateTimestamp *uint256.Int
}

func AaveV3_GetUserReservesData(ctx context.Context, client *ethclient.Client, uiPoolDataProviderV3ContractABI abi.ABI, userAddress string) ([]UserReservesData, error) {
	functionName := "getUserReservesData"
	params := []interface{}{AaveEthereumMainnetPoolAddressesProviderContractAddress, userAddress}
	results, err := eth.CallConstantFunction(ctx, client, uiPoolDataProviderV3ContractABI, AaveEthereumMainnetUiPoolDataProviderV3ContractAddress, functionName, params...)
	if err != nil {
		return nil, err
	}
	list := results[0].([]struct {
		UnderlyingAsset                 common.Address `json:"underlyingAsset"`
		ScaledATokenBalance             *big.Int       `json:"scaledATokenBalance"`
		UsageAsCollateralEnabledOnUser  bool           `json:"usageAsCollateralEnabledOnUser"`
		StableBorrowRate                *big.Int       `json:"stableBorrowRate"`
		ScaledVariableDebt              *big.Int       `json:"scaledVariableDebt"`
		PrincipalStableDebt             *big.Int       `json:"principalStableDebt"`
		StableBorrowLastUpdateTimestamp *big.Int       `json:"stableBorrowLastUpdateTimestamp"`
	})
	var res []UserReservesData
	for _, item := range list {
		userReserve := UserReservesData{
			UnderlyingAsset:                 item.UnderlyingAsset,
			ScaledATokenBalance:             uint256.MustFromBig(item.ScaledATokenBalance),
			UsageAsCollateralEnabledOnUser:  item.UsageAsCollateralEnabledOnUser,
			StableBorrowRate:                uint256.MustFromBig(item.StableBorrowRate),
			ScaledVariableDebt:              uint256.MustFromBig(item.ScaledVariableDebt),
			PrincipalStableDebt:             uint256.MustFromBig(item.PrincipalStableDebt),
			StableBorrowLastUpdateTimestamp: uint256.MustFromBig(item.StableBorrowLastUpdateTimestamp),
		}
		res = append(res, userReserve)
	}
	return res, nil
}

type ReservesData struct {
	UnderlyingAsset     common.Address
	LiquidityIndex      *uint256.Int
	VariableBorrowIndex *uint256.Int
}

const (
	RedisPrefix_UiPoolDataProviderV3_GetReservesData = "Aave::UiPoolDataProviderV3::GetReservesData"
)

func AaveV3_GetReservesData(ctx context.Context, client *ethclient.Client, uiPoolDataProviderV3ContractABI abi.ABI) ([]ReservesData, error) {
	functionName := "getReservesData"
	params := []interface{}{AaveEthereumMainnetPoolAddressesProviderContractAddress}
	results, err := eth.CallConstantFunction(ctx, client, uiPoolDataProviderV3ContractABI, AaveEthereumMainnetUiPoolDataProviderV3ContractAddress, functionName, params...)
	if err != nil {
		return nil, err
	}
	list := results[0].([]struct {
		UnderlyingAsset                common.Address `json:"underlyingAsset"`
		Name                           string         `json:"name"`
		Symbol                         string         `json:"symbol"`
		Decimals                       *big.Int       `json:"decimals"`
		BaseLTVasCollateral            *big.Int       `json:"baseLTVasCollateral"`
		ReserveLiquidationThreshold    *big.Int       `json:"reserveLiquidationThreshold"`
		ReserveLiquidationBonus        *big.Int       `json:"reserveLiquidationBonus"`
		ReserveFactor                  *big.Int       `json:"reserveFactor"`
		UsageAsCollateralEnabled       bool           `json:"usageAsCollateralEnabled"`
		BorrowingEnabled               bool           `json:"borrowingEnabled"`
		StableBorrowRateEnabled        bool           `json:"stableBorrowRateEnabled"`
		IsActive                       bool           `json:"isActive"`
		IsFrozen                       bool           `json:"isFrozen"`
		LiquidityIndex                 *big.Int       `json:"liquidityIndex"`
		VariableBorrowIndex            *big.Int       `json:"variableBorrowIndex"`
		LiquidityRate                  *big.Int       `json:"liquidityRate"`
		VariableBorrowRate             *big.Int       `json:"variableBorrowRate"`
		StableBorrowRate               *big.Int       `json:"stableBorrowRate"`
		LastUpdateTimestamp            *big.Int       `json:"lastUpdateTimestamp"`
		ATokenAddress                  common.Address `json:"aTokenAddress"`
		StableDebtTokenAddress         common.Address `json:"stableDebtTokenAddress"`
		VariableDebtTokenAddress       common.Address `json:"variableDebtTokenAddress"`
		InterestRateStrategyAddress    common.Address `json:"interestRateStrategyAddress"`
		AvailableLiquidity             *big.Int       `json:"availableLiquidity"`
		TotalPrincipalStableDebt       *big.Int       `json:"totalPrincipalStableDebt"`
		AverageStableRate              *big.Int       `json:"averageStableRate"`
		StableDebtLastUpdateTimestamp  *big.Int       `json:"stableDebtLastUpdateTimestamp"`
		TotalScaledVariableDebt        *big.Int       `json:"totalScaledVariableDebt"`
		PriceInMarketReferenceCurrency *big.Int       `json:"priceInMarketReferenceCurrency"`
		PriceOracle                    common.Address `json:"priceOracle"`
		VariableRateSlope1             *big.Int       `json:"variableRateSlope1"`
		VariableRateSlope2             *big.Int       `json:"variableRateSlope2"`
		StableRateSlope1               *big.Int       `json:"stableRateSlope1"`
		StableRateSlope2               *big.Int       `json:"stableRateSlope2"`
		BaseStableBorrowRate           *big.Int       `json:"baseStableBorrowRate"`
		BaseVariableBorrowRate         *big.Int       `json:"baseVariableBorrowRate"`
		OptimalUsageRatio              *big.Int       `json:"optimalUsageRatio"`
		IsPaused                       bool           `json:"isPaused"`
		IsSiloedBorrowing              bool           `json:"isSiloedBorrowing"`
		AccruedToTreasury              *big.Int       `json:"accruedToTreasury"`
		Unbacked                       *big.Int       `json:"unbacked"`
		IsolationModeTotalDebt         *big.Int       `json:"isolationModeTotalDebt"`
		FlashLoanEnabled               bool           `json:"flashLoanEnabled"`
		DebtCeiling                    *big.Int       `json:"debtCeiling"`
		DebtCeilingDecimals            *big.Int       `json:"debtCeilingDecimals"`
		EModeCategoryId                uint8          `json:"eModeCategoryId"`
		BorrowCap                      *big.Int       `json:"borrowCap"`
		SupplyCap                      *big.Int       `json:"supplyCap"`
		EModeLtv                       uint16         `json:"eModeLtv"`
		EModeLiquidationThreshold      uint16         `json:"eModeLiquidationThreshold"`
		EModeLiquidationBonus          uint16         `json:"eModeLiquidationBonus"`
		EModePriceSource               common.Address `json:"eModePriceSource"`
		EModeLabel                     string         `json:"eModeLabel"`
		BorrowableInIsolation          bool           `json:"borrowableInIsolation"`
		VirtualAccActive               bool           `json:"virtualAccActive"`
		VirtualUnderlyingBalance       *big.Int       `json:"virtualUnderlyingBalance"`
	})
	var res []ReservesData
	for _, item := range list {
		reservesData := ReservesData{
			UnderlyingAsset:     item.UnderlyingAsset,
			LiquidityIndex:      uint256.MustFromBig(item.LiquidityIndex),
			VariableBorrowIndex: uint256.MustFromBig(item.VariableBorrowIndex),
		}
		res = append(res, reservesData)
	}
	return res, nil
}

func AaveV3_GetReserveTokensAddresses(ctx context.Context, client *ethclient.Client, poolDataProviderContractABI abi.ABI, assetAddress string) ([]string, error) {
	functionName := "getReserveTokensAddresses"
	params := []interface{}{assetAddress}
	results, err := eth.CallConstantFunction(ctx, client, poolDataProviderContractABI, AaveEthereumMainnetPoolDataProviderContractAddress, functionName, params...)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, result := range results {
		addr, ok := result.(common.Address)
		if !ok {
			return nil, eth.ErrUnexpectType
		}
		res = append(res, addr.Hex())
	}
	return res, nil
}
