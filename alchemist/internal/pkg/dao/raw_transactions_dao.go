package dao

import (
	"context"
	"database/sql"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type RawTransactionsDAO struct {
	db *gorm.DB
}

func NewRawTransactionsDAO(db *gorm.DB) *RawTransactionsDAO {
	return &RawTransactionsDAO{db: db}
}

func (obj *RawTransactionsDAO) GetTableName() string {
	return TableNameRawTransaction
}

func (obj *RawTransactionsDAO) Create(ctx context.Context, one *RawTransaction) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *RawTransactionsDAO) GetUnhandledRows(ctx context.Context) (*sql.Rows, error) {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("handled = ?", false).
		Where("deleted_at = 0").
		Rows()
}

func (obj *RawTransactionsDAO) UpdateHandleErrorByID(ctx context.Context, id int64, err string) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("handled = ?", false).
		Where("deleted_at = 0").
		Updates(RawTransaction{
			Handled: true,
			Error:   err,
		}).Error
}

func (obj *RawTransactionsDAO) DeleteHandleByID(ctx context.Context, id int64, err string) error {
	m := RawTransaction{
		Handled:   true,
		DeletedAt: time.Now().UnixMilli(),
	}
	if err != "" {
		m.Error = err
	}
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("handled = ?", false).
		Where("deleted_at = 0").
		Updates(m).Error
}
