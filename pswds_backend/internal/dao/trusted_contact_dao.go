package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type TrustedContactDAO struct {
	db *gorm.DB
}

func NewTrustedContactDAO(db *gorm.DB) *TrustedContactDAO {
	return &TrustedContactDAO{db: db}
}

func (obj *TrustedContactDAO) GetTableName() string {
	return TableNameTrustedContact
}

func (obj *TrustedContactDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *TrustedContactDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *TrustedContactDAO) GetByUserID(ctx context.Context, userID int64) ([]*TrustedContact, error) {
	var result []*TrustedContact
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *TrustedContactDAO) GetByUserIDAndContactEmail(ctx context.Context, userID int64, contactEmail string) (*TrustedContact, error) {
	var result TrustedContact
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("contact_email = ?", contactEmail).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *TrustedContactDAO) DeleteByID(ctx context.Context, userID int64, id int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("id = ?", id).
		Delete(&TrustedContact{}).Error
}

func (obj *TrustedContactDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Delete(&TrustedContact{}).Error
}

func (obj *TrustedContactDAO) UpdateByID(ctx context.Context, id int64, one *TrustedContact) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id = ?", id).
		Updates(one).Error
}
