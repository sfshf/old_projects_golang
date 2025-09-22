package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	CollectionName_ProxyTokens = "proxy_tokens"
)

type ProxyToken struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Key   string             `bson:"key,omitempty"`
	Value ProxyToken_Value   `bson:"value,omitempty"`
}

type ProxyToken_Value struct {
	IsProxy     bool   `bson:"isProxy"`
	Type        string `bson:"type,omitempty"`
	Target      string `bson:"target,omitempty"`
	Immutable   bool   `bson:"immutable,omitempty"`
	BlockNumber uint64 `bson:"blockNumber,omitempty"`
}
