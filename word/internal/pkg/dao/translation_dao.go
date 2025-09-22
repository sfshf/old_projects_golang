package dao

import (
	"context"

	. "github.com/nextsurfer/word/internal/pkg/model"
	"gorm.io/gorm"
)

type TranslationDAO struct {
	db *gorm.DB
}

func NewTranslationDAO(db *gorm.DB) *TranslationDAO {
	return &TranslationDAO{db: db}
}

func (obj *TranslationDAO) GetTableName() string {
	return TableNameTranslation
}

func (obj *TranslationDAO) RemoveAll(ctx context.Context) error {
	return obj.db.WithContext(ctx).Delete(&Translation{}, "1=1").Error
}

func (obj *TranslationDAO) Create(ctx context.Context, one *Translation) error {
	return obj.db.WithContext(ctx).Create(one).Error
}
