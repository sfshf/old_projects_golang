package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type FamilyRecoverDAO struct {
	db *gorm.DB
}

func NewFamilyRecoverDAO(db *gorm.DB) *FamilyRecoverDAO {
	return &FamilyRecoverDAO{db: db}
}

func (obj *FamilyRecoverDAO) GetTableName() string {
	return TableNameFamilyRecover
}

func (obj *FamilyRecoverDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

const (
	FamilyRecoverOperationBySelf = iota + 1
	FamilyRecoverOperationFromFamily
)

func (obj *FamilyRecoverDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *FamilyRecoverDAO) GetByUUID(ctx context.Context, uuid string) (*FamilyRecover, error) {
	var result FamilyRecover
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("uuid = ?", uuid).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *FamilyRecoverDAO) GetFamilyRecoverByCreator(ctx context.Context, createdBy, targetID int64) (*FamilyRecover, error) {
	var result FamilyRecover
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("created_by=? AND operation=? AND target_id=?", createdBy, FamilyRecoverOperationFromFamily, targetID).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

type FamilyCiphertextRecover struct {
	FamilyRecover
	Ciphertext string `gorm:"column:ciphertext;not null;comment:ciphertext" json:"ciphertext"` // ciphertext
}

func (obj *FamilyRecoverDAO) GetWithCiphertextByUserID(ctx context.Context, userID, otherID int64) ([]*FamilyCiphertextRecover, error) {
	var result []*FamilyCiphertextRecover
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("(family_recover.created_by=? AND family_recover.operation=? AND family_recover.target_id=?) OR (family_recover.created_by=? AND family_recover.operation=? AND family_recover.target_id=?)",
			userID, FamilyRecoverOperationFromFamily, otherID, // 我（被动）帮助某人找回
			otherID, FamilyRecoverOperationBySelf, userID, // 某人（主动）找我帮忙
		).
		Joins("LEFT JOIN family_backup ON family_backup.id=family_recover.backup_id").
		Select(
			"family_recover.id",
			"family_recover.created_at",
			"family_recover.checked_at",
			"family_recover.created_by",
			"family_recover.target_id",
			"family_recover.operation",
			"family_backup.ciphertext",
		).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return result, nil
	}
}

func (obj *FamilyRecoverDAO) CheckedByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id=?", id).
		Update("checked_at", time.Now().Unix()).Error
}

func (obj *FamilyRecoverDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("created_by=? OR target_id=?", userID, userID).
		Delete(&FamilyRecover{}).Error
}

func (obj *FamilyRecoverDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Delete(&FamilyRecover{ID: id}).Error
}

func (obj *FamilyRecoverDAO) DeleteExpiredByUserID(ctx context.Context, userID, otherID int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("(created_by=? AND operation=? AND target_id=? AND created_at<UNIX_TIMESTAMP(DATE_ADD(NOW(), INTERVAL -1 DAY))) OR (created_by=? AND operation=? AND target_id=? AND created_at<UNIX_TIMESTAMP(DATE_ADD(NOW(), INTERVAL -7 DAY)))",
			userID, FamilyRecoverOperationFromFamily, otherID,
			otherID, FamilyRecoverOperationBySelf, userID,
		).
		Delete(&FamilyRecover{}).Error
}
