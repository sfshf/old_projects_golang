package config

import (
	"context"
	"strings"

	. "github.com/nextsurfer/doom-go/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReputableToken struct {
	Type          string
	Address       string
	Symbol        string
	Name          string
	Decimals      uint8
	BinanceSymbol []string
	OkxInstIds    []string
}

var (
	ExcludedReputableTokens = []string{"LEO", "XAUT", "SAFE", "PYUSD"}                     // 在Binance/Okx平台查不到实时价格的代币Symbol，不在GetTokens等接口返回
	ExcludedOkxInstIds      = []string{"LEO-USDT", "XAUT-USDT", "SAFE-USDT", "PYUSD-USDT"} // Okx平台未供应服务的InstId列表
	ExcludedBinanceSymbols  = []string{}                                                   // Binance平台未供应服务的Binance Symbol列表
	DefaultReputableTokens  []ReputableToken
)

func InReputableTokens(symbolOrAddress string) *ReputableToken {
	symbolOrAddress = strings.TrimSpace(symbolOrAddress)
	if symbolOrAddress == "" {
		return nil
	}
	for _, token := range DefaultReputableTokens {
		if strings.EqualFold(token.Symbol, symbolOrAddress) ||
			strings.EqualFold(token.Address, symbolOrAddress) {
			return &token
		}
	}
	return nil
}

func LoadReputableTokens(ctx context.Context, mongoDB *mongo.Database) error {
	coll := mongoDB.Collection(CollectionName_ReputableTokens)
	cursor, err := coll.Find(ctx, bson.D{}, options.Find().SetBatchSize(200))
	if err != nil {
		return err
	}
	for cursor.Next(ctx) {
		var token ReputableTokens
		if err := cursor.Decode(&token); err != nil {
			return err
		}
		one := ReputableToken{
			Type:          token.Value.Type,
			Address:       token.Key,
			Symbol:        strings.ToUpper(token.Value.Symbol),
			Name:          token.Value.Name,
			Decimals:      token.Value.Decimals,
			BinanceSymbol: token.Value.BinanceSymbols,
			OkxInstIds:    token.Value.OkxInstIds,
		}
		DefaultReputableTokens = append(DefaultReputableTokens, one)
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	return nil
}
