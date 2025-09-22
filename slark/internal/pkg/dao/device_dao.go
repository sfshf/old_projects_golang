package dao

import (
	"context"
	"fmt"
	"time"

	. "github.com/nextsurfer/slark/internal/pkg/model"
	"gorm.io/gorm"
)

// DeviceDAO is a dao service
type DeviceDAO struct {
	db *gorm.DB
}

// NewDeviceDAO to create a dao service
func NewDeviceDAO(db *gorm.DB) *DeviceDAO {
	return &DeviceDAO{db: db}
}

// GetTableName get sql table name
func (obj *DeviceDAO) GetTableName() string {
	return TableNameSlkDevice
}

// GetFromID get from id
func (obj *DeviceDAO) GetFromID(ctx context.Context, id int64) (*SlkDevice, error) {
	var result SlkDevice
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	if result.ID == 0 {
		return nil, nil
	}
	return &result, nil
}

// GetFromDeviceID get from device id
func (obj *DeviceDAO) GetFromDeviceID(ctx context.Context, deviceID string) (*SlkDevice, error) {
	var result SlkDevice
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("device_id = ?", deviceID).
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	if result.ID == 0 {
		return nil, nil
	}
	return &result, nil
}

// Create to insert
func (obj *DeviceDAO) Create(ctx context.Context, item *SlkDevice) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(item).Error
	return err
}

// DeleteByID delete by id
func (obj *DeviceDAO) DeleteByID(ctx context.Context, id int64) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().Unix())
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete item with id %d failed in table %s ", id, obj.GetTableName())
	}
	return conn.Error
}
