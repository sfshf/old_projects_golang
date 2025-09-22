package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type UserRegisteredOnOldDeviceDAO struct {
	db *gorm.DB
}

func NewUserRegisteredOnOldDeviceDAO(db *gorm.DB) *UserRegisteredOnOldDeviceDAO {
	return &UserRegisteredOnOldDeviceDAO{db: db}
}

func (obj *UserRegisteredOnOldDeviceDAO) GetTableName() string {
	return TableNameUserRegisteredOnOldDevice
}

func (obj *UserRegisteredOnOldDeviceDAO) Create(ctx context.Context, one *UserRegisteredOnOldDevice) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *UserRegisteredOnOldDeviceDAO) GetByUserIDWithAppAndCode(ctx context.Context, userID int64, app, code string) (*UserRegisteredOnOldDevice, error) {
	var result UserRegisteredOnOldDevice
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("app = ?", app).
		Where("referral_code = ?", code).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *UserRegisteredOnOldDeviceDAO) GetByUserIDWithApp(ctx context.Context, userID int64, app string) (*UserRegisteredOnOldDevice, error) {
	var result UserRegisteredOnOldDevice
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("app = ?", app).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *UserRegisteredOnOldDeviceDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
