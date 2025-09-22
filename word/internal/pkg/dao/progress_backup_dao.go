package dao

import (
	"context"

	. "github.com/nextsurfer/word/internal/pkg/model"
	"gorm.io/gorm"
)

type ProgressBackupDAO struct {
	db *gorm.DB
}

func NewProgressBackupDAO(db *gorm.DB) *ProgressBackupDAO {
	return &ProgressBackupDAO{db: db}
}

func (obj *ProgressBackupDAO) GetTableName() string {
	return TableNameProgressBackup
}

func (obj *ProgressBackupDAO) GetByID(ctx context.Context, id int64) (*ProgressBackup, error) {
	var result ProgressBackup
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *ProgressBackupDAO) GetLatestByUserIDAndVersion(ctx context.Context, userID int64, version int32) (*ProgressBackup, error) {
	var result ProgressBackup
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("version = ?", version).
		Where("deleted_at = 0").
		Order(`timestamp DESC`).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *ProgressBackupDAO) Create(ctx context.Context, one *ProgressBackup) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}
