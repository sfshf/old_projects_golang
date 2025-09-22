package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type TransactionsProdDAO struct {
	db *gorm.DB
}

func NewTransactionsProdDAO(db *gorm.DB) *TransactionsProdDAO {
	return &TransactionsProdDAO{db: db}
}

func (obj *TransactionsProdDAO) GetTableName() string {
	return TableNameTransactionsProd
}

func (obj *TransactionsProdDAO) Create(ctx context.Context, one *TransactionsProd) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *TransactionsProdDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
