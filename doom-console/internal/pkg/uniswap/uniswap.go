package uniswap

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nextsurfer/doom-console/internal/pkg/eth"
)

const (
	UniswapV2FactoryAddress = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"
)

// factory functions --------------------------------------------------------------------------------------------

func UniswapV2_Factory_AllPairsLength(ctx context.Context, client *ethclient.Client, v2FactoryABI abi.ABI) (uint64, error) {
	functionName := "allPairsLength"
	results, err := eth.CallConstantFunction(ctx, client, v2FactoryABI, UniswapV2FactoryAddress, functionName)
	if err != nil {
		return 0, err
	}
	res, ok := results[0].(*big.Int)
	if !ok {
		return 0, eth.ErrUnexpectType
	}
	return res.Uint64(), nil
}

func UniswapV2_Factory_AllPairs(ctx context.Context, client *ethclient.Client, v2FactoryABI abi.ABI, idx uint64) (string, error) {
	functionName := "allPairs"
	params := []interface{}{idx}
	results, err := eth.CallConstantFunction(ctx, client, v2FactoryABI, UniswapV2FactoryAddress, functionName, params...)
	if err != nil {
		return "", err
	}
	res, ok := results[0].(common.Address)
	if !ok {
		return "", eth.ErrUnexpectType
	}
	return res.Hex(), nil
}

// pair functions --------------------------------------------------------------------------------------------

func UniswapV2_Pair_Token0(ctx context.Context, client *ethclient.Client, pairABI abi.ABI, lpAddress string) (string, error) {
	functionName := "token0"
	results, err := eth.CallConstantFunction(ctx, client, pairABI, lpAddress, functionName)
	if err != nil {
		return "", err
	}
	res, ok := results[0].(common.Address)
	if !ok {
		return "", eth.ErrUnexpectType
	}
	return res.Hex(), nil
}

func UniswapV2_Pair_Token1(ctx context.Context, client *ethclient.Client, pairABI abi.ABI, lpAddress string) (string, error) {
	functionName := "token1"
	results, err := eth.CallConstantFunction(ctx, client, pairABI, lpAddress, functionName)
	if err != nil {
		return "", err
	}
	res, ok := results[0].(common.Address)
	if !ok {
		return "", eth.ErrUnexpectType
	}
	return res.Hex(), nil
}

// erc20 pair functions --------------------------------------------------------------------------------------------

func UniswapV2_Erc20Pair_TotalSupply(ctx context.Context, client *ethclient.Client, erc20PairABI abi.ABI, lpAddress string) (*big.Int, error) {
	functionName := "totalSupply"
	results, err := eth.CallConstantFunction(ctx, client, erc20PairABI, lpAddress, functionName)
	if err != nil {
		return nil, err
	}
	res, ok := results[0].(*big.Int)
	if !ok {
		return nil, eth.ErrUnexpectType
	}
	return res, nil
}

func UniswapV2_Erc20Pair_BalanceOf(ctx context.Context, client *ethclient.Client, erc20PairABI abi.ABI, lpAddress, userAddress string) (*big.Int, error) {
	functionName := "balanceOf"
	params := []interface{}{userAddress}
	results, err := eth.CallConstantFunction(ctx, client, erc20PairABI, lpAddress, functionName, params...)
	if err != nil {
		return nil, err
	}
	res, ok := results[0].(*big.Int)
	if !ok {
		return nil, eth.ErrUnexpectType
	}
	return res, nil
}

func UniswapV2_Erc20Pair_Decimals(ctx context.Context, client *ethclient.Client, erc20PairABI abi.ABI, lpAddress string) (uint8, error) {
	return 18, nil
}
