package main

import (
	"context"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
	"github.com/nextsurfer/doom-go/internal/common/eth"
	ethabi "github.com/nextsurfer/doom-go/internal/common/eth/abi"
)

// mainnet: 0xC36442b4a4522E871399CD717aBDD847Ab11FE88
// testnet: 0x1238536071E1c677A632429e3655c799b22cDA52
const NON_FUNGIBLE_POSITION_MANAGER_ADDRESS = "0xC36442b4a4522E871399CD717aBDD847Ab11FE88"

// mainnet: 0x1F98431c8aD98523631AE4a59f267346ea31F984
// testnet: 0x0227628f3F023bb0B980b67D528571c95c6DaC1c
const FACTORY_ADDRESS = "0x1F98431c8aD98523631AE4a59f267346ea31F984"

func NON_FUNGIBLE_POSITION_MANAGER_BalanceOf(ctx context.Context, client *ethclient.Client, abi abi.ABI, userAddress string) (uint64, error) {
	functionName := "balanceOf"
	params := []interface{}{userAddress}
	results, err := eth.CallConstantFunction(ctx, client, abi, NON_FUNGIBLE_POSITION_MANAGER_ADDRESS, functionName, params...)
	if err != nil {
		return 0, err
	}
	return results[0].(*big.Int).Uint64(), nil
}

func NON_FUNGIBLE_POSITION_MANAGER_TokenOfOwnerByIndex(ctx context.Context, client *ethclient.Client, abi abi.ABI, userAddress string, idx uint64) (uint64, error) {
	functionName := "tokenOfOwnerByIndex"
	params := []interface{}{userAddress, idx}
	results, err := eth.CallConstantFunction(ctx, client, abi, NON_FUNGIBLE_POSITION_MANAGER_ADDRESS, functionName, params...)
	if err != nil {
		return 0, err
	}
	return results[0].(*big.Int).Uint64(), nil
}

type PositionsResult struct {
	Nonce                    *big.Int       `json:"nonce"`
	Operator                 common.Address `json:"operator"`
	Token0                   common.Address `json:"token0"`
	Token1                   common.Address `json:"token1"`
	Fee                      *big.Int       `json:"fee"`
	TickLower                *big.Int       `json:"tickLower"`
	TickUpper                *big.Int       `json:"tickUpper"`
	Liquidity                *big.Int       `json:"liquidity"`
	FeeGrowthInside0LastX128 *big.Int       `json:"feeGrowthInside0LastX128"`
	FeeGrowthInside1LastX128 *big.Int       `json:"feeGrowthInside1LastX128"`
	TokensOwed0              *big.Int       `json:"tokensOwed0"`
	TokensOwed1              *big.Int       `json:"tokensOwed1"`
}

func NON_FUNGIBLE_POSITION_MANAGER_Positions(ctx context.Context, client *ethclient.Client, abi abi.ABI, tokenID uint64) (*PositionsResult, error) {
	functionName := "positions"
	params := []interface{}{tokenID}
	results, err := eth.CallConstantFunction(ctx, client, abi, NON_FUNGIBLE_POSITION_MANAGER_ADDRESS, functionName, params...)
	if err != nil {
		return nil, err
	}
	return &PositionsResult{
		Nonce:                    results[0].(*big.Int),
		Operator:                 results[1].(common.Address),
		Token0:                   results[2].(common.Address),
		Token1:                   results[3].(common.Address),
		Fee:                      results[4].(*big.Int),
		TickLower:                results[5].(*big.Int),
		TickUpper:                results[6].(*big.Int),
		Liquidity:                results[7].(*big.Int),
		FeeGrowthInside0LastX128: results[8].(*big.Int),
		FeeGrowthInside1LastX128: results[9].(*big.Int),
		TokensOwed0:              results[10].(*big.Int),
		TokensOwed1:              results[11].(*big.Int),
	}, nil
}

func FACTORY_GetPool(ctx context.Context, client *ethclient.Client, abi abi.ABI, token0, token1 common.Address, fee *big.Int) (common.Address, error) {
	functionName := "getPool"
	params := []interface{}{token0, token1, fee}
	results, err := eth.CallConstantFunction(ctx, client, abi, FACTORY_ADDRESS, functionName, params...)
	if err != nil {
		return common.Address{}, err
	}
	return results[0].(common.Address), nil
}

type Slot0Result struct {
	SqrtPriceX96               *big.Int `json:"sqrtPriceX96"`
	Tick                       *big.Int `json:"tick"`
	ObservationIndex           uint16   `json:"observationIndex"`
	ObservationCardinality     uint16   `json:"observationCardinality"`
	ObservationCardinalityNext uint16   `json:"observationCardinalityNext"`
	FeeProtocol                uint8    `json:"feeProtocol"`
	Unlocked                   bool     `json:"unlocked"`
}

func POOL_Slot0(ctx context.Context, client *ethclient.Client, abi abi.ABI, poolAddress common.Address) (*Slot0Result, error) {
	functionName := "slot0"
	results, err := eth.CallConstantFunction(ctx, client, abi, poolAddress.Hex(), functionName)
	if err != nil {
		return nil, err
	}
	return &Slot0Result{
		SqrtPriceX96:               results[0].(*big.Int),
		Tick:                       results[1].(*big.Int),
		ObservationIndex:           results[2].(uint16),
		ObservationCardinality:     results[3].(uint16),
		ObservationCardinalityNext: results[4].(uint16),
		FeeProtocol:                results[5].(uint8),
		Unlocked:                   results[6].(bool),
	}, nil
}

func POOL_FeeGrowthGlobal0X128(ctx context.Context, client *ethclient.Client, abi abi.ABI, poolAddress common.Address) (*big.Int, error) {
	functionName := "feeGrowthGlobal0X128"
	results, err := eth.CallConstantFunction(ctx, client, abi, poolAddress.Hex(), functionName)
	if err != nil {
		return nil, err
	}
	return results[0].(*big.Int), nil
}

func POOL_FeeGrowthGlobal1X128(ctx context.Context, client *ethclient.Client, abi abi.ABI, poolAddress common.Address) (*big.Int, error) {
	functionName := "feeGrowthGlobal1X128"
	results, err := eth.CallConstantFunction(ctx, client, abi, poolAddress.Hex(), functionName)
	if err != nil {
		return nil, err
	}
	return results[0].(*big.Int), nil
}

func POOL_Liquidity(ctx context.Context, client *ethclient.Client, abi abi.ABI, poolAddress common.Address) (*big.Int, error) {
	functionName := "liquidity"
	results, err := eth.CallConstantFunction(ctx, client, abi, poolAddress.Hex(), functionName)
	if err != nil {
		return nil, err
	}
	return results[0].(*big.Int), nil
}

type PoolPositionsResult struct {
	Liquidity                *big.Int `json:"_liquidity"`
	FeeGrowthInside0LastX128 *big.Int `json:"feeGrowthInside0LastX128"`
	FeeGrowthInside1LastX128 *big.Int `json:"feeGrowthInside1LastX128"`
	TokensOwed0              *big.Int `json:"tokensOwed0"`
	TokensOwed1              *big.Int `json:"tokensOwed1"`
}

func POOL_Positions(ctx context.Context, client *ethclient.Client, poolABI abi.ABI, poolAddress, owner common.Address, tickLower, tickUpper *big.Int) (*PoolPositionsResult, error) {
	addressTy, _ := abi.NewType("address", "string", []abi.ArgumentMarshaling{})
	intTy, _ := abi.NewType("int", "int64", []abi.ArgumentMarshaling{})
	args := abi.Arguments{
		{Type: addressTy},
		{Type: intTy},
		{Type: intTy},
	}
	packed, _ := args.Pack(owner, tickLower, tickUpper)
	key := crypto.Keccak256Hash(packed)
	functionName := "positions"
	params := []interface{}{key}
	results, err := eth.CallConstantFunction(ctx, client, poolABI, poolAddress.Hex(), functionName, params...)
	if err != nil {
		return nil, err
	}
	return &PoolPositionsResult{
		Liquidity:                results[0].(*big.Int),
		FeeGrowthInside0LastX128: results[1].(*big.Int),
		FeeGrowthInside1LastX128: results[2].(*big.Int),
		TokensOwed0:              results[3].(*big.Int),
		TokensOwed1:              results[4].(*big.Int),
	}, nil
}

func ERC20_Symbol(ctx context.Context, client *ethclient.Client, abi abi.ABI, address common.Address) (string, error) {
	functionName := "symbol"
	results, err := eth.CallConstantFunction(ctx, client, abi, address.Hex(), functionName)
	if err != nil {
		return "", err
	}
	return results[0].(string), nil
}

func ERC20_Decimals(ctx context.Context, client *ethclient.Client, abi abi.ABI, address common.Address) (uint8, error) {
	functionName := "decimals"
	results, err := eth.CallConstantFunction(ctx, client, abi, address.Hex(), functionName)
	if err != nil {
		return 0, err
	}
	return results[0].(uint8), nil
}

// func ERC20_BalanceOf(ctx context.Context, client *ethclient.Client, abi abi.ABI, contractAddress common.Address, targetAddress common.Address) (uint8, error) {
// 	functionName := "balanceOf"
// 	params := []interface{}{targetAddress}
// 	results, err := eth.CallConstantFunction(ctx, client, abi, contractAddress.Hex(), functionName, params...)
// 	if err != nil {
// 		return common.Address{}, err
// 	}
// 	return results[0].(uint8), nil
// }

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
		res := divRoundingUp(mulDivRoundingUp(numerator1, numerator2, sqrtRatioBX96), sqrtRatioAX96)
		log.Println("_getAmount0Delta <roundUp>:", res.String())
		return res
	} else {
		res := new(uint256.Int).Div(mulDiv(numerator1, numerator2, sqrtRatioBX96), sqrtRatioAX96)
		log.Println("_getAmount0Delta:", res.String())
		return res
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
		res := mulDivRoundingUp(userLiquidity, new(uint256.Int).Sub(sqrtRatioBX96, sqrtRatioAX96), uint256.MustFromHex("0x1000000000000000000000000"))
		log.Println("_getAmount1Delta <roundUp>:", res.String())
		return res
	} else {
		res := mulDiv(userLiquidity, new(uint256.Int).Sub(sqrtRatioBX96, sqrtRatioAX96), uint256.MustFromHex("0x1000000000000000000000000"))
		log.Println("_getAmount1Delta:", res.String())
		return res
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

func TickToRate(t float64) float64 {
	return math.Pow(math.Ceil(math.Exp(math.Log(1.0001)*t)), 2)
}

func TickToPrice(t float64) float64 {
	return math.Ceil(math.Exp(math.Log(1.0001) * t))
}

func SqrtpToPrice(s float64) float64 {
	return math.Pow(s/math.Exp2(96), 2)
}

func task1() {
	ctx := context.Background()
	// eth client
	client, err := ethclient.DialContext(ctx, "https://eth.llamarpc.com")
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Close()

	// 0x3DdfA8eC3052539b6C9549F12cEA2C295cfF5296
	// 0x392c6a13a6262d2c8a31351c31ba52ac3f413c2d
	// 0x9D72B1250bE1e6364665578cF6B7c4Ee7BB31F93
	userAddress := "0x392c6a13a6262d2c8a31351c31ba52ac3f413c2d"

	nonFungiblePositionManagerABI, err := ethabi.GetABI(ethabi.NonFungiblePositionManagerABI)
	if err != nil {
		log.Println(err)
		return
	}
	factoryABI, err := ethabi.GetABI(ethabi.UniswapV3FactoryABI)
	if err != nil {
		log.Println(err)
		return
	}
	poolABI, err := ethabi.GetABI(ethabi.UniswapV3PoolABI)
	if err != nil {
		log.Println(err)
		return
	}
	erc20ABI, err := ethabi.GetABI(ethabi.ERC20ABI)
	if err != nil {
		log.Println(err)
		return
	}
	nftLength, err := NON_FUNGIBLE_POSITION_MANAGER_BalanceOf(ctx, client, *nonFungiblePositionManagerABI, userAddress)
	if err != nil {
		log.Println(err)
		return
	}
	for i := uint64(0); i < nftLength; i++ {
		tokenID, err := NON_FUNGIBLE_POSITION_MANAGER_TokenOfOwnerByIndex(ctx, client, *nonFungiblePositionManagerABI, userAddress, i)
		if err != nil {
			log.Println(err)
			return
		}
		positionsResult, err := NON_FUNGIBLE_POSITION_MANAGER_Positions(ctx, client, *nonFungiblePositionManagerABI, tokenID)
		if err != nil {
			log.Println(err)
			return
		}
		poolAddress, err := FACTORY_GetPool(ctx, client, *factoryABI, positionsResult.Token0, positionsResult.Token1, positionsResult.Fee)
		if err != nil {
			log.Println(err)
			return
		}
		symbol0, err := ERC20_Symbol(ctx, client, *erc20ABI, positionsResult.Token0)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("symbol0:", symbol0)
		decimals0, err := ERC20_Decimals(ctx, client, *erc20ABI, positionsResult.Token0)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("decimals0:", decimals0)
		symbol1, err := ERC20_Symbol(ctx, client, *erc20ABI, positionsResult.Token1)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("symbol1:", symbol1)
		decimals1, err := ERC20_Decimals(ctx, client, *erc20ABI, positionsResult.Token1)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("decimals1:", decimals1)
		slot0Result, err := POOL_Slot0(ctx, client, *poolABI, poolAddress)
		if err != nil {
			log.Println(err)
			return
		}
		poolPositions, err := POOL_Positions(ctx, client, *poolABI, poolAddress, common.HexToAddress(NON_FUNGIBLE_POSITION_MANAGER_ADDRESS), positionsResult.TickLower, positionsResult.TickUpper)
		if err != nil {
			log.Println(err)
			return
		}
		amount0 := uint256.NewInt(0)
		amount1 := uint256.NewInt(0)
		if positionsResult.Liquidity.Cmp(big.NewInt(0)) == 1 {
			log.Println("tokensOwed0:", mulDiv(
				uint256.MustFromBig(new(big.Int).Sub(poolPositions.FeeGrowthInside0LastX128, positionsResult.FeeGrowthInside0LastX128)),
				uint256.MustFromBig(positionsResult.Liquidity),
				uint256.MustFromHex("0x100000000000000000000000000000000")).String())
			log.Println("tokensOwed1:", mulDiv(
				uint256.MustFromBig(new(big.Int).Sub(poolPositions.FeeGrowthInside1LastX128, positionsResult.FeeGrowthInside1LastX128)),
				uint256.MustFromBig(positionsResult.Liquidity),
				uint256.MustFromHex("0x100000000000000000000000000000000")).String())

			if slot0Result.Tick.Cmp(positionsResult.TickLower) == -1 {
				amount0 = getAmount0Delta(
					getSqrtRatioAtTick(positionsResult.TickLower),
					getSqrtRatioAtTick(positionsResult.TickUpper),
					positionsResult.Liquidity,
				)
			} else if slot0Result.Tick.Cmp(positionsResult.TickUpper) == -1 {
				amount0 = getAmount0Delta(
					uint256.MustFromBig(slot0Result.SqrtPriceX96),
					getSqrtRatioAtTick(positionsResult.TickUpper),
					positionsResult.Liquidity,
				)
				amount1 = getAmount1Delta(
					getSqrtRatioAtTick(positionsResult.TickLower),
					uint256.MustFromBig(slot0Result.SqrtPriceX96),
					positionsResult.Liquidity,
				)
			} else {
				amount1 = getAmount1Delta(
					getSqrtRatioAtTick(positionsResult.TickLower),
					getSqrtRatioAtTick(positionsResult.TickUpper),
					positionsResult.Liquidity,
				)
			}
			log.Printf("pool [%s], token0: [%s], token1: [%s], amount0: %s, amount1: %s",
				poolAddress.Hex(), positionsResult.Token0.Hex(), positionsResult.Token1.Hex(), amount0.String(), amount1.String())
			sqrtPriceX96Value, _ := slot0Result.SqrtPriceX96.Float64()
			log.Println("sqrtPriceX96Value:", SqrtpToPrice(sqrtPriceX96Value))
			tickValue, _ := slot0Result.Tick.Float64()
			tickLowerValue, _ := positionsResult.TickLower.Float64()
			tickUpperValue, _ := positionsResult.TickUpper.Float64()
			log.Println("tickValue:", tickValue)
			log.Println("tickLowerValue:", tickLowerValue)
			log.Println("tickUpperValue:", tickUpperValue)
			rate := TickToPrice(tickValue)
			price0 := 2661.60
			price1 := price0 / rate
			log.Printf("price0 [%.30f], price1: %.30f, rate: %.5f",
				price0, price1, rate)
		}
	}
}

func UniswapV3Periphery_GetUserAssetInfos() (interface{}, error) {
	ctx := context.Background()
	// eth client
	client, err := ethclient.DialContext(ctx, "https://eth.llamarpc.com")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer client.Close()
	peripheryABI, err := ethabi.GetABI(ethabi.UniswapV3PeripheryABI)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	functionName := "getUserAssetInfos"
	params := []interface{}{"0x392c6a13a6262d2c8a31351c31ba52ac3f413c2d"}
	results, err := eth.CallConstantFunction(ctx, client, *peripheryABI, "0xccb574Cd5B23847b2dAF048D5b789E50b284fCFe", functionName, params...)
	if err != nil {
		return nil, err
	}
	log.Printf("%T\n", results[0])
	res := results[0].([]struct {
		Pool                           common.Address `json:"pool"`
		Symbol0                        string         `json:"symbol0"`
		Decimals0                      *big.Int       `json:"decimals0"`
		Symbol1                        string         `json:"symbol1"`
		Decimals1                      *big.Int       `json:"decimals1"`
		Nonce                          *big.Int       `json:"nonce"`
		Operator                       common.Address `json:"operator"`
		Token0                         common.Address `json:"token0"`
		Token1                         common.Address `json:"token1"`
		Fee                            *big.Int       `json:"fee"`
		TickLower                      *big.Int       `json:"tickLower"`
		TickUpper                      *big.Int       `json:"tickUpper"`
		UserLiquidity                  *big.Int       `json:"userLiquidity"`
		FeeGrowthInside0LastX128       *big.Int       `json:"feeGrowthInside0LastX128"`
		FeeGrowthInside1LastX128       *big.Int       `json:"feeGrowthInside1LastX128"`
		TokensOwed0                    *big.Int       `json:"tokensOwed0"`
		TokensOwed1                    *big.Int       `json:"tokensOwed1"`
		SqrtPriceX96                   *big.Int       `json:"sqrtPriceX96"`
		Tick                           *big.Int       `json:"tick"`
		ObservationIndex               uint16         `json:"observationIndex"`
		ObservationCardinality         uint16         `json:"observationCardinality"`
		ObservationCardinalityNext     uint16         `json:"observationCardinalityNext"`
		FeeProtocol                    uint8          `json:"feeProtocol"`
		Unlocked                       bool           `json:"unlocked"`
		FeeGrowthGlobal0X128           *big.Int       `json:"feeGrowthGlobal0X128"`
		FeeGrowthGlobal1X128           *big.Int       `json:"feeGrowthGlobal1X128"`
		Token0ProtocolFee              *big.Int       `json:"token0ProtocolFee"`
		Token1ProtocolFee              *big.Int       `json:"token1ProtocolFee"`
		PoolLiquidity                  *big.Int       `json:"poolLiquidity"`
		LiquidityGross                 *big.Int       `json:"liquidityGross"`
		LiquidityNet                   *big.Int       `json:"liquidityNet"`
		FeeGrowthOutside0X128          *big.Int       `json:"feeGrowthOutside0X128"`
		FeeGrowthOutside1X128          *big.Int       `json:"feeGrowthOutside1X128"`
		TickCumulativeOutside          *big.Int       `json:"tickCumulativeOutside"`
		SecondsPerLiquidityOutsideX128 *big.Int       `json:"secondsPerLiquidityOutsideX128"`
		SecondsOutside                 uint32         `json:"secondsOutside"`
		Initialized                    bool           `json:"initialized"`
		Key                            [32]uint8      `json:"key"`
		Liquidity                      *big.Int       `json:"_liquidity"`
		PoolFeeGrowthInside0LastX128   *big.Int       `json:"poolFeeGrowthInside0LastX128"`
		PoolFeeGrowthInside1LastX128   *big.Int       `json:"poolFeeGrowthInside1LastX128"`
		PoolTokensOwed0                *big.Int       `json:"poolTokensOwed0"`
		PoolTokensOwed1                *big.Int       `json:"poolTokensOwed1"`
	})
	return res, nil
}

var (
	Q96 = uint256.MustFromHex("0x1000000000000000000000000")
)

func SqrtPriceX96ToPriceRate(sqrtPriceX96 *uint256.Int, decimals0, decimals1 *uint256.Int) float64 {
	nominator1 := new(uint256.Int).Exp(sqrtPriceX96, uint256.NewInt(2))
	log.Println("nominator1:", nominator1)
	decimals0Val := new(uint256.Int).Exp(uint256.NewInt(10), decimals0)
	log.Println("decimals0Val:", decimals0Val)
	decimals1Val := new(uint256.Int).Exp(uint256.NewInt(10), decimals1)
	log.Println("decimals1Val:", decimals1Val)
	denominator := new(uint256.Int).Exp(Q96, uint256.NewInt(2))
	mulDiv := mulDiv(nominator1, decimals0Val, denominator)
	log.Println("mulDiv:", mulDiv)
	return mulDiv.Float64() / decimals1Val.Float64()
}

var (
	ZERO = uint256.NewInt(0)
	Q128 = uint256.MustFromHex("0x100000000000000000000000000000000")
	Q256 = new(uint256.Int).Exp(Q128, uint256.NewInt(2))
)

func subIn256(x, y *uint256.Int) *uint256.Int {
	difference := new(uint256.Int).Sub(x, y)
	if difference.Cmp(ZERO) == -1 {
		return new(uint256.Int).Add(Q256, difference)
	} else {
		return difference
	}
}

func getFees(
	feeGrowthGlobal0,
	feeGrowthGlobal1,
	feeGrowth0Low,
	feeGrowth0Hi,
	feeGrowthInside0,
	feeGrowth1Low,
	feeGrowth1Hi,
	feeGrowthInside1,
	liquidity,
	decimals0,
	decimals1 *uint256.Int,
	tickLower,
	tickUpper,
	tickCurrent *big.Int,
) {
	feeGrowthGlobal_0 := feeGrowthGlobal0
	feeGrowthGlobal_1 := feeGrowthGlobal1
	log.Println("feeGrowthGlobal_0:", feeGrowthGlobal_0)
	log.Println("feeGrowthGlobal_1:", feeGrowthGlobal_1)
	tickLowerFeeGrowthOutside_0 := feeGrowth0Low
	tickLowerFeeGrowthOutside_1 := feeGrowth1Low
	log.Println("tickLowerFeeGrowthOutside_0:", tickLowerFeeGrowthOutside_0)
	log.Println("tickLowerFeeGrowthOutside_1:", tickLowerFeeGrowthOutside_1)
	tickUpperFeeGrowthOutside_0 := feeGrowth0Hi
	tickUpperFeeGrowthOutside_1 := feeGrowth1Hi
	log.Println("tickUpperFeeGrowthOutside_0:", tickUpperFeeGrowthOutside_0)
	log.Println("tickUpperFeeGrowthOutside_1:", tickUpperFeeGrowthOutside_1)

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

	log.Println("tickLowerFeeGrowthBelow_0:", tickLowerFeeGrowthBelow_0)
	log.Println("tickLowerFeeGrowthBelow_1:", tickLowerFeeGrowthBelow_1)
	log.Println("tickUpperFeeGrowthAbove_0:", tickUpperFeeGrowthAbove_0)
	log.Println("tickUpperFeeGrowthAbove_1:", tickUpperFeeGrowthAbove_1)

	fr_t1_0 := subIn256(subIn256(feeGrowthGlobal_0, tickLowerFeeGrowthBelow_0), tickUpperFeeGrowthAbove_0)
	fr_t1_1 := subIn256(subIn256(feeGrowthGlobal_1, tickLowerFeeGrowthBelow_1), tickUpperFeeGrowthAbove_1)

	log.Println("fr_t1_0:", fr_t1_0)
	log.Println("fr_t1_1:", fr_t1_1)

	feeGrowthInsideLast_0 := feeGrowthInside0
	feeGrowthInsideLast_1 := feeGrowthInside1

	log.Println("feeGrowthInsideLast_0:", feeGrowthInsideLast_0)
	log.Println("feeGrowthInsideLast_1:", feeGrowthInsideLast_1)

	uncollectedFees_0 := new(uint256.Int).Div(new(uint256.Int).Mul(liquidity, subIn256(fr_t1_0, feeGrowthInsideLast_0)), Q128)
	uncollectedFees_1 := new(uint256.Int).Div(new(uint256.Int).Mul(liquidity, subIn256(fr_t1_1, feeGrowthInsideLast_1)), Q128)

	log.Println("Amount fees token 0 in lowest decimal: ", uncollectedFees_0)
	log.Println("Amount fees token 1 in lowest decimal: ", uncollectedFees_1)

	// Decimal adjustment to get final results
	decimals0Val := new(uint256.Int).Exp(uint256.NewInt(10), decimals0)
	decimals1Val := new(uint256.Int).Exp(uint256.NewInt(10), decimals1)
	uncollectedFeesAdjusted_0 := uncollectedFees_0.Float64() / decimals0Val.Float64()
	uncollectedFeesAdjusted_1 := uncollectedFees_1.Float64() / decimals1Val.Float64()
	log.Println("Amount fees token 0 Human format: ", uncollectedFeesAdjusted_0)
	log.Println("Amount fees token 1 Human format: ", uncollectedFeesAdjusted_1)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	start := time.Now()
	log.Println("start running:", start)
	// UniswapV3Periphery_GetUserAssetInfos()
	// var sqrtPriceX96 = uint256.MustFromDecimal("1580188897131013454712750458832124")
	// log.Println(SqrtPriceX96ToPriceRate(sqrtPriceX96, uint256.NewInt(6), uint256.NewInt(18)))
	getFees(
		uint256.MustFromDecimal("3244923057328358915041562737778710"),
		uint256.MustFromDecimal("1472459342403174043431402000163511217529155"),
		uint256.MustFromDecimal("3129334726876156394658624368360483"),
		uint256.MustFromDecimal("1089412448192672064088500932937488"),
		uint256.MustFromDecimal("115792089237316195423570985008687907853269983677835892949787535233018685921373"),
		uint256.MustFromDecimal("1426187907011064658842676604262922221896484"),
		uint256.MustFromDecimal("733759065674980420847855265548695223301411"),
		uint256.MustFromDecimal("115792089237316195423570985008687907160193372212628071756408211691514122030233"),
		uint256.MustFromDecimal("4614964989348558723"),
		uint256.MustFromDecimal("6"),
		uint256.MustFromDecimal("18"),
		big.NewInt(197250),
		big.NewInt(199920),
		big.NewInt(198031),
	)
	end := time.Now()
	log.Printf("end running: %v, duration: %s\n", end, end.Sub(start).String())
}
