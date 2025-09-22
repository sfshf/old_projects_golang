package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type FamilyMessageDAO struct {
	db *gorm.DB
}

func NewFamilyMessageDAO(db *gorm.DB) *FamilyMessageDAO {
	return &FamilyMessageDAO{db: db}
}

func (obj *FamilyMessageDAO) GetTableName() string {
	return TableNameFamilyMessage
}

func (obj *FamilyMessageDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

const (
	FamilyMessageOperationInviteMember = iota + 1
	FamilyMessageOperationAcceptInvitation
	FamilyMessageOperationRejectInvitation
	FamilyMessageOperationAuthorizeAdmin
	FamilyMessageOperationUnauthorizeAdmin
	FamilyMessageOperationRemoveMember
	FamilyMessageOperationLeaveFamily
	FamilyMessageOperationCancelInvitation
)

func (obj *FamilyMessageDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *FamilyMessageDAO) GetByFamilyID(ctx context.Context, familyID string) ([]*FamilyMessage, error) {
	var result []*FamilyMessage
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Order("created_at DESC").
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *FamilyMessageDAO) DeleteExpiredByFamilyID(ctx context.Context, familyID string) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Where("created_at < ?", time.Now().Add(-time.Hour*24*30).Unix()).
		Delete(&FamilyMessage{}).Error
}

func (obj *FamilyMessageDAO) DeleteByFamilyID(ctx context.Context, familyID string) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Delete(&FamilyMessage{}).Error
}
