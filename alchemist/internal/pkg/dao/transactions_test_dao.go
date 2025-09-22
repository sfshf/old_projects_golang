package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type TransactionsTestDAO struct {
	db *gorm.DB
}

func NewTransactionsTestDAO(db *gorm.DB) *TransactionsTestDAO {
	return &TransactionsTestDAO{db: db}
}

func (obj *TransactionsTestDAO) GetTableName() string {
	return TableNameTransactionsTest
}

func (obj *TransactionsTestDAO) Create(ctx context.Context, one *TransactionsTest) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *TransactionsTestDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
