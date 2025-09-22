package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type FamilyBackupDAO struct {
	db *gorm.DB
}

func NewFamilyBackupDAO(db *gorm.DB) *FamilyBackupDAO {
	return &FamilyBackupDAO{db: db}
}

func (obj *FamilyBackupDAO) GetTableName() string {
	return TableNameFamilyBackup
}

func (obj *FamilyBackupDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *FamilyBackupDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *FamilyBackupDAO) UpdateByID(ctx context.Context, id int64, record *FamilyBackup) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id=?", id).
		Updates(record).Error
}

func (obj *FamilyBackupDAO) DeleteByCreatedBy(ctx context.Context, createdBy int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Delete(&FamilyBackup{}, `created_by=?`, createdBy).Error
}

func (obj *FamilyBackupDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Delete(&FamilyBackup{ID: id}).Error
}

func (obj *FamilyBackupDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Delete(&FamilyBackup{}, `created_by=? OR member_id=?`, userID, userID).Error
}

func (obj *FamilyBackupDAO) GetByCreatedByAndMemberIDWithFamilyID(ctx context.Context, familyID string, createdBy, memberID int64) (*FamilyBackup, error) {
	var result FamilyBackup
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id=?", familyID).
		Where("created_by=?", createdBy).
		Where("member_id=?", memberID).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *FamilyBackupDAO) GetByUserID(ctx context.Context, userID int64) ([]*FamilyBackup, error) {
	var result []*FamilyBackup
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("created_by=? OR member_id=?", userID, userID).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return result, nil
	}
}

func (obj *FamilyBackupDAO) GetByCreator(ctx context.Context, createdBy int64) ([]*FamilyBackup, error) {
	var result []*FamilyBackup
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("created_by=?", createdBy).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return result, nil
	}
}
