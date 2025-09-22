package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
)

// contract functions --------------------------------------------------------------------------------------------

func ERC20_BalanceOf(ctx context.Context, client *ethclient.Client, erc20ABI abi.ABI, tokenAddress, ownerAddress string) (*uint256.Int, error) {
	functionName := "balanceOf"
	params := []interface{}{ownerAddress}
	results, err := CallConstantFunction(ctx, client, erc20ABI, tokenAddress, functionName, params...)
	if err != nil {
		return nil, err
	}
	res, ok := results[0].(*big.Int)
	if !ok {
		return nil, ErrUnexpectType
	}
	return uint256.MustFromBig(res), nil
}

func ERC20_Name(ctx context.Context, client *ethclient.Client, erc20ABI abi.ABI, tokenAddress string) (string, error) {
	functionName := "name"
	results, err := CallConstantFunction(ctx, client, erc20ABI, tokenAddress, functionName)
	if err != nil {
		return "", err
	}
	res, ok := results[0].(string)
	if !ok {
		return "", ErrUnexpectType
	}
	return res, nil
}

func ERC20_Symbol(ctx context.Context, client *ethclient.Client, erc20ABI abi.ABI, tokenAddress string) (string, error) {
	functionName := "symbol"
	results, err := CallConstantFunction(ctx, client, erc20ABI, tokenAddress, functionName)
	if err != nil {
		return "", err
	}
	res, ok := results[0].(string)
	if !ok {
		return "", ErrUnexpectType
	}
	return res, nil
}

func ERC20_Decimals(ctx context.Context, client *ethclient.Client, erc20ABI abi.ABI, tokenAddress string) (uint8, error) {
	functionName := "decimals"
	results, err := CallConstantFunction(ctx, client, erc20ABI, tokenAddress, functionName)
	if err != nil {
		return 0, err
	}
	res, ok := results[0].(uint8)
	if !ok {
		return 0, ErrUnexpectType
	}
	return res, nil
}

func ERC20_Allowance(ctx context.Context, client *ethclient.Client, erc20ABI abi.ABI, tokenAddress, tokenOwner, spender string) (*uint256.Int, error) {
	functionName := "allowance"
	params := []interface{}{tokenOwner, spender}
	results, err := CallConstantFunction(ctx, client, erc20ABI, tokenAddress, functionName, params...)
	if err != nil {
		return nil, err
	}
	res, ok := results[0].(*big.Int)
	if !ok {
		return nil, ErrUnexpectType
	}
	return uint256.MustFromBig(res), nil
}
