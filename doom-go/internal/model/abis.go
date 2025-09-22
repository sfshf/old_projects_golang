package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	CollectionName_ABIs = "abis"
)

type ABIs struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Key   string             `bson:"key,omitempty"`
	Value ABIs_Value         `bson:"value,omitempty"`
}

type ABIs_Value struct {
	ABI           string `bson:"ABI,omitempty"`
	IsProxy       bool   `bson:"isProxy,omitempty"`
	ProxyType     string `bson:"proxyType,omitempty"`
	TargetAddress string `bson:"targetAddress,omitempty"`
	Immutable     bool   `bson:"immutable,omitempty"`
}
