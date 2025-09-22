package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type BackupDAO struct {
	db *gorm.DB
}

func NewBackupDAO(db *gorm.DB) *BackupDAO {
	return &BackupDAO{db: db}
}

func (obj *BackupDAO) GetTableName() string {
	return TableNameBackup
}

func (obj *BackupDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *BackupDAO) GetByUserID(ctx context.Context, userID int64) (*Backup, error) {
	var result Backup
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

func (obj *BackupDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *BackupDAO) UpdateByUserID(ctx context.Context, userID int64, one *Backup) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Updates(one).Error
}

func (obj *BackupDAO) DeleteSecurityQuestions(ctx context.Context, userID int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Select("security_questions", "security_questions_ciphertext").
		Where("user_id = ?", userID).
		Updates(&Backup{
			SecurityQuestions:           "",
			SecurityQuestionsCiphertext: "",
		}).Error
}

func (obj *BackupDAO) UpdateFamilyKeyByUserID(ctx context.Context, userID int64, key string) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Update("encrypted_family_key", key).Error
}
