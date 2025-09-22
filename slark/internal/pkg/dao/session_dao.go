package dao

import (
	"context"
	"fmt"
	"time"

	. "github.com/nextsurfer/slark/internal/pkg/model"
	"gorm.io/gorm"
)

// SessionDAO is a dao service
type SessionDAO struct {
	db *gorm.DB
}

// NewSessionDAO to create a dao service
func NewSessionDAO(db *gorm.DB) *SessionDAO {
	return &SessionDAO{db: db}
}

// GetTableName get sql table name.获取数据库名字
func (obj *SessionDAO) GetTableName() string {
	return TableNameSlkSession
}

// GetFromID 通过id获取内容 Primary key
func (obj *SessionDAO) GetFromID(ctx context.Context, id int64) (*SlkSession, error) {
	var result SlkSession
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		First(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (obj *SessionDAO) GetFromUserID(ctx context.Context, userID int64) ([]SlkSession, error) {
	var result []SlkSession
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *SessionDAO) GetFromSessionID(ctx context.Context, sessionID string) (*SlkSession, error) {
	var result SlkSession
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("session_id = ?", sessionID).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

// Create to insert
func (obj *SessionDAO) Create(ctx context.Context, item *SlkSession) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(item).Error
	return err
}

// DeleteByID simple delete function
func (obj *SessionDAO) DeleteByID(ctx context.Context, id int64) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().Unix())
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete item with id %d failed in table %s ", id, obj.GetTableName())
	}
	return conn.Error
}

// DeleteBySessionID delete session
func (obj *SessionDAO) DeleteBySessionID(ctx context.Context, sessionID string) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("session_id = ?", sessionID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().Unix())
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete item with session_id %s failed in table %s ", sessionID, obj.GetTableName())
	}
	return conn.Error
}

// DeleteSession there is only on session for one user in one application on one device.
// So we should clean other session when login
func (obj *SessionDAO) DeleteSession(ctx context.Context, userID int64, deviceID string, application string) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("device_id = ?", deviceID).
		Where("application = ?", application).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().Unix())
	return conn.Error
}

// UpdateLoginIPInSession update login ip
func (obj *SessionDAO) UpdateLoginIPInSession(ctx context.Context, sessionID string, loginIP string) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("session_id = ?", sessionID).
		Where("deleted_at = 0").
		Update("login_ip", loginIP)
	return conn.Error
}
