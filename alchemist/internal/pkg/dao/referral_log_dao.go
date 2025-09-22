package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	. "github.com/nextsurfer/alchemist/pkg/consts"
	"gorm.io/gorm"
)

type ReferralLogDAO struct {
	db *gorm.DB
}

func NewReferralLogDAO(db *gorm.DB) *ReferralLogDAO {
	return &ReferralLogDAO{db: db}
}

func (obj *ReferralLogDAO) GetTableName() string {
	return TableNameReferralLog
}

func (obj *ReferralLogDAO) Create(ctx context.Context, one *ReferralLog) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *ReferralLogDAO) GetListByReferralPointID(ctx context.Context, referralPointID int64) ([]*ReferralLog, error) {
	var list []*ReferralLog
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("referral_point_id = ?", referralPointID).
		Where("deleted_at = 0").
		Order("timestamp DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *ReferralLogDAO) GetListByUserID(ctx context.Context, userID int64) ([]*ReferralLog, error) {
	var list []*ReferralLog
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Order("timestamp DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *ReferralLogDAO) CountFirstTimeReferral(ctx context.Context, userID int64) (int64, error) {
	var count int64
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Where("reason = ?", ReferralLogReasonNewUserFirstTime).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (obj *ReferralLogDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
