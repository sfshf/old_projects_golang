package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type ReferralPointDAO struct {
	db *gorm.DB
}

func NewReferralPointDAO(db *gorm.DB) *ReferralPointDAO {
	return &ReferralPointDAO{db: db}
}

func (obj *ReferralPointDAO) GetTableName() string {
	return TableNameReferralPoint
}

func (obj *ReferralPointDAO) Create(ctx context.Context, one *ReferralPoint) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *ReferralPointDAO) GetByUserIDAndApp(ctx context.Context, userID int64, app string) (*ReferralPoint, error) {
	var result ReferralPoint
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

func (obj *ReferralPointDAO) UpdatePointByID(ctx context.Context, id int64, points int32) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("points", points).Error
}

func (obj *ReferralPointDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
