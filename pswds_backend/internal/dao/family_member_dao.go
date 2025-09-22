package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type FamilyMemberDAO struct {
	db *gorm.DB
}

func NewFamilyMember(db *gorm.DB) *FamilyMemberDAO {
	return &FamilyMemberDAO{db: db}
}

func (obj *FamilyMemberDAO) GetTableName() string {
	return TableNameFamilyMember
}

func (obj *FamilyMemberDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

const (
	FamilyMemberIsNotAdmin = 1
	FamilyMemberIsAdmin    = 2
)

func (obj *FamilyMemberDAO) GetByUserID(ctx context.Context, userID int64) (*FamilyMember, error) {
	var result FamilyMember
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *FamilyMemberDAO) CountByFamilyID(ctx context.Context, familyID string) (int64, error) {
	var result int64
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Count(&result).Error; err != nil {
		return 0, err
	}
	return result, nil
}

func (obj *FamilyMemberDAO) GetByFamilyID(ctx context.Context, familyID string) ([]*FamilyMember, error) {
	var result []*FamilyMember
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

type FamilyMemberInfo struct {
	FamilyMember
	UserPublicKey string `gorm:"column:user_public_key;not null;comment:user public key" json:"user_public_key"` // user public key
}

func (obj *FamilyMemberDAO) GetWithUserPublicKeyByFamilyID(ctx context.Context, familyID string) ([]*FamilyMemberInfo, error) {
	var result []*FamilyMemberInfo
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id=?", familyID).
		Joins("LEFT JOIN backup ON backup.user_id=family_member.user_id").
		Select(
			"family_member.id",
			"family_member.created_at",
			"family_member.user_id",
			"family_member.family_id",
			"family_member.is_admin",
			"backup.user_public_key",
		).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return result, nil
	}
}

func (obj *FamilyMemberDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *FamilyMemberDAO) DeleteByFamilyIDAndUserID(ctx context.Context, familyID string, userID int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Where("user_id = ?", userID).
		Delete(&FamilyMember{}).Error
}

func (obj *FamilyMemberDAO) DeleteByFamilyID(ctx context.Context, familyID string) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Delete(&FamilyMember{}).Error
}

func (obj *FamilyMemberDAO) UpdateAdminByUserID(ctx context.Context, userID int64, state int) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Update("is_admin", state).Error
}
