package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type ReferralCodeDAO struct {
	db *gorm.DB
}

func NewReferralCodeDAO(db *gorm.DB) *ReferralCodeDAO {
	return &ReferralCodeDAO{db: db}
}

func (obj *ReferralCodeDAO) GetTableName() string {
	return TableNameReferralCode
}

func (obj *ReferralCodeDAO) Create(ctx context.Context, one *ReferralCode) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *ReferralCodeDAO) GetByCode(ctx context.Context, code string) (*ReferralCode, error) {
	var result ReferralCode
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
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

func (obj *ReferralCodeDAO) GetByUserIDAndApp(ctx context.Context, userID int64, app string) (*ReferralCode, error) {
	var result ReferralCode
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ? AND app = ?", userID, app).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *ReferralCodeDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
