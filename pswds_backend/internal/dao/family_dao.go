package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type FamilyDAO struct {
	db *gorm.DB
}

func NewFamilyDAO(db *gorm.DB) *FamilyDAO {
	return &FamilyDAO{db: db}
}

func (obj *FamilyDAO) GetTableName() string {
	return TableNameFamily
}

func (obj *FamilyDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *FamilyDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *FamilyDAO) Delete(ctx context.Context, familyID string) error {
	return obj.db.WithContext(ctx).Delete(&Family{}, `family_id=?`, familyID).Error
}

func (obj *FamilyDAO) GetByFamilyID(ctx context.Context, familyID string) (*Family, error) {
	var result Family
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *FamilyDAO) GetByCreator(ctx context.Context, userID int64) (*Family, error) {
	var result Family
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("created_by = ?", userID).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}
