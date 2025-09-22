package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	CollectionName_UserTokenApprovals = "user_token_approvals"
)

type UserTokenApprovals struct {
	ID    primitive.ObjectID       `bson:"_id,omitempty"`
	Key   string                   `bson:"key,omitempty"`
	Value UserTokenApprovals_Value `bson:"value,omitempty"`
}

type UserTokenApprovals_Value struct {
	ToBlock   string                             `bson:"toBlock,omitempty"`
	Approvals []UserTokenApprovals_ValueApproval `bson:"approvals,omitempty"`
}

type UserTokenApprovals_ValueApproval struct {
	Address   string `bson:"address,omitempty"`
	Target    string `bson:"target,omitempty"`
	Allowance string `bson:"allowance,omitempty"`
}
