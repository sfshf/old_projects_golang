package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const CollectionName_PrivateKeyBackup = "private_key_backup"

type PrivateKeyBackup struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt int64              `bson:"createdAt,omitempty"`
	UpdatedAt int64              `bson:"updatedAt,omitempty"`
	DeletedAt int64              `bson:"deletedAt,omitempty"`
	UserID    int64              `bson:"userID,omitempty"`
	DataID    string             `bson:"dataID,omitempty"`
	Title     string             `bson:"title,omitempty"`
}
