package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/connector/internal/pkg/model"
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

func (obj *AppConfigDAO) UpdateByApp(ctx context.Context, app string, config string) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where(`app = ?`, app).
		Where(`deleted_at = 0`).
		Update("config", config).Error
}

func (obj *AppConfigDAO) DeleteByApp(ctx context.Context, app string) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where(`app = ?`, app).
		Where(`deleted_at = 0`).
		Update("deleted_at", time.Now().UnixMilli()).Error
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

func (obj *AppConfigDAO) GetByApp(ctx context.Context, app string) (*AppConfig, error) {
	var result AppConfig
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
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

func (obj *AppConfigDAO) GetAllApps(ctx context.Context) ([]string, error) {
	var list []string
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Pluck("app", &list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *AppConfigDAO) GetListByApp(ctx context.Context, app string) ([]*AppConfig, error) {
	var list []*AppConfig
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("app = ?", app).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
