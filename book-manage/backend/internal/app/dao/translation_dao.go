package dao

import (
	"context"
	"fmt"
	"time"

	. "github.com/nextsurfer/book-manage-api/internal/app/model"
	"gorm.io/gorm"
)

type TranslationDAO struct {
	db *gorm.DB
}

func NewTranslationDAO(db *gorm.DB) *TranslationDAO {
	return &TranslationDAO{db: db}
}

func (obj *TranslationDAO) GetTableName() string {
	return "translation"
}

func (obj *TranslationDAO) GetFromID(ctx context.Context, id int64) (*Translation, error) {
	var result Translation
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	if result.ID == 0 {
		return nil, fmt.Errorf("can't find item with id %d in table %s", id, obj.GetTableName())
	}
	return &result, nil
}

func (obj *TranslationDAO) GetTranslationsByDefinitionID(ctx context.Context, definitionID int64) ([]Translation, error) {
	var result []Translation

	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("item_type = 'definition'").
		Where("item_id = ?", definitionID).
		Where("deleted_at = 0")

	if err := conn.Find(&result).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	return result, nil
}

func (obj *TranslationDAO) GetTranslationByExampleID(ctx context.Context, exampleID int64) (*Translation, error) {
	var result Translation

	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("item_type = 'example'").
		Where("item_id = ?", exampleID).
		Where("deleted_at = 0")

	if err := conn.First(&result).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		} else {
			return nil, nil
		}
	}

	return &result, nil
}

func (obj *TranslationDAO) GetTranslationsByExampleIDs(ctx context.Context, exampleIDs []int64) ([]Translation, error) {
	var result []Translation

	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("item_type = 'example'").
		Where("item_id IN (?)", exampleIDs).
		Where("deleted_at = 0")

	if err := conn.Find(&result).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	return result, nil
}

func (obj *TranslationDAO) Create(ctx context.Context, translation *Translation) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(translation).Error
}

func (obj *TranslationDAO) Update(ctx context.Context, translation *Translation, selects ...interface{}) error {
	conn := obj.db.WithContext(ctx).Model(translation)
	if len(selects) > 0 {
		conn = conn.Select(selects[0], selects[1:]...)
	}
	conn.Where("deleted_at = 0").Updates(translation)
	if conn.RowsAffected != 1 {
		return fmt.Errorf("update item with id %d failed in table %s ", translation.ID, obj.GetTableName())
	}
	return conn.Error
}

func (obj *TranslationDAO) DeleteByID(ctx context.Context, id int64) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli())
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete item with id %d failed in table %s ", id, obj.GetTableName())
	}
	return conn.Error
}
