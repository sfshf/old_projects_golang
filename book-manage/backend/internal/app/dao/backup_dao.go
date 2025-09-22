package dao

import (
	"context"
	"fmt"

	. "github.com/nextsurfer/book-manage-api/internal/app/model"
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

func (obj *BackupDAO) GetAll(ctx context.Context) ([]Backup, error) {
	result := make([]Backup, 0)
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *BackupDAO) GetFromID(ctx context.Context, id int64) (*Backup, error) {
	var result Backup
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	if result.ID == 0 {
		return nil, fmt.Errorf("can't find item with id %d in table %s", id, obj.GetTableName())
	}
	return &result, nil
}

func (obj *BackupDAO) GetByBookID(ctx context.Context, bookID int64) ([]Backup, error) {
	result := make([]Backup, 0)
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("book_id = ?", bookID).
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *BackupDAO) Create(ctx context.Context, backup *Backup) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(backup).Error
	return err
}

func (obj *BackupDAO) Update(ctx context.Context, backup *Backup) error {
	conn := obj.db.WithContext(ctx).Model(backup).
		Where("deleted_at = 0").
		Updates(backup)
	if conn.RowsAffected != 1 {
		return fmt.Errorf("update item with id %d failed in table %s ", backup.ID, obj.GetTableName())
	}
	return conn.Error
}

func (obj *BackupDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).Delete(&Backup{ID: id}, "deleted_at = 0").Error
}
