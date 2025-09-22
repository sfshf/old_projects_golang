package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type SlarkUserDAO struct {
	db *gorm.DB
}

func NewSlarkUserDAO(db *gorm.DB) *SlarkUserDAO {
	return &SlarkUserDAO{db: db}
}

func (obj *SlarkUserDAO) GetTableName() string {
	return TableNameSlarkUser
}

func (obj *SlarkUserDAO) GetByUserID(ctx context.Context, userID int64) (*SlarkUser, error) {
	var result SlarkUser
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *SlarkUserDAO) GetByAppAccountToken(ctx context.Context, token string) (*SlarkUser, error) {
	var result SlarkUser
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("app_account_token = ?", token).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *SlarkUserDAO) Create(ctx context.Context, one *SlarkUser) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *SlarkUserDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}

func (obj *SlarkUserDAO) UpdateColumnByID(ctx context.Context, id int64, column string, value interface{}) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update(column, value).Error
	return err
}
