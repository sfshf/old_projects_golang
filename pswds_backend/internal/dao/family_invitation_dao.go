package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type FamilyInvitationDAO struct {
	db *gorm.DB
}

func NewFamilyInvitationDAO(db *gorm.DB) *FamilyInvitationDAO {
	return &FamilyInvitationDAO{db: db}
}

func (obj *FamilyInvitationDAO) GetTableName() string {
	return TableNameFamilyInvitation
}

func (obj *FamilyInvitationDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *FamilyInvitationDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *FamilyInvitationDAO) GetByID(ctx context.Context, id int64) (*FamilyInvitation, error) {
	var result FamilyInvitation
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id = ?", id).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *FamilyInvitationDAO) GetByEmail(ctx context.Context, email string) (*FamilyInvitation, error) {
	var result FamilyInvitation
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("email = ?", email).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *FamilyInvitationDAO) GetAllByEmail(ctx context.Context, email string) ([]*FamilyInvitation, error) {
	var result []*FamilyInvitation
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("email = ?", email).
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *FamilyInvitationDAO) GetByFamilyIDAndEmail(ctx context.Context, familyID, email string) (*FamilyInvitation, error) {
	var result FamilyInvitation
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Where("email = ?", email).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *FamilyInvitationDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id = ?", id).
		Delete(&FamilyInvitation{}).Error
}

func (obj *FamilyInvitationDAO) DeleteByFamilyID(ctx context.Context, familyID string) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Delete(&FamilyInvitation{}).Error
}

func (obj *FamilyInvitationDAO) DeleteByEmail(ctx context.Context, email string) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("email = ?", email).
		Delete(&FamilyInvitation{}).Error
}
