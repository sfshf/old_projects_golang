package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	CollectionName_UniswapTokens = "uniswap_tokens"
	UniswapTokenTypeV2           = "Uniswap V2"
	UniswapTokenTypeV3           = "Uniswap V3"

	KeyUniswapV2Index       = "uniswapV2Index"
	KeyUniswapV3BlockNumber = "uniswapV3BlockNumber"
)

type UniswapV2Index struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Key   string             `bson:"key,omitempty"`
	Value []int64            `bson:"value,omitempty"`
}

type UniswapV3BlockNumber struct {
	ID    primitive.ObjectID         `bson:"_id,omitempty"`
	Key   string                     `bson:"key,omitempty"`
	Value UniswapV3BlockNumber_Value `bson:"value,omitempty"`
}

type UniswapV3BlockNumber_Value struct {
	Timestamp   int64  `bson:"timestamp,omitempty"`
	BlockNumber string `bson:"blockNumber,omitempty"`
}

type UniswapTokens struct {
	ID    primitive.ObjectID  `bson:"_id,omitempty"`
	Key   string              `bson:"key,omitempty"`
	Index int64               `bson:"index,omitempty"`
	Value UniswapTokens_Value `bson:"value,omitempty"`
}

type UniswapTokens_Value struct {
	Type   string `bson:"type,omitempty"` // Uniswap V2; Uniswap V3
	Token0 string `bson:"token0,omitempty"`
	Token1 string `bson:"token1,omitempty"`
}
