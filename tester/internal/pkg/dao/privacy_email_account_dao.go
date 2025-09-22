package dao

import (
	"context"

	. "github.com/nextsurfer/tester/internal/pkg/model"
	"gorm.io/gorm"
)

type PrivacyEmailAccountDAO struct {
	db *gorm.DB
}

func NewPrivacyEmailAccountDAO(db *gorm.DB) *PrivacyEmailAccountDAO {
	return &PrivacyEmailAccountDAO{db: db}
}

func (obj *PrivacyEmailAccountDAO) GetTableName() string {
	return TableNamePrivacyEmailAccount
}

func (obj *PrivacyEmailAccountDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *PrivacyEmailAccountDAO) GetAll(ctx context.Context) ([]*PrivacyEmailAccount, error) {
	var result []*PrivacyEmailAccount
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}
