package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nextsurfer/monitor/internal/common/eth"
	"github.com/nextsurfer/monitor/internal/common/notification"
)

type Web3Service struct {
	*MonitorService
}

func NewWeb3Service(ctx context.Context, monitorService *MonitorService) (*Web3Service, error) {
	s := &Web3Service{
		MonitorService: monitorService,
	}
	// monitor transfer to trump
	go s.MonitorTransferToTrump(ctx)
	return s, nil
}

func (s *Web3Service) MonitorTransferToTrump(ctx context.Context) {
	defer func() {
		if msg := recover(); msg != nil {
			log.Println("MonitorTransferToTrump panic:", msg)
			time.Sleep(5 * time.Second)
			// redo
			go s.MonitorTransferToTrump(ctx)
		}
	}()
	client, err := ethclient.DialContext(ctx, "wss://mainnet.infura.io/ws/v3/ae1cb3ad3a4542e294f99a5f92be46c9")
	if err != nil {
		panic(err)
	}
	usdt := eth.MixedcaseAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
	usdtAddress := common.HexToAddress(usdt)
	usdc := eth.MixedcaseAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	usdcAddress := common.HexToAddress(usdc)
	transferSig := []byte("Transfer(address,address,uint256)")
	transferSigHash := crypto.Keccak256Hash(transferSig)
	query := ethereum.FilterQuery{
		Topics: [][]common.Hash{
			{transferSigHash},
			{common.HexToHash("0x5be9a4959308a0d0c7bc0870e319314d8d957dbb")},
			{},
		},
		Addresses: []common.Address{usdtAddress, usdcAddress},
	}
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		panic(err)
	}
	log.Println("MonitorTransferToTrump is running ...")
	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)
	for {
		select {
		case err := <-sub.Err():
			panic(err)
		case vLog := <-logs:
			val := new(big.Int)
			val.SetString(hex.EncodeToString(vLog.Data), 16)
			valNum := new(big.Int).Div(val, decimals).Uint64()
			if valNum > 10000 {
				var b strings.Builder
				fmt.Fprintf(&b, "There are %d ", valNum)
				switch vLog.Address.Hex() {
				case usdt:
					b.WriteString("USDT token ")
				case usdc:
					b.WriteString("USDC token ")
				}
				b.WriteString("transferred from Trump project.\n\nTransaction link: https://etherscan.io/tx/")
				b.WriteString(vLog.TxHash.Hex())
				if err := notification.SendMessage(ctx, s.redisOption.Client, s.MongoDB, b.String()); err != nil {
					log.Println("MonitorTransferToTrump SendMessage error:", err)
				}
			}
		}
	}
}
