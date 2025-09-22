package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type PasswordRecordDAO struct {
	db *gorm.DB
}

func NewPasswordRecordDAO(db *gorm.DB) *PasswordRecordDAO {
	return &PasswordRecordDAO{db: db}
}

func (obj *PasswordRecordDAO) GetTableName() string {
	return TableNamePasswordRecord
}

func (obj *PasswordRecordDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *PasswordRecordDAO) GetByUserIDAndDataID(ctx context.Context, userID int64, dataID string) (*PasswordRecord, error) {
	var result PasswordRecord
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("data_id = ?", dataID).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *PasswordRecordDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *PasswordRecordDAO) UpdateByUserIDAndDataID(ctx context.Context, userID int64, dataID string, one *PasswordRecord) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("data_id = ?", dataID).
		Updates(one).Error
}

func (obj *PasswordRecordDAO) GetByUserID(ctx context.Context, userID int64) ([]*PasswordRecord, error) {
	var result []*PasswordRecord
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *PasswordRecordDAO) DeleteByUserIDAndDataID(ctx context.Context, userID int64, dataID string) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("data_id = ?", dataID).
		Delete(&PasswordRecord{}).Error
}
