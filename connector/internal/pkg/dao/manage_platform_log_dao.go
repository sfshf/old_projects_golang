package dao

import (
	"context"

	. "github.com/nextsurfer/connector/internal/pkg/model"
	"gorm.io/gorm"
)

type ManagePlatformLogDAO struct {
	db *gorm.DB
}

func NewManagePlatformLogDAO(db *gorm.DB) *ManagePlatformLogDAO {
	return &ManagePlatformLogDAO{db: db}
}

func (obj *ManagePlatformLogDAO) GetTableName() string {
	return TableNameManagePlatformLog
}

func (obj *ManagePlatformLogDAO) Create(ctx context.Context, one *ManagePlatformLog) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *ManagePlatformLogDAO) TotalAll(ctx context.Context) (int64, error) {
	var total int64
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Count(&total).Error; err != nil {
		return total, err
	}
	return total, nil
}

func (obj *ManagePlatformLogDAO) GetPagination(ctx context.Context, offset, limit int) ([]*ManagePlatformLog, error) {
	var list []*ManagePlatformLog
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
