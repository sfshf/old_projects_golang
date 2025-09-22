package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/oracle/internal/model"
	"gorm.io/gorm"
)

type ApplicationDAO struct {
	db *gorm.DB
}

func NewApplicationDAO(db *gorm.DB) *ApplicationDAO {
	return &ApplicationDAO{db: db}
}

func (obj *ApplicationDAO) GetTableName() string {
	return TableNameApplication
}

func (obj *ApplicationDAO) GetAll(ctx context.Context) ([]*Application, error) {
	var list []*Application
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *ApplicationDAO) GetByName(ctx context.Context, name string) (*Application, error) {
	var result Application
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("name = ?", name).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *ApplicationDAO) GetByID(ctx context.Context, id int64) (*Application, error) {
	var result Application
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

func (obj *ApplicationDAO) Create(ctx context.Context, one *Application) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *ApplicationDAO) DeleteByName(ctx context.Context, name string) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("name = ?", name).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
