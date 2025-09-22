package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type SubscriptionStateProdDAO struct {
	db *gorm.DB
}

func NewSubscriptionStateProdDAO(db *gorm.DB) *SubscriptionStateProdDAO {
	return &SubscriptionStateProdDAO{db: db}
}

func (obj *SubscriptionStateProdDAO) GetTableName() string {
	return TableNameSubscriptionStateProd
}

func (obj *SubscriptionStateProdDAO) Create(ctx context.Context, one *SubscriptionStateProd) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *SubscriptionStateProdDAO) GetByUserID(ctx context.Context, userID int64) (*SubscriptionStateProd, error) {
	var result SubscriptionStateProd
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

func (obj *SubscriptionStateProdDAO) UpdateByID(ctx context.Context, id int64, state *SubscriptionStateProd) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Select("*").
		Omit("id", "created_at", "user_id", "app").
		Updates(state).Error
}

func (obj *SubscriptionStateProdDAO) CountSubscriptionByApp(ctx context.Context, app string) (int64, error) {
	var cnt int64
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("app = ?", app).
		Where("subscribed = ?", true).
		Where("deleted_at = 0").
		Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}

type Count struct {
	App   string `gorm:"app"`
	Count int64  `gorm:"count"`
}

func (obj *SubscriptionStateProdDAO) CountSubscription(ctx context.Context) ([]*Count, error) {
	var subscriptionCounts []*Count
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Select("app", "COUNT(id) AS count").
		Where("subscribed = ?", true).
		Where("deleted_at = 0").
		Group("app").
		Find(&subscriptionCounts).Error; err != nil {
		return nil, err
	}
	return subscriptionCounts, nil
}

func (obj *SubscriptionStateProdDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
