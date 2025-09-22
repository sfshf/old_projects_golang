package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/connector/internal/pkg/model"
	"gorm.io/gorm"
)

type ApiKeyDAO struct {
	db *gorm.DB
}

func NewApiKeyDAO(db *gorm.DB) *ApiKeyDAO {
	return &ApiKeyDAO{db: db}
}

func (obj *ApiKeyDAO) GetTableName() string {
	return TableNameAPIKey
}

func (obj *ApiKeyDAO) Create(ctx context.Context, one *APIKey) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *ApiKeyDAO) GetByName(ctx context.Context, name string) (*APIKey, error) {
	var result APIKey
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("name = ?", name).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *ApiKeyDAO) GetByID(ctx context.Context, id int64) (*APIKey, error) {
	var result APIKey
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (obj *ApiKeyDAO) GetByAppWithKeyID(ctx context.Context, app, keyID string) (*APIKey, error) {
	var result APIKey
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("app = ?", app).
		Where("key_id = ?", keyID).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *ApiKeyDAO) GetAll(ctx context.Context) ([]*APIKey, error) {
	var list []*APIKey
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *ApiKeyDAO) GetListByApp(ctx context.Context, app string) ([]*APIKey, error) {
	var list []*APIKey
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("app = ?", app).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *ApiKeyDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
}
