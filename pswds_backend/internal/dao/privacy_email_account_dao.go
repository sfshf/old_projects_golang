package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
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

func (obj *PrivacyEmailAccountDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *PrivacyEmailAccountDAO) GetByUserID(ctx context.Context, userID int64) ([]*PrivacyEmailAccount, error) {
	var result []*PrivacyEmailAccount
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id=?", userID).
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *PrivacyEmailAccountDAO) GetByUserIDAndAccount(ctx context.Context, userID int64, emailAccount string) (*PrivacyEmailAccount, error) {
	var result PrivacyEmailAccount
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id=?", userID).
		Where("email_account=?", emailAccount).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *PrivacyEmailAccountDAO) GetByAccount(ctx context.Context, emailAccount string) (*PrivacyEmailAccount, error) {
	var result PrivacyEmailAccount
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("email_account=?", emailAccount).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
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
