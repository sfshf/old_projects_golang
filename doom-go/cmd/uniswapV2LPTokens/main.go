package main

import (
	"context"
	"log"
	"math/big"
	"net/url"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nextsurfer/doom-go/internal/common/eth"
	ethabi "github.com/nextsurfer/doom-go/internal/common/eth/abi"
	. "github.com/nextsurfer/doom-go/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func filterLpTokens(ctx context.Context, workers []*ethclient.Client, unhandled *roaring64.Bitmap, v2FactoryABI, pairABI abi.ABI, coll *mongo.Collection) (*roaring64.Bitmap, error) {
	newUnhandled := roaring64.New()
	workersLen := len(workers)
	setp1Ticker := time.NewTicker(500 * time.Millisecond)
	setp2Ticker := time.NewTicker(500 * time.Millisecond)
	setp3Ticker := time.NewTicker(500 * time.Millisecond)
	lpIndexChannel := make(chan int64, workersLen)
	errChannel := make(chan error)
	lpTokenChannel1 := make(chan UniswapTokens, workersLen)
	lpTokenChannel2 := make(chan UniswapTokens, workersLen)
	lpTokenChannel3 := make(chan UniswapTokens, workersLen)
	// generator
	go func() {
		iter := unhandled.Iterator()
		for iter.HasNext() {
			index := int64(iter.Next())
			cnt, err := coll.CountDocuments(ctx, bson.D{{Key: "index", Value: index}, {Key: "value.type", Value: UniswapTokenTypeV2}})
			if err != nil {
				errChannel <- err
				return
			}
			if cnt == 0 {
				lpIndexChannel <- index
			}
		}
		close(lpIndexChannel)
	}()
	var wg sync.WaitGroup
	// step1
	go func(ctx context.Context) {
		var cleaner int
		for range setp1Ticker.C {
			for i, worker := range workers {
				lpIndex, ok := <-lpIndexChannel
				if ok {
					wg.Add(1)
					go func(idx int, cli *ethclient.Client, lpIndex int64) {
						defer wg.Done()
						defer func() {
							if msg := recover(); msg != nil {
								log.Println("step1 recover:", msg)
							}
						}()
						lpAddress, err := UniswapV2_Factory_AllPairs(ctx, cli, v2FactoryABI, uint64(lpIndex))
						if err != nil {
							newUnhandled.Add(uint64(lpIndex))
							return
						}
						lpTokenChannel1 <- UniswapTokens{Key: eth.MixedcaseAddress(lpAddress), Index: lpIndex}
					}(i, worker, lpIndex)
				} else {
					time.Sleep(1 * time.Second)
					cleaner++
					if cleaner == workersLen {
						setp1Ticker.Stop()
						close(lpTokenChannel1)
						return
					}
				}
			}
		}
	}(ctx)
	// step2
	go func(ctx context.Context) {
		var cleaner int
		for range setp2Ticker.C {
			for i, worker := range workers {
				lpToken, ok := <-lpTokenChannel1
				if ok {
					wg.Add(1)
					go func(idx int, cli *ethclient.Client, lpToken UniswapTokens) {
						defer wg.Done()
						defer func() {
							if msg := recover(); msg != nil {
								log.Println("step2 recover:", msg)
							}
						}()
						token0, err := UniswapV2_Pair_Token0(ctx, cli, pairABI, lpToken.Key)
						if err != nil {
							newUnhandled.Add(uint64(lpToken.Index))
							return
						}
						lpToken.Value.Type = UniswapTokenTypeV2
						lpToken.Value.Token0 = eth.MixedcaseAddress(token0)
						lpTokenChannel2 <- lpToken
					}(i, worker, lpToken)
				} else {
					time.Sleep(1 * time.Second)
					cleaner++
					if cleaner == workersLen {
						setp2Ticker.Stop()
						close(lpTokenChannel2)
						return
					}
				}
			}
		}
	}(ctx)
	// step3
	go func(ctx context.Context) {
		var cleaner int
		for range setp3Ticker.C {
			for i, worker := range workers {
				lpToken, ok := <-lpTokenChannel2
				if ok {
					wg.Add(1)
					go func(idx int, cli *ethclient.Client, lpToken UniswapTokens) {
						defer wg.Done()
						defer func() {
							if msg := recover(); msg != nil {
								log.Println("step3 recover:", msg)
							}
						}()
						token1, err := UniswapV2_Pair_Token1(ctx, cli, pairABI, lpToken.Key)
						if err != nil {
							newUnhandled.Add(uint64(lpToken.Index))
							return
						}
						lpToken.Value.Token1 = eth.MixedcaseAddress(token1)
						lpTokenChannel3 <- lpToken
					}(i, worker, lpToken)
				} else {
					time.Sleep(1 * time.Second)
					cleaner++
					if cleaner == workersLen {
						setp3Ticker.Stop()
						close(lpTokenChannel3)
						return
					}
				}
			}
		}
	}(ctx)
	// aggregator
	wg.Add(1)
	go func() {
		defer wg.Done()
		erc20TokensCollection := coll.Database().Collection(CollectionName_ERC20Tokens)
	outer:
		for {
			select {
			case lpToken, ok := <-lpTokenChannel3:
				if ok {
					// write a uniswap_tokens record
					if _, err := coll.ReplaceOne(ctx, bson.D{{Key: "key", Value: lpToken.Key}}, lpToken, options.Replace().SetUpsert(true)); err != nil {
						log.Fatalln(err)
					}
					// write a erc20_tokens record
					erc20Token := Erc20Tokens{
						Key: lpToken.Key,
						Value: Erc20Tokens_Value{
							Type:     TokenTypeERC20,
							Name:     "Uniswap V2",
							Symbol:   "UNI-V2",
							Decimals: 18,
						},
					}
					if _, err := erc20TokensCollection.ReplaceOne(ctx, bson.D{{Key: "key", Value: erc20Token.Key}}, erc20Token, options.Replace().SetUpsert(true)); err != nil {
						log.Fatalln(err)
					}
				} else {
					break outer
				}
			case err, ok := <-errChannel:
				if ok {
					log.Fatalln(err)
				} else {
					break outer
				}
			}

		}
	}()
	wg.Wait()
	return newUnhandled, nil
}

func uniswap_v2_lp_list(ctx context.Context, coll *mongo.Collection) error {
	// fetch uniswap factory allPairsLength
	client, err := ethclient.DialContext(ctx, "wss://mainnet.infura.io/ws/v3/ae1cb3ad3a4542e294f99a5f92be46c9")
	if err != nil {
		return err
	}
	defer client.Close()
	v2FactoryABI, err := ethabi.GetABI(ethabi.UniswapV2FactoryABI)
	if err != nil {
		return err
	}
	allPairsLength, err := UniswapV2_Factory_AllPairsLength(ctx, client, *v2FactoryABI)
	if err != nil {
		return err
	}
	// handle allPairsLength
	unhandled := roaring64.NewBitmap()
	var uniswapV2Index UniswapV2Index
	if err := coll.FindOne(ctx, bson.D{{Key: "key", Value: KeyUniswapV2Index}}).Decode(&uniswapV2Index); err != nil {
		if err != mongo.ErrNoDocuments {
			return err
		} else {
			unhandled.AddRange(0, allPairsLength) // [0, allPairsLength)
		}
	} else {
		for _, idx := range uniswapV2Index.Value {
			unhandled.Add(uint64(idx))
		}
	}
	pairABI, err := ethabi.GetABI(ethabi.UniswapV2PairABI)
	if err != nil {
		return err
	}
	workers, clean, err := eth.GetEthClientWorkers()
	if err != nil {
		return err
	}
	defer clean()
	lastUnhandled := unhandled
	for {
		newUnhandled, err := filterLpTokens(ctx, workers, unhandled, *v2FactoryABI, *pairABI, coll)
		if err != nil {
			return err
		}
		if newUnhandled != nil {
			if newUnhandled.IsEmpty() {
				break
			}
			unhandled = newUnhandled
		}
		if lastUnhandled.Equals(unhandled) {
			break
		}
		lastUnhandled = unhandled
		time.Sleep(1 * time.Second)
	}
	if !lastUnhandled.IsEmpty() {
		var value []int64
		arr := lastUnhandled.ToArray()
		for _, idx := range arr {
			value = append(value, int64(idx))
		}
		uniswapV2Index := UniswapV2Index{
			Key:   KeyUniswapV2Index,
			Value: value,
		}
		if _, err := coll.ReplaceOne(ctx, bson.D{{Key: "key", Value: KeyUniswapV2Index}}, uniswapV2Index, options.Replace().SetUpsert(true)); err != nil {
			return err
		}
	} else {
		if _, err := coll.DeleteOne(ctx, bson.D{{Key: "key", Value: KeyUniswapV2Index}}); err != nil {
			return err
		}
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
	coll := mgoCli.Database(dbName).Collection(CollectionName_UniswapTokens)
	if err := uniswap_v2_lp_list(ctx, coll); err != nil {
		log.Fatalln(err)
	}
	log.Println("ok")
}
