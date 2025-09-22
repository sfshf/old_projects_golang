package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	CollectionName_ReputableTokens = "reputable_tokens"
)

type ReputableTokens struct {
	ID    primitive.ObjectID    `bson:"_id,omitempty"`
	Key   string                `bson:"key,omitempty"`
	Value ReputableTokens_Value `bson:"value,omitempty"`
}

type ReputableTokens_Value struct {
	Type           string   `bson:"type,omitempty"`
	Name           string   `bson:"name,omitempty"`
	Symbol         string   `bson:"symbol,omitempty"`
	Decimals       uint8    `bson:"decimals,omitempty"`
	BinanceSymbols []string `bson:"binanceSymbols,omitempty"`
	OkxInstIds     []string `bson:"okxInstIds,omitempty"`
}
