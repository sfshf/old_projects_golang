package dao

import (
	"context"
	"fmt"
	"time"

	. "github.com/nextsurfer/slark/internal/pkg/model"
	"gorm.io/gorm"
)

// UserDAO is a dao service
type UserDAO struct {
	db *gorm.DB
}

// NewUserDAO to create a dao service
func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

// GetTableName get sql table name.获取数据库名字
func (obj *UserDAO) GetTableName() string {
	return TableNameSlkUser
}

// GetFromID 通过id获取内容 Primary key
func (obj *UserDAO) GetFromID(ctx context.Context, id int64) (*SlkUser, error) {
	var result SlkUser
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

func (obj *UserDAO) GetFromIDs(ctx context.Context, ids []int64) ([]*SlkUser, error) {
	var result []*SlkUser
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id IN (?)", ids).
		Where("deleted_at = 0").
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return result, nil
	}
}

// GetFromPhone get account info by phone number
func (obj *UserDAO) GetFromPhone(ctx context.Context, phone string) (*SlkUser, error) {
	var result SlkUser
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("phone = ?", phone).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

// GetFromEmail get account info by email
func (obj *UserDAO) GetFromEmail(ctx context.Context, email string) (*SlkUser, error) {
	var result SlkUser
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("email = ?", email).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *UserDAO) GetByNickname(ctx context.Context, nickname string) (*SlkUser, error) {
	var result SlkUser
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("nickname = ?", nickname).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *UserDAO) UpdateSecondaryPassword(ctx context.Context, id int64, passwordHash string) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("secondary_password_hash", passwordHash).Error
}

func (obj *UserDAO) UpdateNickname(ctx context.Context, id int64, nickname string) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("nickname", nickname).Error
}

// Create to insert
func (obj *UserDAO) Create(ctx context.Context, user *SlkUser) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(user).Error
	return err
}

// DeleteByID simple delete function
func (obj *UserDAO) DeleteByID(ctx context.Context, id int64) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().Unix())
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete item with id %d failed in table %s ", id, obj.GetTableName())
	}
	return conn.Error
}
