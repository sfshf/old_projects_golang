package service

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
	doom_api "github.com/nextsurfer/doom-go/api"
	"github.com/nextsurfer/doom-go/api/response"
	"github.com/nextsurfer/doom-go/internal/common/config"
	"github.com/nextsurfer/doom-go/internal/common/eth"
	ethabi "github.com/nextsurfer/doom-go/internal/common/eth/abi"
	. "github.com/nextsurfer/doom-go/internal/model"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type UniswapService struct {
	*DoomService
}

func NewUniswapService(DoomService *DoomService) *UniswapService {
	return &UniswapService{
		DoomService: DoomService,
	}
}

// uniswap v2 ------------------------------------------------------------------------------------------------

func (s *UniswapService) lpTokenGenerator(ctx context.Context, uniswapColl *mongo.Collection, userTokens *UpsertUserErc20TokensResult, lpTokenChannel chan<- UniswapTokens, errChannel chan<- error) {
	defer close(lpTokenChannel)
	defer func() {
		if err := recover(); err != nil {
			errChannel <- fmt.Errorf("%v", err)
		}
	}()
	var cnt int
	arr := make([]string, 0, 500)
	lastIdx := len(userTokens.Tokens) - 1
	for idx, token := range userTokens.Tokens {
		if token.Type != TokenTypeERC20 {
			continue
		}
		arr = append(arr, token.Address)
		cnt++
		if cnt == 500 || idx == lastIdx {
			ts := time.Now()
			cursor, err := uniswapColl.Find(ctx,
				bson.D{
					{Key: "key", Value: bson.D{{Key: "$in", Value: arr}}},
					{Key: "value.type", Value: UniswapTokenTypeV2},
				},
				options.Find().SetBatchSize(500),
			)
			StatisticMongoCall(ctx, ts)
			if err != nil {
				errChannel <- err
				return
			}
			for cursor.Next(ctx) {
				var lpToken UniswapTokens
				if err := cursor.Decode(&lpToken); err != nil {
					errChannel <- err
					return
				}
				lpTokenChannel <- lpToken
			}
			if err := cursor.Err(); err != nil {
				errChannel <- err
				return
			}
			cnt = 0
			arr = arr[:0]
		}
	}
}

func (s *UniswapService) generateUniswapV2Asset(ctx context.Context, wg *sync.WaitGroup, workers []*ethclient.Client, ticker *time.Ticker, uniswapTokenChannel <-chan UniswapTokens, errChannel chan<- error, dappAssetChannel chan<- *DappAsset) {
	var cleaner int
	erc20ABI, err := ethabi.GetABI(ethabi.ERC20ABI)
	if err != nil {
		errChannel <- err
		return
	}
	for range ticker.C {
		for _, worker := range workers {
			uniswapToken, ok := <-uniswapTokenChannel
			if ok {
				wg.Add(1)
				go func(cli *ethclient.Client, uniswapToken UniswapTokens) {
					defer wg.Done()
					defer func() {
						if err := recover(); err != nil {
							errChannel <- fmt.Errorf("%v", err)
						}
					}()
					token := uniswapToken.Value.Token0
					var symbol string
					if t0 := config.InReputableTokens(token); t0 == nil || (len(t0.BinanceSymbol) == 0 && len(t0.OkxInstIds) == 0 && !strings.EqualFold(t0.Symbol, "USDT")) {
						token = uniswapToken.Value.Token1
						if t1 := config.InReputableTokens(token); t1 != nil {
							symbol = t1.Symbol
						}
					} else {
						symbol = t0.Symbol
					}
					ts := time.Now()
					balance, err := eth.ERC20_BalanceOf(ctx, cli, *erc20ABI, token, uniswapToken.Key)
					StatisticWeb3Call(ctx, ts)
					if err != nil {
						errChannel <- err
						return
					}
					ts = time.Now()
					decimals, err := eth.ERC20_Decimals(ctx, cli, *erc20ABI, token)
					StatisticWeb3Call(ctx, ts)
					if err != nil {
						errChannel <- err
						return
					}
					tokenAmount := eth.Uint256ToFloat64(balance, decimals)
					dappAssetChannel <- &DappAsset{
						Name:         UniswapTokenTypeV2,
						TokenAddress: uniswapToken.Key,
						IsDebt:       false,
						Holdings: []TokenAsset{
							{
								Symbol: symbol,
								Amount: tokenAmount,
							},
						},
					}
				}(worker, uniswapToken)
			} else {
				time.Sleep(1 * time.Second)
				cleaner++
				if cleaner == len(workers) {
					ticker.Stop()
					close(dappAssetChannel)
					return
				}
			}
		}
	}
}

func (s *UniswapService) GetUniswapV2Assets(ctx context.Context, rpcCtx *rpc.Context, client *ethclient.Client, userAddress string, upsertUserErc20TokensResult *UpsertUserErc20TokensResult) ([]*DappAsset, error) {
	workers, clear, err := eth.GetEthClientWorkers()
	if err != nil {
		return nil, err
	}
	defer clear()
	workersLen := len(workers)
	// first, get token balances
	if upsertUserErc20TokensResult == nil {
		var appError *gerror.AppError
		upsertUserErc20TokensResult, appError = s.EthService.UpsertUserErc20Tokens(ctx, rpcCtx, client, userAddress, false)
		if appError != nil {
			return nil, appError.Error
		}
	}
	// second, filter lp tokens
	lpTokenChannel := make(chan UniswapTokens, workersLen)
	errChannel := make(chan error)
	uniswapColl := s.MongoDB.Collection(CollectionName_UniswapTokens)
	// generator
	go s.lpTokenGenerator(ctx, uniswapColl, upsertUserErc20TokensResult, lpTokenChannel, errChannel)
	// forth, get value
	var wg sync.WaitGroup
	ticker := time.NewTicker(500 * time.Millisecond)
	dappAssetChannel := make(chan *DappAsset, workersLen)
	go s.generateUniswapV2Asset(ctx, &wg, workers, ticker, lpTokenChannel, errChannel, dappAssetChannel)
	var list []*DappAsset
	// aggregator
	wg.Add(1)
	go func() {
		defer wg.Done()
	outer:
		for {
			select {
			case dappAsset, ok := <-dappAssetChannel:
				if !ok {
					break outer
				}
				for idx, item := range dappAsset.Holdings {
					if item.Symbol != "" {
						price, _ := s.MarketService.GetLatestSpotPrice(ctx, rpcCtx, &doom_api.GetLatestSpotPriceRequest{
							Symbol:   item.Symbol,
							BaseCoin: "USDT",
						})
						if price == nil || price.Price == "" {
							continue
						}
						priceValue, _ := strconv.ParseFloat(price.Price, 64)
						value := priceValue * item.Amount
						dappAsset.Holdings[idx].Price = item.Price
						dappAsset.Holdings[idx].Value = value
						dappAsset.TotalValue += value
					}
				}
				list = append(list, dappAsset)
			case e, ok := <-errChannel:
				if ok {
					err = e
				}
			}
		}
	}()
	wg.Wait()
	return list, err
}

const (
	UniswapV3PeripheryContractAddress = "0x778766A928C173BCB49bd31e70ec3b7e12A440Bf"
)

type UniswapAssetInfo struct {
	Pool                                common.Address `json:"pool"`
	Symbol0                             string         `json:"symbol0"`
	Decimals0                           *uint256.Int   `json:"decimals0"`
	Symbol1                             string         `json:"symbol1"`
	Decimals1                           *uint256.Int   `json:"decimals1"`
	Nonce                               *uint256.Int   `json:"nonce"`
	Operator                            common.Address `json:"operator"`
	Token0                              common.Address `json:"token0"`
	Token1                              common.Address `json:"token1"`
	Fee                                 *uint256.Int   `json:"fee"`
	TickLower                           *big.Int       `json:"tickLower"`
	TickUpper                           *big.Int       `json:"tickUpper"`
	UserLiquidity                       *big.Int       `json:"userLiquidity"`
	FeeGrowthInside0LastX128            *uint256.Int   `json:"feeGrowthInside0LastX128"`
	FeeGrowthInside1LastX128            *uint256.Int   `json:"feeGrowthInside1LastX128"`
	TokensOwed0                         *uint256.Int   `json:"tokensOwed0"`
	TokensOwed1                         *uint256.Int   `json:"tokensOwed1"`
	SqrtPriceX96                        *uint256.Int   `json:"sqrtPriceX96"`
	Tick                                *big.Int       `json:"tick"`
	ObservationIndex                    uint16         `json:"observationIndex"`
	ObservationCardinality              uint16         `json:"observationCardinality"`
	ObservationCardinalityNext          uint16         `json:"observationCardinalityNext"`
	FeeProtocol                         uint8          `json:"feeProtocol"`
	Unlocked                            bool           `json:"unlocked"`
	FeeGrowthGlobal0X128                *uint256.Int   `json:"feeGrowthGlobal0X128"`
	FeeGrowthGlobal1X128                *uint256.Int   `json:"feeGrowthGlobal1X128"`
	Token0ProtocolFee                   *uint256.Int   `json:"token0ProtocolFee"`
	Token1ProtocolFee                   *uint256.Int   `json:"token1ProtocolFee"`
	PoolLiquidity                       *uint256.Int   `json:"poolLiquidity"`
	LiquidityGross                      *uint256.Int   `json:"liquidityGross"`
	LiquidityNet                        *big.Int       `json:"liquidityNet"`
	FeeGrowthOutside0X128               *uint256.Int   `json:"feeGrowthOutside0X128"`
	FeeGrowthOutside1X128               *uint256.Int   `json:"feeGrowthOutside1X128"`
	TickCumulativeOutside               *big.Int       `json:"tickCumulativeOutside"`
	SecondsPerLiquidityOutsideX128      *uint256.Int   `json:"secondsPerLiquidityOutsideX128"`
	SecondsOutside                      uint32         `json:"secondsOutside"`
	Initialized                         bool           `json:"initialized"`
	LiquidityGrossLower                 *uint256.Int   `json:"liquidityGrossLower"`
	LiquidityNetLower                   *big.Int       `json:"liquidityNetLower"`
	FeeGrowthOutside0X128Lower          *uint256.Int   `json:"feeGrowthOutside0X128Lower"`
	FeeGrowthOutside1X128Lower          *uint256.Int   `json:"feeGrowthOutside1X128Lower"`
	TickCumulativeOutsideLower          *big.Int       `json:"tickCumulativeOutsideLower"`
	SecondsPerLiquidityOutsideX128Lower *uint256.Int   `json:"secondsPerLiquidityOutsideX128Lower"`
	SecondsOutsideLower                 uint32         `json:"secondsOutsideLower"`
	InitializedLower                    bool           `json:"initializedLower"`
	LiquidityGrossUpper                 *uint256.Int   `json:"liquidityGrossUpper"`
	LiquidityNetUpper                   *big.Int       `json:"liquidityNetUpper"`
	FeeGrowthOutside0X128Upper          *uint256.Int   `json:"feeGrowthOutside0X128Upper"`
	FeeGrowthOutside1X128Upper          *uint256.Int   `json:"feeGrowthOutside1X128Upper"`
	TickCumulativeOutsideUpper          *big.Int       `json:"tickCumulativeOutsideUpper"`
	SecondsPerLiquidityOutsideX128Upper *uint256.Int   `json:"secondsPerLiquidityOutsideX128Upper"`
	SecondsOutsideUpper                 uint32         `json:"secondsOutsideUpper"`
	InitializedUpper                    bool           `json:"initializedUpper"`
	Key                                 [32]uint8      `json:"key"`
	Liquidity                           *uint256.Int   `json:"_liquidity"`
	PoolFeeGrowthInside0LastX128        *uint256.Int   `json:"poolFeeGrowthInside0LastX128"`
	PoolFeeGrowthInside1LastX128        *uint256.Int   `json:"poolFeeGrowthInside1LastX128"`
	PoolTokensOwed0                     *uint256.Int   `json:"poolTokensOwed0"`
	PoolTokensOwed1                     *uint256.Int   `json:"poolTokensOwed1"`
}

func UniswapV3Periphery_GetUserAssetInfos(ctx context.Context, client *ethclient.Client, uniswapV3PeripheryABI abi.ABI, userAddress string) ([]UniswapAssetInfo, error) {
	functionName := "getUserAssetInfos"
	params := []interface{}{userAddress}
	results, err := eth.CallConstantFunction(ctx, client, uniswapV3PeripheryABI, UniswapV3PeripheryContractAddress, functionName, params...)
	if err != nil {
		return nil, err
	}
	res := results[0].([]struct {
		Pool                                common.Address `json:"pool"`
		Symbol0                             string         `json:"symbol0"`
		Decimals0                           *big.Int       `json:"decimals0"`
		Symbol1                             string         `json:"symbol1"`
		Decimals1                           *big.Int       `json:"decimals1"`
		Nonce                               *big.Int       `json:"nonce"`
		Operator                            common.Address `json:"operator"`
		Token0                              common.Address `json:"token0"`
		Token1                              common.Address `json:"token1"`
		Fee                                 *big.Int       `json:"fee"`
		TickLower                           *big.Int       `json:"tickLower"`
		TickUpper                           *big.Int       `json:"tickUpper"`
		UserLiquidity                       *big.Int       `json:"userLiquidity"`
		FeeGrowthInside0LastX128            *big.Int       `json:"feeGrowthInside0LastX128"`
		FeeGrowthInside1LastX128            *big.Int       `json:"feeGrowthInside1LastX128"`
		TokensOwed0                         *big.Int       `json:"tokensOwed0"`
		TokensOwed1                         *big.Int       `json:"tokensOwed1"`
		SqrtPriceX96                        *big.Int       `json:"sqrtPriceX96"`
		Tick                                *big.Int       `json:"tick"`
		ObservationIndex                    uint16         `json:"observationIndex"`
		ObservationCardinality              uint16         `json:"observationCardinality"`
		ObservationCardinalityNext          uint16         `json:"observationCardinalityNext"`
		FeeProtocol                         uint8          `json:"feeProtocol"`
		Unlocked                            bool           `json:"unlocked"`
		FeeGrowthGlobal0X128                *big.Int       `json:"feeGrowthGlobal0X128"`
		FeeGrowthGlobal1X128                *big.Int       `json:"feeGrowthGlobal1X128"`
		Token0ProtocolFee                   *big.Int       `json:"token0ProtocolFee"`
		Token1ProtocolFee                   *big.Int       `json:"token1ProtocolFee"`
		PoolLiquidity                       *big.Int       `json:"poolLiquidity"`
		LiquidityGross                      *big.Int       `json:"liquidityGross"`
		LiquidityNet                        *big.Int       `json:"liquidityNet"`
		FeeGrowthOutside0X128               *big.Int       `json:"feeGrowthOutside0X128"`
		FeeGrowthOutside1X128               *big.Int       `json:"feeGrowthOutside1X128"`
		TickCumulativeOutside               *big.Int       `json:"tickCumulativeOutside"`
		SecondsPerLiquidityOutsideX128      *big.Int       `json:"secondsPerLiquidityOutsideX128"`
		SecondsOutside                      uint32         `json:"secondsOutside"`
		Initialized                         bool           `json:"initialized"`
		LiquidityGrossLower                 *big.Int       `json:"liquidityGrossLower"`
		LiquidityNetLower                   *big.Int       `json:"liquidityNetLower"`
		FeeGrowthOutside0X128Lower          *big.Int       `json:"feeGrowthOutside0X128Lower"`
		FeeGrowthOutside1X128Lower          *big.Int       `json:"feeGrowthOutside1X128Lower"`
		TickCumulativeOutsideLower          *big.Int       `json:"tickCumulativeOutsideLower"`
		SecondsPerLiquidityOutsideX128Lower *big.Int       `json:"secondsPerLiquidityOutsideX128Lower"`
		SecondsOutsideLower                 uint32         `json:"secondsOutsideLower"`
		InitializedLower                    bool           `json:"initializedLower"`
		LiquidityGrossUpper                 *big.Int       `json:"liquidityGrossUpper"`
		LiquidityNetUpper                   *big.Int       `json:"liquidityNetUpper"`
		FeeGrowthOutside0X128Upper          *big.Int       `json:"feeGrowthOutside0X128Upper"`
		FeeGrowthOutside1X128Upper          *big.Int       `json:"feeGrowthOutside1X128Upper"`
		TickCumulativeOutsideUpper          *big.Int       `json:"tickCumulativeOutsideUpper"`
		SecondsPerLiquidityOutsideX128Upper *big.Int       `json:"secondsPerLiquidityOutsideX128Upper"`
		SecondsOutsideUpper                 uint32         `json:"secondsOutsideUpper"`
		InitializedUpper                    bool           `json:"initializedUpper"`
		Key                                 [32]uint8      `json:"key"`
		Liquidity                           *big.Int       `json:"_liquidity"`
		PoolFeeGrowthInside0LastX128        *big.Int       `json:"poolFeeGrowthInside0LastX128"`
		PoolFeeGrowthInside1LastX128        *big.Int       `json:"poolFeeGrowthInside1LastX128"`
		PoolTokensOwed0                     *big.Int       `json:"poolTokensOwed0"`
		PoolTokensOwed1                     *big.Int       `json:"poolTokensOwed1"`
	})
	var list []UniswapAssetInfo
	for _, item := range res {
		list = append(list, UniswapAssetInfo{
			Pool:                                item.Pool,
			Symbol0:                             item.Symbol0,
			Decimals0:                           uint256.MustFromBig(item.Decimals0),
			Symbol1:                             item.Symbol1,
			Decimals1:                           uint256.MustFromBig(item.Decimals1),
			Nonce:                               uint256.MustFromBig(item.Nonce),
			Operator:                            item.Operator,
			Token0:                              item.Token0,
			Token1:                              item.Token1,
			Fee:                                 uint256.MustFromBig(item.Fee),
			TickLower:                           item.TickLower,
			TickUpper:                           item.TickUpper,
			UserLiquidity:                       item.UserLiquidity,
			FeeGrowthInside0LastX128:            uint256.MustFromBig(item.FeeGrowthInside0LastX128),
			FeeGrowthInside1LastX128:            uint256.MustFromBig(item.FeeGrowthInside1LastX128),
			TokensOwed0:                         uint256.MustFromBig(item.TokensOwed0),
			TokensOwed1:                         uint256.MustFromBig(item.TokensOwed1),
			SqrtPriceX96:                        uint256.MustFromBig(item.SqrtPriceX96),
			Tick:                                item.Tick,
			ObservationIndex:                    item.ObservationIndex,
			ObservationCardinality:              item.ObservationCardinality,
			ObservationCardinalityNext:          item.ObservationCardinalityNext,
			FeeProtocol:                         item.FeeProtocol,
			Unlocked:                            item.Unlocked,
			FeeGrowthGlobal0X128:                uint256.MustFromBig(item.FeeGrowthGlobal0X128),
			FeeGrowthGlobal1X128:                uint256.MustFromBig(item.FeeGrowthGlobal1X128),
			Token0ProtocolFee:                   uint256.MustFromBig(item.Token0ProtocolFee),
			Token1ProtocolFee:                   uint256.MustFromBig(item.Token1ProtocolFee),
			PoolLiquidity:                       uint256.MustFromBig(item.PoolLiquidity),
			LiquidityGross:                      uint256.MustFromBig(item.LiquidityGross),
			LiquidityNet:                        item.LiquidityNet,
			FeeGrowthOutside0X128:               uint256.MustFromBig(item.FeeGrowthOutside0X128),
			FeeGrowthOutside1X128:               uint256.MustFromBig(item.FeeGrowthOutside1X128),
			TickCumulativeOutside:               item.TickCumulativeOutside,
			SecondsPerLiquidityOutsideX128:      uint256.MustFromBig(item.SecondsPerLiquidityOutsideX128),
			SecondsOutside:                      item.SecondsOutside,
			Initialized:                         item.Initialized,
			LiquidityGrossLower:                 uint256.MustFromBig(item.LiquidityGrossLower),
			LiquidityNetLower:                   item.LiquidityNetLower,
			FeeGrowthOutside0X128Lower:          uint256.MustFromBig(item.FeeGrowthOutside0X128Lower),
			FeeGrowthOutside1X128Lower:          uint256.MustFromBig(item.FeeGrowthOutside1X128Lower),
			TickCumulativeOutsideLower:          item.TickCumulativeOutsideLower,
			SecondsPerLiquidityOutsideX128Lower: uint256.MustFromBig(item.SecondsPerLiquidityOutsideX128Lower),
			SecondsOutsideLower:                 item.SecondsOutsideLower,
			InitializedLower:                    item.InitializedLower,
			LiquidityGrossUpper:                 uint256.MustFromBig(item.LiquidityGrossUpper),
			LiquidityNetUpper:                   item.LiquidityNetUpper,
			FeeGrowthOutside0X128Upper:          uint256.MustFromBig(item.FeeGrowthOutside0X128Upper),
			FeeGrowthOutside1X128Upper:          uint256.MustFromBig(item.FeeGrowthOutside1X128Upper),
			TickCumulativeOutsideUpper:          item.TickCumulativeOutsideUpper,
			SecondsPerLiquidityOutsideX128Upper: uint256.MustFromBig(item.SecondsPerLiquidityOutsideX128Upper),
			SecondsOutsideUpper:                 item.SecondsOutsideUpper,
			InitializedUpper:                    item.InitializedUpper,
			Key:                                 item.Key,
			Liquidity:                           uint256.MustFromBig(item.Liquidity),
			PoolFeeGrowthInside0LastX128:        uint256.MustFromBig(item.PoolFeeGrowthInside0LastX128),
			PoolFeeGrowthInside1LastX128:        uint256.MustFromBig(item.PoolFeeGrowthInside1LastX128),
			PoolTokensOwed0:                     uint256.MustFromBig(item.PoolTokensOwed0),
			PoolTokensOwed1:                     uint256.MustFromBig(item.PoolTokensOwed1),
		})
	}
	return list, nil
}

var (
	Q96  = uint256.MustFromHex("0x1000000000000000000000000")
	Q128 = uint256.MustFromHex("0x100000000000000000000000000000000")
	Q256 = new(uint256.Int).Exp(Q128, uint256.NewInt(2))
)

func divRoundingUp(x, y *uint256.Int) *uint256.Int {
	part1 := new(uint256.Int).Div(x, y)
	mod := new(uint256.Int).Mod(x, y)
	var part2 *uint256.Int
	if zero := uint256.NewInt(0); mod.Gt(zero) {
		part2 = uint256.NewInt(1)
	} else {
		part2 = zero
	}
	return part1.Add(part1, part2)
}

func mulDivRoundingUp(a, b, denominator *uint256.Int) *uint256.Int {
	res, overflow := new(uint256.Int).MulDivOverflow(a, b, denominator)
	if overflow {
		if !(res.Cmp(uint256.MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")) == -1) {
			panic("result of mulDiv is greater than Uint256_Max")
		}
		res = new(uint256.Int).Add(res, uint256.NewInt(1))
	}
	return res
}

func mulDiv(a, b, denominator *uint256.Int) *uint256.Int {
	res, _ := a.MulDivOverflow(a, b, denominator)
	return res
}

func _getAmount0Delta(
	sqrtRatioAX96 *uint256.Int,
	sqrtRatioBX96 *uint256.Int,
	userLiquidity *uint256.Int,
	roundUp bool,
) *uint256.Int {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) == 1 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}
	numerator1 := new(uint256.Int).Lsh(userLiquidity, 96)
	numerator2 := new(uint256.Int).Sub(sqrtRatioBX96, sqrtRatioAX96)

	if sqrtRatioAX96.Cmp(uint256.NewInt(0)) != 1 {
		panic("sqrtRatioAX96 not greater than 0")
	}

	if roundUp {
		return divRoundingUp(mulDivRoundingUp(numerator1, numerator2, sqrtRatioBX96), sqrtRatioAX96)
	} else {
		return new(uint256.Int).Div(mulDiv(numerator1, numerator2, sqrtRatioBX96), sqrtRatioAX96)
	}
}

func getAmount0Delta(
	sqrtRatioAX96 *uint256.Int,
	sqrtRatioBX96 *uint256.Int,
	userLiquidity *big.Int,
) *uint256.Int {
	if userLiquidity.Cmp(big.NewInt(0)) == -1 {
		res := _getAmount0Delta(sqrtRatioAX96, sqrtRatioBX96, uint256.MustFromBig(userLiquidity.Abs(userLiquidity)), false)
		return res.Neg(res)
	} else {
		return _getAmount0Delta(sqrtRatioAX96, sqrtRatioBX96, uint256.MustFromBig(userLiquidity), true)
	}
}

func _getAmount1Delta(
	sqrtRatioAX96 *uint256.Int,
	sqrtRatioBX96 *uint256.Int,
	userLiquidity *uint256.Int,
	roundUp bool,
) *uint256.Int {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) == 1 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}

	if roundUp {
		return mulDivRoundingUp(userLiquidity, new(uint256.Int).Sub(sqrtRatioBX96, sqrtRatioAX96), Q96)
	} else {
		return mulDiv(userLiquidity, new(uint256.Int).Sub(sqrtRatioBX96, sqrtRatioAX96), Q96)
	}
}

func getAmount1Delta(
	sqrtRatioAX96 *uint256.Int,
	sqrtRatioBX96 *uint256.Int,
	userLiquidity *big.Int,
) *uint256.Int {
	if userLiquidity.Cmp(big.NewInt(0)) == -1 {
		res := _getAmount1Delta(sqrtRatioAX96, sqrtRatioBX96, uint256.MustFromBig(userLiquidity.Abs(userLiquidity)), false)
		return res.Neg(res)
	} else {
		return _getAmount1Delta(sqrtRatioAX96, sqrtRatioBX96, uint256.MustFromBig(userLiquidity), true)
	}
}

func getSqrtRatioAtTick(tick *big.Int) *uint256.Int {
	zero := uint256.NewInt(0)
	absTick := uint256.MustFromBig(new(big.Int).Abs(tick))
	if absTick.Cmp(uint256.NewInt(887272)) == 1 {
		panic("tick not greater than MAX_TICK")
	}
	var ratio *uint256.Int
	if new(uint256.Int).And(absTick, uint256.MustFromHex("0x1")).Cmp(zero) == 0 {
		ratio = uint256.MustFromHex("0x100000000000000000000000000000000")
	} else {
		ratio = uint256.MustFromHex("0xfffcb933bd6fad37aa2d162d1a594001")
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x2")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xfff97272373d413259a46990580e213a")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x4")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xfff2e50f5f656932ef12357cf3c7fdcc")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x8")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xffe5caca7e10e4e61c3624eaa0941cd0")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x10")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xffcb9843d60f6159c9db58835c926644")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x20")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xff973b41fa98c081472e6896dfb254c0")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x40")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xff2ea16466c96a3843ec78b326b52861")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x80")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xfe5dee046a99a2a811c461f1969c3053")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x100")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xfcbe86c7900a88aedcffc83b479aa3a4")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x200")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xf987a7253ac413176f2b074cf7815e54")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x400")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xf3392b0822b70005940c7a398e4b70f3")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x800")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xe7159475a2c29b7443b29c7fa6e889d9")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x1000")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xd097f3bdfd2022b8845ad8f792aa5825")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x2000")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0xa9f746462d870fdf8a65dc1f90e061e5")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x4000")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0x70d869a156d2a1b890bb3df62baf32f7")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x8000")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0x31be135f97d08fd981231505542fcfa6")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x10000")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0x9aa508b5b7a84e1c677de54f3e99bc9")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x20000")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0x5d6af8dedb81196699c329225ee604")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x40000")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0x2216e584f5fa1ea926041bedfe98")).
			Rsh(ratio, 128)
	}
	if !(new(uint256.Int).And(absTick, uint256.MustFromHex("0x80000")).Cmp(zero) == 0) {
		ratio.Mul(ratio, uint256.MustFromHex("0x48a170391f7dc42444e8fa2")).
			Rsh(ratio, 128)
	}
	Uint256Max := uint256.MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	if tick.Cmp(big.NewInt(0)) == 1 {
		ratio = new(uint256.Int).Div(Uint256Max, ratio)
	}
	one := uint256.NewInt(1)
	oneRsh32 := new(uint256.Int).Rsh(one, 32)
	var part2 *uint256.Int
	if new(uint256.Int).Mod(ratio, oneRsh32).Cmp(zero) == 0 {
		part2 = zero
	} else {
		part2 = one
	}
	part1 := new(uint256.Int).Rsh(ratio, 32)
	return new(uint256.Int).Add(part1, part2)
}

func getAmount0AndAmount1(sqrtPriceX96 *uint256.Int, userLiquidity, tick, tickLower, tickUpper *big.Int) (*uint256.Int, *uint256.Int) {
	amount0 := uint256.NewInt(0)
	amount1 := uint256.NewInt(0)
	if tick.Cmp(tickLower) == -1 {
		amount0 = getAmount0Delta(
			getSqrtRatioAtTick(tickLower),
			getSqrtRatioAtTick(tickUpper),
			userLiquidity,
		)
	} else if tick.Cmp(tickUpper) == -1 {
		amount0 = getAmount0Delta(
			sqrtPriceX96,
			getSqrtRatioAtTick(tickUpper),
			userLiquidity,
		)
		amount1 = getAmount1Delta(
			getSqrtRatioAtTick(tickLower),
			sqrtPriceX96,
			userLiquidity,
		)
	} else {
		amount1 = getAmount1Delta(
			getSqrtRatioAtTick(tickLower),
			getSqrtRatioAtTick(tickUpper),
			userLiquidity,
		)
	}
	return amount0, amount1
}

// price = mulDiv(sqrtPriceX96**2, 100**decimalsToken0, Q96**2) / decimalsToken1
func SqrtPriceX96ToPriceRate(sqrtPriceX96 *uint256.Int, decimals0, decimals1 *uint256.Int) float64 {
	nominator1 := new(uint256.Int).Exp(sqrtPriceX96, uint256.NewInt(2))
	decimals0Val := new(uint256.Int).Exp(uint256.NewInt(10), decimals0)
	decimals1Val := new(uint256.Int).Exp(uint256.NewInt(10), decimals1)
	denominator := new(uint256.Int).Exp(Q96, uint256.NewInt(2))
	mulDiv := mulDiv(nominator1, decimals0Val, denominator)
	return mulDiv.Float64() / decimals1Val.Float64()
}

func subIn256(x, y *uint256.Int) *uint256.Int {
	difference := new(uint256.Int).Sub(x, y)
	if difference.Cmp(eth.Uint256Zero) == -1 {
		return new(uint256.Int).Add(Q256, difference)
	} else {
		return difference
	}
}

func getRewards(
	feeGrowthGlobal_0,
	feeGrowthGlobal_1,
	tickLowerFeeGrowthOutside_0,
	tickUpperFeeGrowthOutside_0,
	feeGrowthInsideLast_0,
	tickLowerFeeGrowthOutside_1,
	tickUpperFeeGrowthOutside_1,
	feeGrowthInsideLast_1,
	liquidity,
	decimals0,
	decimals1 *uint256.Int,
	tickLower,
	tickUpper,
	tickCurrent *big.Int,
) (float64, float64) {
	var tickLowerFeeGrowthBelow_0 *uint256.Int
	var tickLowerFeeGrowthBelow_1 *uint256.Int
	var tickUpperFeeGrowthAbove_0 *uint256.Int
	var tickUpperFeeGrowthAbove_1 *uint256.Int

	if tickCurrent.Cmp(tickUpper) != -1 {
		tickUpperFeeGrowthAbove_0 = subIn256(feeGrowthGlobal_0, tickUpperFeeGrowthOutside_0)
		tickUpperFeeGrowthAbove_1 = subIn256(feeGrowthGlobal_1, tickUpperFeeGrowthOutside_1)
	} else {
		tickUpperFeeGrowthAbove_0 = tickUpperFeeGrowthOutside_0
		tickUpperFeeGrowthAbove_1 = tickUpperFeeGrowthOutside_1
	}

	if tickCurrent.Cmp(tickLower) != -1 {
		tickLowerFeeGrowthBelow_0 = tickLowerFeeGrowthOutside_0
		tickLowerFeeGrowthBelow_1 = tickLowerFeeGrowthOutside_1
	} else {
		tickLowerFeeGrowthBelow_0 = subIn256(feeGrowthGlobal_0, tickLowerFeeGrowthOutside_0)
		tickLowerFeeGrowthBelow_1 = subIn256(feeGrowthGlobal_1, tickLowerFeeGrowthOutside_1)
	}

	fr_t1_0 := subIn256(subIn256(feeGrowthGlobal_0, tickLowerFeeGrowthBelow_0), tickUpperFeeGrowthAbove_0)
	fr_t1_1 := subIn256(subIn256(feeGrowthGlobal_1, tickLowerFeeGrowthBelow_1), tickUpperFeeGrowthAbove_1)

	uncollectedFees_0 := new(uint256.Int).Div(new(uint256.Int).Mul(liquidity, subIn256(fr_t1_0, feeGrowthInsideLast_0)), Q128)
	uncollectedFees_1 := new(uint256.Int).Div(new(uint256.Int).Mul(liquidity, subIn256(fr_t1_1, feeGrowthInsideLast_1)), Q128)

	// Decimal adjustment to get final results
	decimals0Val := new(uint256.Int).Exp(uint256.NewInt(10), decimals0)
	decimals1Val := new(uint256.Int).Exp(uint256.NewInt(10), decimals1)
	uncollectedFeesAdjusted_0 := uncollectedFees_0.Float64() / decimals0Val.Float64()
	uncollectedFeesAdjusted_1 := uncollectedFees_1.Float64() / decimals1Val.Float64()
	return uncollectedFeesAdjusted_0, uncollectedFeesAdjusted_1
}

func (s *UniswapService) generateTokenAsset(assetInfo *UniswapAssetInfo, price0, price1 string, price0Value, price1Value, tokenAmount0, tokenAmount1, rate float64) ([]TokenAsset, float64, float64) {
	var tokenAssets []TokenAsset
	var token0Value float64
	var token1Value float64
	if (price0Value != 0 && price1Value != 0) || (price0Value == 0 && price1Value == 0) {
		token0Value = price0Value * tokenAmount0
		token1Value = price1Value * tokenAmount1
		tokenAssets = []TokenAsset{
			{
				Symbol: assetInfo.Symbol0,
				Amount: tokenAmount0,
				Price:  price0,
				Value:  token0Value,
			},
			{
				Symbol: assetInfo.Symbol1,
				Amount: tokenAmount1,
				Price:  price1,
				Value:  token1Value,
			},
		}
	} else if price0Value != 0 && price1Value == 0 {
		token0Value = price0Value * tokenAmount0
		token1Value = (price0Value / rate) * tokenAmount1
		tokenAssets = []TokenAsset{
			{
				Symbol: assetInfo.Symbol0,
				Amount: tokenAmount0,
				Price:  price0,
				Value:  token0Value,
			},
			{
				Symbol: assetInfo.Symbol1,
				Amount: tokenAmount1,
				Price:  eth.FormatFloat64(price0Value / rate),
				Value:  token1Value,
			},
		}
	} else if price0Value == 0 && price1Value != 0 {
		token0Value = (price1Value * rate) * tokenAmount0
		token1Value = price1Value * tokenAmount1
		tokenAssets = []TokenAsset{
			{
				Symbol: assetInfo.Symbol0,
				Amount: tokenAmount0,
				Price:  eth.FormatFloat64(price1Value * rate),
				Value:  token0Value,
			},
			{
				Symbol: assetInfo.Symbol1,
				Amount: tokenAmount1,
				Price:  price1,
				Value:  token1Value,
			},
		}
	}
	return tokenAssets, token0Value, token1Value
}

func (s *UniswapService) generateTokenReward(assetInfo *UniswapAssetInfo, price0, price1 string, price0Value, price1Value, rate float64) ([]*DappReward, float64, float64) {
	var rewards []*DappReward
	// compute reward amounts
	token0Reward, token1Reward := getRewards(
		assetInfo.FeeGrowthGlobal0X128,
		assetInfo.FeeGrowthGlobal1X128,
		assetInfo.FeeGrowthOutside0X128Lower,
		assetInfo.FeeGrowthOutside0X128Upper,
		assetInfo.FeeGrowthInside0LastX128,
		assetInfo.FeeGrowthOutside1X128Lower,
		assetInfo.FeeGrowthOutside1X128Upper,
		assetInfo.FeeGrowthInside1LastX128,
		uint256.MustFromBig(assetInfo.UserLiquidity),
		assetInfo.Decimals0,
		assetInfo.Decimals1,
		assetInfo.TickLower,
		assetInfo.TickUpper,
		assetInfo.Tick,
	)
	var token0Value float64
	var token1Value float64
	if (price0Value != 0 && price1Value != 0) || (price0Value == 0 && price1Value == 0) {
		token0Value = price0Value * token0Reward
		token1Value = price1Value * token1Reward
		rewards = []*DappReward{
			{
				Token:  assetInfo.Symbol0,
				Amount: token0Reward,
				Price:  price0,
				Value:  token0Value,
			},
			{
				Token:  assetInfo.Symbol1,
				Amount: token1Reward,
				Price:  price1,
				Value:  token1Value,
			},
		}
	} else if price0Value != 0 && price1Value == 0 {
		token0Value = price0Value * token0Reward
		token1Value = (price0Value / rate) * token1Reward
		rewards = []*DappReward{
			{
				Token:  assetInfo.Symbol0,
				Amount: token0Reward,
				Price:  price0,
				Value:  token0Value,
			},
			{
				Token:  assetInfo.Symbol1,
				Amount: token1Reward,
				Price:  eth.FormatFloat64(price0Value / rate),
				Value:  token1Value,
			},
		}
	} else if price0Value == 0 && price1Value != 0 {
		token0Value = (price1Value * rate) * token0Reward
		token1Value = price1Value * token1Reward
		rewards = []*DappReward{
			{
				Token:  assetInfo.Symbol0,
				Amount: token0Reward,
				Price:  eth.FormatFloat64(price1Value * rate),
				Value:  token0Value,
			},
			{
				Token:  assetInfo.Symbol1,
				Amount: token1Reward,
				Price:  price1,
				Value:  token1Value,
			},
		}
	}
	return rewards, token0Value, token1Value
}

func (s *UniswapService) generateDappAssetAndReward(ctx context.Context, rpcCtx *rpc.Context, assetInfo *UniswapAssetInfo, tokenAmount0, tokenAmount1, rate float64) (*DappAsset, []*DappReward) {
	// 1. fetch price0, price1
	var price0Value float64
	var price0String string
	price0, _ := s.MarketService.GetLatestSpotPrice(ctx, rpcCtx, &doom_api.GetLatestSpotPriceRequest{
		Symbol:   assetInfo.Symbol0,
		BaseCoin: "USDT",
	})
	if price0 != nil && price0.Price != "" {
		price0String = price0.Price
		price0Value, _ = strconv.ParseFloat(price0.Price, 64)
	}
	var price1Value float64
	var price1String string
	if price0 == nil {
		price1, _ := s.MarketService.GetLatestSpotPrice(ctx, rpcCtx, &doom_api.GetLatestSpotPriceRequest{
			Symbol:   assetInfo.Symbol1,
			BaseCoin: "USDT",
		})
		if price1 != nil && price1.Price != "" {
			price1String = price1.Price
			price1Value, _ = strconv.ParseFloat(price1.Price, 64)
		}
	}
	tokenAssets, token0Value, token1Value := s.generateTokenAsset(assetInfo, price0String, price1String, price0Value, price1Value, tokenAmount0, tokenAmount1, rate)
	tokenRewards, _, _ := s.generateTokenReward(assetInfo, price0String, price1String, price0Value, price1Value, rate)
	return &DappAsset{
		Name:         UniswapTokenTypeV3,
		TokenAddress: assetInfo.Pool.Hex(),
		IsDebt:       false,
		TotalValue:   token0Value + token1Value,
		Holdings:     tokenAssets,
	}, tokenRewards
}

func (s *UniswapService) GetUniswapV3Assets(ctx context.Context, rpcCtx *rpc.Context, client *ethclient.Client, userAddress string) ([]*DappAsset, []*DappReward, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// assets
	uniswapV3PeripheryABI, err := ethabi.GetABI(ethabi.UniswapV3PeripheryABI)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	ts := time.Now()
	userAssetInfos, err := UniswapV3Periphery_GetUserAssetInfos(ctx, client, *uniswapV3PeripheryABI, userAddress)
	StatisticWeb3Call(ctx, ts)
	if err != nil {
		logger.Error("internal error", zap.NamedError("appError", err))
		return nil, nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	var dappAssets []*DappAsset
	var dappRewards []*DappReward
	for _, assetInfo := range userAssetInfos {
		if assetInfo.UserLiquidity.Cmp(eth.BigIntZero) == 0 {
			continue
		}
		// compute rate
		rate := SqrtPriceX96ToPriceRate(assetInfo.SqrtPriceX96, assetInfo.Decimals0, assetInfo.Decimals1) // token1/token0
		amount0, amount1 := getAmount0AndAmount1(
			assetInfo.SqrtPriceX96,
			assetInfo.UserLiquidity,
			assetInfo.Tick,
			assetInfo.TickLower,
			assetInfo.TickUpper,
		)
		// tokenAmount0
		tokenAmount0 := eth.Uint256ToFloat64(amount0, uint8(assetInfo.Decimals0.Uint64()))
		// tokenAmount1
		tokenAmount1 := eth.Uint256ToFloat64(amount1, uint8(assetInfo.Decimals1.Uint64()))
		// compute asset
		// compute reward
		dappAsset, rewards := s.generateDappAssetAndReward(ctx, rpcCtx, &assetInfo, tokenAmount0, tokenAmount1, rate)
		dappAssets = append(dappAssets, dappAsset)
		dappRewards = rewards
	}
	return dappAssets, dappRewards, nil
}
