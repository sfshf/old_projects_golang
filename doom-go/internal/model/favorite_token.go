package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const CollectionName_FavoriteToken = "favorite_token"

type FavoriteToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt int64              `bson:"createdAt,omitempty"`
	UpdatedAt int64              `bson:"updatedAt,omitempty"`
	DeletedAt int64              `bson:"deletedAt,omitempty"`
	Symbol    string             `bson:"symbol,omitempty"`
	UserID    int64              `bson:"userID,omitempty"`
}
