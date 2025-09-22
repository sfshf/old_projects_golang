package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const CollectionName_UserERC20Tokens = "user_erc20_tokens"

type UserERC20Tokens struct {
	ID    primitive.ObjectID    `bson:"_id,omitempty"`
	Key   string                `bson:"key,omitempty"`
	Value UserERC20Tokens_Value `bson:"value,omitempty"`
}

type UserERC20Tokens_Value struct {
	ToBlock string                       `bson:"toBlock,omitempty"`
	Tokens  []UserERC20Tokens_ValueToken `bson:"tokens,omitempty"`
}

type UserERC20Tokens_ValueToken struct {
	Type     string `bson:"type,omitempty"`
	Balance  string `bson:"balance,omitempty"`
	Address  string `bson:"address,omitempty"`
	Name     string `bson:"name,omitempty"`
	Symbol   string `bson:"symbol,omitempty"`
	Decimals uint8  `bson:"decimals,omitempty"`
	Price    string `bson:"price,omitempty"`
	Value    string `bson:"value,omitempty"`
}
