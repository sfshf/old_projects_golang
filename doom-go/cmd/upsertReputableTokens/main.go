package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nextsurfer/doom-go/internal/common/eth"
	ethabi "github.com/nextsurfer/doom-go/internal/common/eth/abi"
	"github.com/nextsurfer/doom-go/internal/common/simplehttp"
	. "github.com/nextsurfer/doom-go/internal/model"
	"github.com/nextsurfer/doom-go/internal/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var baseCoins = []string{"USD", "USDT"} // must upper case

// gecko ------------------------------------------------------------------------------------

type GeckoCoin struct {
	ID        string `json:"id"`
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	Platforms struct {
		Ethereum string `json:"ethereum"`
	} `json:"platforms"`
}

func fetchReputableTokensFromCoinGecko(pageSize int) ([]GeckoCoin, error) {
	// fetch from remote
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1", pageSize)
	var respData []GeckoCoin
	resp, err := simplehttp.Get(url, map[string]string{"User-Agent": "curl/8.5.0"}, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	return respData, nil
}

func fetchSpecialFromCoinGecko(ids ...string) ([]GeckoCoin, error) {
	// fetch from remote
	url := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&ids=" + strings.Join(ids, ",")
	var respData []GeckoCoin
	resp, err := simplehttp.Get(url, map[string]string{"User-Agent": "curl/8.5.0"}, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	return respData, nil
}

func fetchAllCoinsFromCoinGecko() ([]GeckoCoin, error) {
	// fetch from remote
	url := "https://api.coingecko.com/api/v3/coins/list?include_platform=true"
	var respData []GeckoCoin
	resp, err := simplehttp.Get(url, map[string]string{"User-Agent": "curl/8.5.0"}, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	return respData, nil
}

// binance ------------------------------------------------------------------------------------

type ExchangeInfoResponse struct {
	Timezone   string `json:"timezone"`
	ServerTime int64  `json:"serverTime"`
	Symbols    []struct {
		Symbol string `json:"symbol"`
	} `json:"symbols"`
}

func binanceExchangeInfo() (*ExchangeInfoResponse, error) {
	// fetch from remote
	url := "https://api.binance.com/api/v3/exchangeInfo"
	var respData ExchangeInfoResponse
	resp, err := simplehttp.Get(url, map[string]string{"User-Agent": "curl/8.5.0"}, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	return &respData, nil
}

// okx ------------------------------------------------------------------------------------

type OkxSpotInstrumentResponse struct {
	Code string              `json:"code"`
	Msg  string              `json:"msg"`
	Data []OkxSpotInstrument `json:"data"`
}

type OkxSpotInstrument struct {
	BaseCcy  string `json:"baseCcy"`
	InstId   string `json:"instId"`
	InstType string `json:"instType"`
	QuoteCcy string `json:"quoteCcy"`
}

func okxSpotInstruments() ([]OkxSpotInstrument, error) {
	// fetch from remote
	url := "https://www.okx.com/api/v5/public/instruments?instType=SPOT"
	var respData OkxSpotInstrumentResponse
	resp, err := simplehttp.Get(url, map[string]string{"User-Agent": "curl/8.5.0"}, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	if respData.Code != "0" {
		return nil, simplehttp.ErrResponseDataStatusNotOK
	}
	return respData.Data, nil
}

// upsert reputable tokens ------------------------------------------------------------------------------------

// 脚本的数据为更高优先级，覆盖掉现在数据库旧数据
func upsertReputableTokens(pageSize int) {
	// first, fetch reputable coins from coingecko
	reputableTokens, err := fetchReputableTokensFromCoinGecko(pageSize)
	if err != nil {
		log.Println(err)
		return
	}
	specials, err := fetchSpecialFromCoinGecko("weth")
	if err != nil {
		log.Println(err)
		return
	}
	// get special coins
	for _, coin := range specials {
		var has bool
		for _, item := range reputableTokens {
			if item.ID == coin.ID {
				has = true
			}
		}
		if !has {
			reputableTokens = append(reputableTokens, coin)
		}
	}
	// second, fetch all coins from coingecko
	allCoins, err := fetchAllCoinsFromCoinGecko()
	if err != nil {
		log.Println(err)
		return
	}
	// third, binance exchange info
	exchangeInfos, err := binanceExchangeInfo()
	if err != nil {
		log.Println(err)
		return
	}
	// forth, okx spot instruments
	okxInstruments, err := okxSpotInstruments()
	if err != nil {
		log.Println(err)
		return
	}
	ctx := context.Background()
	// mongo
	mongodbUri := "mongodb+srv://sheldon:obZZHKYcMhrPIavE@test1.6sj0f.mongodb.net/?retryWrites=true&w=majority&appName=Test1"
	uri, err := url.Parse(mongodbUri)
	if err != nil {
		log.Println(err)
		return
	}
	mgoCli, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbUri))
	if err != nil {
		log.Println(err)
		return
	}
	if err := mgoCli.Ping(ctx, nil); err != nil {
		log.Println(err)
		return
	}
	log.Println("mongodb connect successfully")
	dbName := uri.Path[1:]
	if dbName == "" {
		dbName = "doom"
	}
	mongoDB := mgoCli.Database(dbName)
	reputableTokensCollection := mongoDB.Collection(CollectionName_ReputableTokens)
	erc20TokensCollection := mongoDB.Collection(CollectionName_ERC20Tokens)
	proxyTokensCollection := mongoDB.Collection(CollectionName_ProxyTokens)
	// eth client
	client, err := ethclient.DialContext(ctx, "wss://mainnet.infura.io/ws/v3/ae1cb3ad3a4542e294f99a5f92be46c9")
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Close()
	erc20ABI, err := ethabi.GetABI(ethabi.ERC20ABI)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("length of reputable tokens:", len(reputableTokens))
	ethService, err := service.NewSimpleDoomService(nil, mongoDB, "", "")
	if err != nil {
		log.Fatalln(err)
	}
	for _, coin := range reputableTokens {
		var ethAddress string
		// coin eth address
		for _, item := range allCoins {
			if item.ID == coin.ID {
				ethAddress = eth.MixedcaseAddress(item.Platforms.Ethereum)
				break
			}
		}
		// delete old records, if has
		if ethAddress != "0x0000000000000000000000000000000000000000" {
			if _, err := erc20TokensCollection.DeleteMany(ctx, bson.D{{Key: "key", Value: ethAddress}}); err != nil {
				log.Println(err)
				return
			}
			// EvmProxyDetection内部有查&存逻辑，所以这里要删除
			if _, err := proxyTokensCollection.DeleteMany(ctx, bson.D{{Key: "key", Value: ethAddress}}); err != nil {
				log.Println(err)
				return
			}
		}

		reputableToken := ReputableTokens{
			Key: ethAddress,
			Value: ReputableTokens_Value{
				Type: TokenTypeError, // initial type is ERROR
			},
		}
		var erc20Token *Erc20Tokens
		if ethAddress != "0x0000000000000000000000000000000000000000" {
			erc20Token = &Erc20Tokens{
				Key: ethAddress,
				Value: Erc20Tokens_Value{
					Type: TokenTypeError, // initial type is ERROR
				},
			}
		}

		// is priced
		var priced bool
		if strings.EqualFold(ethAddress, "0xdac17f958d2ee523a2206206994597c13d831ec7") /*USDT*/ ||
			strings.EqualFold(ethAddress, "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2") /*WETH*/ {
			priced = true
		} else {
			for _, symbol := range exchangeInfos.Symbols {
				if us := strings.ToUpper(coin.Symbol); symbol.Symbol == us+"USDT" || symbol.Symbol == us+"USD" {
					reputableToken.Value.BinanceSymbols = append(reputableToken.Value.BinanceSymbols, symbol.Symbol)
				}
			}
			if len(reputableToken.Value.BinanceSymbols) == 0 {
				for _, item := range okxInstruments {
					if strings.EqualFold(coin.Symbol, item.BaseCcy) {
						reputableToken.Value.OkxInstIds = append(reputableToken.Value.OkxInstIds, item.InstId)
					}
				}
			}
			if len(reputableToken.Value.BinanceSymbols) > 0 ||
				len(reputableToken.Value.OkxInstIds) > 0 {
				priced = true
			}
		}
		if erc20Token != nil {
			erc20Token.Value.Priced = priced
			// upsert erc20 infos
			contractAddress := common.HexToAddress(ethAddress)
			if strings.EqualFold(ethAddress, "0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2") {
				reputableToken.Value.Type = TokenTypeERC20
				reputableToken.Value.Name = "Maker"
				reputableToken.Value.Symbol = "MKR"
				reputableToken.Value.Decimals = 18
				erc20Token.Value.Type = TokenTypeERC20
				erc20Token.Value.Name = "Maker"
				erc20Token.Value.Symbol = "MKR"
				erc20Token.Value.Decimals = 18
			} else {
				result, err := ethService.ERC20Detection(ctx, client, erc20ABI, contractAddress)
				if err != nil {
					log.Printf("ERC20Detection fail: %v\n", err)
				} else {
					reputableToken.Value.Type = result.Value.Type
					reputableToken.Value.Name = result.Value.Name
					reputableToken.Value.Symbol = result.Value.Symbol
					reputableToken.Value.Decimals = result.Value.Decimals
					erc20Token.Value.Type = result.Value.Type
					erc20Token.Value.Name = result.Value.Name
					erc20Token.Value.Symbol = result.Value.Symbol
					erc20Token.Value.Decimals = result.Value.Decimals
				}
			}
		} else {
			reputableToken.Value.Name = strings.ToUpper(coin.Name)
			reputableToken.Value.Symbol = strings.ToUpper(coin.Symbol)
		}
		// upsert the record
		if erc20Token != nil {
			if _, err := erc20TokensCollection.ReplaceOne(ctx, bson.D{{Key: "key", Value: ethAddress}}, erc20Token, options.Replace().SetUpsert(true)); err != nil {
				log.Println(err)
				return
			}
		}
		if ethAddress != "0x0000000000000000000000000000000000000000" {
			if _, err := reputableTokensCollection.ReplaceOne(ctx, bson.D{{Key: "key", Value: ethAddress}}, reputableToken, options.Replace().SetUpsert(true)); err != nil {
				log.Println(err)
				return
			}
		} else {
			if _, err := reputableTokensCollection.ReplaceOne(
				ctx,
				bson.D{
					{Key: "key", Value: ethAddress},
					{Key: "value.name", Value: reputableToken.Value.Name},
					{Key: "value.symbol", Value: reputableToken.Value.Symbol},
				},
				reputableToken, options.Replace().SetUpsert(true),
			); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	start := time.Now()
	log.Println("start running:", start)
	upsertReputableTokens(200)
	end := time.Now()
	log.Printf("end running: %v, duration: %s\n", end, end.Sub(start).String())
}
