package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type FreeTrialStateDAO struct {
	db *gorm.DB
}

func NewFreeTrialStateDAO(db *gorm.DB) *FreeTrialStateDAO {
	return &FreeTrialStateDAO{db: db}
}

func (obj *FreeTrialStateDAO) GetTableName() string {
	return TableNameFreeTrialState
}

func (obj *FreeTrialStateDAO) GetByUserIDAndApp(ctx context.Context, userID int64, app string) (*FreeTrialState, error) {
	var result FreeTrialState
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

func (obj *FreeTrialStateDAO) Create(ctx context.Context, one *FreeTrialState) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *FreeTrialStateDAO) UpdateByID(ctx context.Context, id int64, one *FreeTrialState) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Updates(one).Error
	return err
}

func (obj *FreeTrialStateDAO) UpdateColumnByID(ctx context.Context, id int64, column string, value interface{}) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update(column, value).Error
	return err
}

func (obj *FreeTrialStateDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
