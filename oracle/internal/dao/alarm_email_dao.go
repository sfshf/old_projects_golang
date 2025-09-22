package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/oracle/internal/model"
	"gorm.io/gorm"
)

type AlarmEmailDAO struct {
	db *gorm.DB
}

func NewAlarmEmailDAO(db *gorm.DB) *AlarmEmailDAO {
	return &AlarmEmailDAO{db: db}
}

func (obj *AlarmEmailDAO) GetTableName() string {
	return TableNameAlarmEmail
}

func (obj *AlarmEmailDAO) GetAll(ctx context.Context) ([]*AlarmEmail, error) {
	var list []*AlarmEmail
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *AlarmEmailDAO) GetByAddress(ctx context.Context, address string) (*AlarmEmail, error) {
	var result AlarmEmail
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("address = ?", address).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *AlarmEmailDAO) Create(ctx context.Context, one *AlarmEmail) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *AlarmEmailDAO) DeleteByID(ctx context.Context, id int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}

func (obj *AlarmEmailDAO) Update(ctx context.Context, one *AlarmEmail) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", one.ID).
		Where("deleted_at = 0").
		Updates(one).Error
	return err
}
