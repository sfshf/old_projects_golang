package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type NewUserDiscountStateDAO struct {
	db *gorm.DB
}

func NewNewUserDiscountStateDAO(db *gorm.DB) *NewUserDiscountStateDAO {
	return &NewUserDiscountStateDAO{db: db}
}

func (obj *NewUserDiscountStateDAO) GetTableName() string {
	return TableNameNewUserDiscountState
}

func (obj *NewUserDiscountStateDAO) GetByUserIDAndApp(ctx context.Context, userID int64, app string) (*NewUserDiscountState, error) {
	var result NewUserDiscountState
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("app = ?", app).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *NewUserDiscountStateDAO) GetByUserID(ctx context.Context, userID int64) (*NewUserDiscountState, error) {
	var result NewUserDiscountState
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

func (obj *NewUserDiscountStateDAO) Create(ctx context.Context, one *NewUserDiscountState) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *NewUserDiscountStateDAO) UpdateByID(ctx context.Context, id int64, one *NewUserDiscountState) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Updates(one).Error
	return err
}

func (obj *NewUserDiscountStateDAO) UpdateColumnByID(ctx context.Context, id int64, column string, value interface{}) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update(column, value).Error
	return err
}

func (obj *NewUserDiscountStateDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
