package dao

import (
	"context"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type SubscriptionCountDAO struct {
	db *gorm.DB
}

func NewSubscriptionCountDAO(db *gorm.DB) *SubscriptionCountDAO {
	return &SubscriptionCountDAO{db: db}
}

func (obj *SubscriptionCountDAO) GetTableName() string {
	return TableNameSubscriptionCount
}

func (obj *SubscriptionCountDAO) GetList(ctx context.Context, app string, startTS, endTS int64) ([]*SubscriptionCount, error) {
	var list []*SubscriptionCount
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("app = ?", app).
		Where("time >= ?", startTS).
		Where("time <= ?", endTS).
		Where("deleted_at = 0").
		Order("time DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *SubscriptionCountDAO) Create(ctx context.Context, value interface{}) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(value).Error
}
