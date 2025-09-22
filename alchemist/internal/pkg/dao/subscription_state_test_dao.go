package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type SubscriptionStateTestDAO struct {
	db *gorm.DB
}

func NewSubscriptionStateTestDAO(db *gorm.DB) *SubscriptionStateTestDAO {
	return &SubscriptionStateTestDAO{db: db}
}

func (obj *SubscriptionStateTestDAO) GetTableName() string {
	return TableNameSubscriptionStateTest
}

func (obj *SubscriptionStateTestDAO) Create(ctx context.Context, one *SubscriptionStateTest) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *SubscriptionStateTestDAO) GetByUserID(ctx context.Context, userID int64) (*SubscriptionStateTest, error) {
	var result SubscriptionStateTest
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

func (obj *SubscriptionStateTestDAO) UpdateByID(ctx context.Context, id int64, state *SubscriptionStateTest) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Select("*").
		Omit("id", "created_at", "user_id", "app").
		Updates(state).Error
}

func (obj *SubscriptionStateTestDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
