package main

import (
	"context"
	"log"
	"math/big"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nextsurfer/doom-go/internal/common/eth"
	ethabi "github.com/nextsurfer/doom-go/internal/common/eth/abi"
	. "github.com/nextsurfer/doom-go/internal/model"
	"github.com/nextsurfer/doom-go/internal/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func handleLogs(ctx context.Context, wg *sync.WaitGroup, worker *ethclient.Client, erc20ABI *abi.ABI, coll *mongo.Collection, logChannel chan types.Log, sigHash common.Hash) {
	defer wg.Done()
	ethService, err := service.NewSimpleDoomService(nil, coll.Database(), "", "")
	if err != nil {
		log.Fatalln(err)
	}
	for vLog := range logChannel {
		switch vLog.Topics[0].Hex() {
		case sigHash.Hex():
			if vLog.Removed {
				continue
			}
			contractAddress := common.HexToAddress(eth.MixedcaseAddress(vLog.Address.Hex()))
			if _, err := ethService.ERC20Detection(ctx, worker, erc20ABI, contractAddress); err != nil {
				// try again, if it is a specific error
				if errStr := strings.ToLower(err.Error()); strings.Contains(errStr, "429") || strings.Contains(errStr, "too many requests") {
					logChannel <- vLog
				}
				continue
			}
		}
	}
}

func handleTransferLogs(ctx context.Context, client *ethclient.Client, erc20ABI *abi.ABI, coll *mongo.Collection, sigHash common.Hash, fromBlockNumber, toBlockNumber int64, errRegexp *regexp.Regexp) error {
	var err error
	// topics
	topics := [][]common.Hash{{sigHash}, {}, {}}
	fromBlock := big.NewInt(fromBlockNumber)
	toBlock := big.NewInt(toBlockNumber)
	workers, clean, err := eth.GetEthClientWorkers()
	if err != nil {
		return err
	}
	defer clean()
	logChannel := make(chan types.Log, len(workers))
	var wg sync.WaitGroup
	for _, worker := range workers {
		wg.Add(1)
		go handleLogs(ctx, &wg, worker, erc20ABI, coll, logChannel, sigHash)
	}
	// generator
	ethService := service.NewEthService(nil)
	for {
		var query ethereum.FilterQuery
		var logs []types.Log
		logs, query, err = ethService.FilterLogs(ctx, client, fromBlock, toBlock, topics, errRegexp, "")
		if err != nil {
			time.Sleep(30 * time.Second)
			close(logChannel)
			break
		}
		for _, log := range logs {
			logChannel <- log
		}
		log.Printf("one turn, fromBlock: 0x%x, toBlock: 0x%x\n", query.FromBlock.String(), query.ToBlock.String())
		// next turn
		if query.ToBlock.Cmp(toBlock) == 0 {
			time.Sleep(30 * time.Second)
			close(logChannel)
			break
		} else {
			fromBlock = big.NewInt(query.ToBlock.Int64() + 1)
		}
	}
	// update erc20TokenBlockNumber record
	var val string
	if err != nil {
		val = fromBlock.String()
	} else {
		val = toBlock.String()
	}
	if _, dbErr := coll.ReplaceOne(ctx, bson.D{{Key: "key", Value: "erc20TokenBlockNumber"}}, Erc20TokenBlockNumber{
		Key:   "erc20TokenBlockNumber",
		Value: val,
	}, options.Replace().SetUpsert(true)); dbErr != nil {
		log.Printf("replace erc20TokenBlockNumber record fail: %v\n", dbErr)
	}
	wg.Wait()
	return nil
}

func erc20_list(ctx context.Context, coll *mongo.Collection) error {
	var err error
	// erc20TokenBlockNumber
	var erc20TokenBlockNumber uint64
	var erc20TokenBlockNumberM Erc20TokenBlockNumber
	if err := coll.FindOne(ctx, bson.D{{Key: "key", Value: "erc20TokenBlockNumber"}}).Decode(&erc20TokenBlockNumberM); err != nil {
		if err != mongo.ErrNoDocuments {
			return err
		}
	} else {
		erc20TokenBlockNumber, err = strconv.ParseUint(erc20TokenBlockNumberM.Value, 10, 64)
		if err != nil {
			return err
		}
	}
	// startBlockNumber/fromBlockNumber
	client, err := ethclient.DialContext(ctx, "wss://mainnet.infura.io/ws/v3/ae1cb3ad3a4542e294f99a5f92be46c9")
	if err != nil {
		return err
	}
	defer client.Close()
	toBlockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		return err
	}
	fromBlockNumber := toBlockNumber - 7*24*60*(60/12)
	if fromBlockNumber < erc20TokenBlockNumber {
		fromBlockNumber = erc20TokenBlockNumber + 1
	}
	// block error regexp
	errRegexp := regexp.MustCompile(`.*0x([[:xdigit:]]+).*0x([[:xdigit:]]+).*`)
	// erc20ABI
	erc20ABI, err := ethabi.GetABI(ethabi.ERC20ABI)
	if err != nil {
		return err
	}
	// event Transfer signature
	logTransferSig := []byte("Transfer(address,address,uint256)")
	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
	log.Println("fromBlockNumber:", fromBlockNumber, "toBlockNumber:", toBlockNumber)
	if err := handleTransferLogs(ctx, client, erc20ABI, coll, logTransferSigHash, int64(fromBlockNumber), int64(toBlockNumber), errRegexp); err != nil {
		return err
	}
	return nil
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	ctx := context.Background()
	mongodbUri := "mongodb+srv://sheldon:obZZHKYcMhrPIavE@test1.6sj0f.mongodb.net/?retryWrites=true&w=majority&appName=Test1"
	uri, err := url.Parse(mongodbUri)
	if err != nil {
		log.Fatalln(err)
	}
	dbName := uri.Path[1:]
	if dbName == "" {
		dbName = "doom"
	}
	// init mongo client
	mgoCli, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbUri))
	if err != nil {
		log.Fatalln(err)
	}
	if err := mgoCli.Ping(ctx, nil); err != nil {
		log.Fatalln(err)
	}
	log.Println("mongodb connect successfully")
	// init collection
	coll := mgoCli.Database(dbName).Collection(CollectionName_ERC20Tokens)
	if err := erc20_list(ctx, coll); err != nil {
		log.Fatalln(err)
	}
	log.Println("ok")
}
