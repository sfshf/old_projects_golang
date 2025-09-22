package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type AppConfigDAO struct {
	db *gorm.DB
}

func NewAppConfigDAO(db *gorm.DB) *AppConfigDAO {
	return &AppConfigDAO{db: db}
}

func (obj *AppConfigDAO) GetTableName() string {
	return TableNameAppConfig
}

func (obj *AppConfigDAO) Create(ctx context.Context, one *AppConfig) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *AppConfigDAO) UpdateByID(ctx context.Context, id int64, config string) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where(`id = ?`, id).
		Where(`deleted_at = 0`).
		Update("config", config).Error
}

func (obj *AppConfigDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where(`id = ?`, id).
		Where(`deleted_at = 0`).
		Update("deleted_at", time.Now().UnixMilli()).Error
}

func (obj *AppConfigDAO) GetByID(ctx context.Context, id int64) (*AppConfig, error) {
	var result AppConfig
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *AppConfigDAO) GetAll(ctx context.Context) ([]*AppConfig, error) {
	var list []*AppConfig
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
