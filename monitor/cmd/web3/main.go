package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nextsurfer/monitor/internal/common/eth"
	ethabi "github.com/nextsurfer/monitor/internal/common/eth/abi"
)

func main() {
	ctx := context.Background()
	client, err := ethclient.DialContext(ctx, "wss://mainnet.infura.io/ws/v3/ae1cb3ad3a4542e294f99a5f92be46c9")
	if err != nil {
		panic(err)
	}
	defer client.Close()
	// USDT 0xdAC17F958D2ee523a2206206994597C13D831ec7
	// USDC 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
	// usdtAddress := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
	usdcAddress := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	erc20ABI, _ := ethabi.GetABI(ethabi.ERC20ABI)
	USDT_decimals, _ := eth.ERC20_Decimals(ctx, client, *erc20ABI, "0xdAC17F958D2ee523a2206206994597C13D831ec7")
	log.Printf("USDT_decimals: %d\n", USDT_decimals)
	USDC_decimals, _ := eth.ERC20_Decimals(ctx, client, *erc20ABI, "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	log.Printf("USDC_decimals: %d\n", USDC_decimals)
	transferSig := []byte("Transfer(address,address,uint256)")
	log.Println("sig:", string(transferSig))
	transferSigHash := crypto.Keccak256Hash(transferSig)
	log.Println("transferSigHash:", transferSigHash.Hex())
	if err := filterLogs(ctx, client, usdcAddress, transferSigHash); err != nil {
		panic(err)
	}
}

func FilterLogs(ctx context.Context, client *ethclient.Client, fromBlock, toBlock *big.Int, topics [][]common.Hash, errRegexp *regexp.Regexp, contractAddress common.Address) ([]types.Log, ethereum.FilterQuery, error) {
	addresses := []common.Address{contractAddress}
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Topics:    topics,
		Addresses: addresses,
	}
	log.Println("here1")
	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		log.Println("here2:", err)
		matches := errRegexp.FindStringSubmatch(err.Error())
		if len(matches) == 3 {
			log.Println("here3:", matches)
			from, err := strconv.ParseInt(matches[1], 16, 64)
			if err != nil {
				return nil, ethereum.FilterQuery{}, err
			}
			to, err := strconv.ParseInt(matches[2], 16, 64)
			if err != nil {
				return nil, ethereum.FilterQuery{}, err
			}
			return FilterLogs(ctx, client, big.NewInt(from), big.NewInt(to), topics, errRegexp, contractAddress)
		} else {
			return nil, ethereum.FilterQuery{}, err
		}
	}
	return logs, query, nil
}

func handleFilterLogs(ctx context.Context, client *ethclient.Client, sigHash common.Hash, errRegexp *regexp.Regexp, contractAddress common.Address) error {
	var err error
	topics := [][]common.Hash{{sigHash}, {common.HexToHash("0x5be9a4959308a0d0c7bc0870e319314d8d957dbb")}, {}}
	var from *big.Int
	var to *big.Int
	latestBlockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		return err
	}
	latestBlock := big.NewInt(int64(latestBlockNumber))
	var logs []types.Log
	var query ethereum.FilterQuery
	for {
		logs, query, err = FilterLogs(ctx, client, from, to, topics, errRegexp, contractAddress)
		if err != nil {
			return err
		}
		for _, log := range logs {
			switch log.Topics[0].Hex() {
			case sigHash.Hex():
				// https://etherscan.io/tx/0xfceaecc518ad80dcd988bfc5c4e44e9dd5ca6c58774ae77f3a93ae858932d347
				if strings.EqualFold(log.TxHash.Hex(), "0xfceaecc518ad80dcd988bfc5c4e44e9dd5ca6c58774ae77f3a93ae858932d347") {
					fmt.Printf("log: %#v\n", log)
					dataHex := hex.EncodeToString(log.Data)
					fmt.Printf("data hex: %s\n", dataHex)
					val := new(big.Int)
					val.SetString(dataHex, 16)
					fmt.Printf("data: %v\n", val.Uint64())
					valNum, err := strconv.ParseUint(dataHex, 16, 64)
					if err != nil {
						panic(err)
					}
					fmt.Printf("data value: %d\n", valNum)
					decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)
					fmt.Printf("final data value: %d\n", new(big.Int).Div(val, decimals).Uint64())
				}
			}
		}
		if query.ToBlock != nil && query.ToBlock.Cmp(latestBlock) == -1 {
			from = big.NewInt(query.ToBlock.Int64() + 1)
			to = latestBlock
		} else {
			break
		}
	}
	return nil
}

func filterLogs(ctx context.Context, client *ethclient.Client, contractAddress common.Address, sigHash common.Hash) error {
	// block error regexp
	errRegexp := regexp.MustCompile(`.*0x([[:xdigit:]]+).*0x([[:xdigit:]]+).*`)
	if err := handleFilterLogs(ctx, client, sigHash, errRegexp, contractAddress); err != nil {
		return err
	}
	return nil
}
