package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type NonPasswordRecordDAO struct {
	db *gorm.DB
}

func NewNonPasswordRecordDAO(db *gorm.DB) *NonPasswordRecordDAO {
	return &NonPasswordRecordDAO{db: db}
}

func (obj *NonPasswordRecordDAO) GetTableName() string {
	return TableNameNonPasswordRecord
}

func (obj *NonPasswordRecordDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *NonPasswordRecordDAO) GetByUserIDAndDataID(ctx context.Context, userID int64, dataID string) (*NonPasswordRecord, error) {
	var result NonPasswordRecord
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

func (obj *NonPasswordRecordDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *NonPasswordRecordDAO) UpdateByUserIDAndDataID(ctx context.Context, userID int64, dataID string, one *NonPasswordRecord) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("data_id = ?", dataID).
		Updates(one).Error
}

func (obj *NonPasswordRecordDAO) GetByUserID(ctx context.Context, userID int64) ([]*NonPasswordRecord, error) {
	var result []*NonPasswordRecord
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *NonPasswordRecordDAO) DeleteByUserIDAndDataID(ctx context.Context, userID int64, dataID string) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("data_id = ?", dataID).
		Delete(&NonPasswordRecord{}).Error
}
