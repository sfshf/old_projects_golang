package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type UnlockPasswordRecoverDAO struct {
	db *gorm.DB
}

func NewUnlockPasswordRecoverDAO(db *gorm.DB) *UnlockPasswordRecoverDAO {
	return &UnlockPasswordRecoverDAO{db: db}
}

func (obj *UnlockPasswordRecoverDAO) GetTableName() string {
	return TableNameUnlockPasswordRecover
}

func (obj *UnlockPasswordRecoverDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

const (
	UnlockPasswordRecoverTypeSecurityQuestions = iota + 1
	UnlockPasswordRecoverTypeFamilyRecover
	UnlockPasswordRecoverTypeHelpFamilyRecover
)

func (obj *UnlockPasswordRecoverDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *UnlockPasswordRecoverDAO) GetFamilyRecoverByUserID(ctx context.Context, userID int64) (*UnlockPasswordRecover, error) {
	var result UnlockPasswordRecover
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("created_by = ?", userID).
		Where("type = ?", UnlockPasswordRecoverTypeFamilyRecover).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *UnlockPasswordRecoverDAO) GetHelpFamilyRecoverByUserID(ctx context.Context, userID int64) (*UnlockPasswordRecover, error) {
	var result UnlockPasswordRecover
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("created_by = ?", userID).
		Where("type = ?", UnlockPasswordRecoverTypeHelpFamilyRecover).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *UnlockPasswordRecoverDAO) GetSecurityQuestionsRecoverByUserID(ctx context.Context, userID int64) (*UnlockPasswordRecover, error) {
	var result UnlockPasswordRecover
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("created_by = ?", userID).
		Where("type = ?", UnlockPasswordRecoverTypeSecurityQuestions).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *UnlockPasswordRecoverDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Delete(&UnlockPasswordRecover{ID: id}).Error
}
