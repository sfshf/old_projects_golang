package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/word/internal/pkg/model"
	"gorm.io/gorm"
)

type FavoriteDefinitionDAO struct {
	db *gorm.DB
}

func NewFavoriteDefinitionDAO(db *gorm.DB) *FavoriteDefinitionDAO {
	return &FavoriteDefinitionDAO{db: db}
}

func (obj *FavoriteDefinitionDAO) GetTableName() string {
	return TableNameFavoriteDefinition
}

func (obj *FavoriteDefinitionDAO) GetByID(ctx context.Context, id int64) (*FavoriteDefinition, error) {
	var result FavoriteDefinition
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

func (obj *FavoriteDefinitionDAO) GetByUserIDAndDefinitionID(ctx context.Context, userID, definitionID int64) (*FavoriteDefinition, error) {
	var result FavoriteDefinition
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("definition_id = ?", definitionID).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *FavoriteDefinitionDAO) DeleteByID(ctx context.Context, id int64) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli())
	return conn.Error
}

func (obj *FavoriteDefinitionDAO) RecoverByID(ctx context.Context, id int64) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at > 0").
		Update("deleted_at", 0)
	return conn.Error
}

func (obj *FavoriteDefinitionDAO) Create(ctx context.Context, one *FavoriteDefinition) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *FavoriteDefinitionDAO) GetAllByUserID(ctx context.Context, userID int64) ([]FavoriteDefinition, error) {
	var res []FavoriteDefinition
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Find(&res)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	return res, nil
}
