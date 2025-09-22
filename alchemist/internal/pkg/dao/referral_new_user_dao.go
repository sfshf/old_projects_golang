package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type ReferralNewUserDAO struct {
	db *gorm.DB
}

func NewReferralNewUserDAO(db *gorm.DB) *ReferralNewUserDAO {
	return &ReferralNewUserDAO{db: db}
}

func (obj *ReferralNewUserDAO) GetTableName() string {
	return TableNameReferralNewUser
}

func (obj *ReferralNewUserDAO) Create(ctx context.Context, one *ReferralNewUser) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *ReferralNewUserDAO) GetByUserIDAndAppAndReferralCode(ctx context.Context, userID int64, app, code string) (*ReferralNewUser, error) {
	var result ReferralNewUser
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

func (obj *ReferralNewUserDAO) GetByUserIDAndApp(ctx context.Context, userID int64, app string) (*ReferralNewUser, error) {
	var result ReferralNewUser
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

func (obj *ReferralNewUserDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
