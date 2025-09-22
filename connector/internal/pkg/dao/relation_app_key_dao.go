package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/connector/internal/pkg/model"
	"gorm.io/gorm"
)

type RelationAppKeyDAO struct {
	db *gorm.DB
}

func NewRelationAppKeyDAO(db *gorm.DB) *RelationAppKeyDAO {
	return &RelationAppKeyDAO{db: db}
}

func (obj *RelationAppKeyDAO) GetTableName() string {
	return TableNameRelationAppKey
}

func (obj *RelationAppKeyDAO) Create(ctx context.Context, one *RelationAppKey) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *RelationAppKeyDAO) GetByAppWithKeyID(ctx context.Context, app, keyID string) (*RelationAppKey, error) {
	var result RelationAppKey
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

func (obj *RelationAppKeyDAO) GetByAppWithPasswordHash(ctx context.Context, app, passwordHash string) (*RelationAppKey, error) {
	var result RelationAppKey
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("app = ?", app).
		Where("password_hash = ?", passwordHash).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *RelationAppKeyDAO) GetByKeyID(ctx context.Context, keyID string) (*RelationAppKey, error) {
	var result RelationAppKey
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
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

func (obj *RelationAppKeyDAO) UpdatePasswordHashByID(ctx context.Context, id int64, passwordHash string) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("password_hash", passwordHash).Error
}

func (obj *RelationAppKeyDAO) DeleteByAppWithKeyID(ctx context.Context, app, keyID string) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where(`app = ?`, app).
		Where(`key_id = ?`, keyID).
		Where(`deleted_at = 0`).
		Update("deleted_at", time.Now().UnixMilli()).Error
}

func (obj *RelationAppKeyDAO) GetListByApp(ctx context.Context, app string) ([]*RelationAppKey, error) {
	var list []*RelationAppKey
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("app = ?", app).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *RelationAppKeyDAO) GetListWithApp(ctx context.Context, app string) ([]*RelationAppKey, error) {
	var list []*RelationAppKey
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName())
	if app != "" {
		conn = conn.Where("app = ?", app)
	}
	if err := conn.Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *RelationAppKeyDAO) GetAll(ctx context.Context) ([]*RelationAppKey, error) {
	var list []*RelationAppKey
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *RelationAppKeyDAO) GetAllKeyIDs(ctx context.Context) ([]string, error) {
	var list []string
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Pluck("key_id", &list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
