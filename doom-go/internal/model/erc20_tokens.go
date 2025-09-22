package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	CollectionName_ERC20Tokens = "erc20_tokens"
	TokenTypeERC20             = "ERC20"
	TokenTypeNotERC20          = "NOT ERC20"
	TokenTypeError             = "ERROR"
)

type Erc20TokenBlockNumber struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Key   string             `bson:"key,omitempty"`
	Value string             `bson:"value,omitempty"`
}

type Erc20Tokens struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Key   string             `bson:"key,omitempty"`
	Value Erc20Tokens_Value  `bson:"value,omitempty"`
}

type Erc20Tokens_Value struct {
	Type     string `bson:"type,omitempty"`
	Name     string `bson:"name,omitempty"`
	Symbol   string `bson:"symbol,omitempty"`
	Decimals uint8  `bson:"decimals,omitempty"`
	Priced   bool   `bson:"priced,omitempty"`
	Checked  bool   `bson:"checked,omitempty"`
}
