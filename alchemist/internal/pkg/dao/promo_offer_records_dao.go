package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"gorm.io/gorm"
)

type PromoOfferRecordsDAO struct {
	db *gorm.DB
}

func NewPromoOfferRecordsDAO(db *gorm.DB) *PromoOfferRecordsDAO {
	return &PromoOfferRecordsDAO{db: db}
}

func (obj *PromoOfferRecordsDAO) GetTableName() string {
	return TableNamePromoOfferRecord
}

func (obj *PromoOfferRecordsDAO) Create(ctx context.Context, one *PromoOfferRecord) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *PromoOfferRecordsDAO) UpdateByID(ctx context.Context, id int64, one *PromoOfferRecord) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Updates(one).Error
	return err
}

func (obj *PromoOfferRecordsDAO) GetByUserIDAndOfferIDAppEnv(ctx context.Context, userID int64, offerID, app string, env int32) (*PromoOfferRecord, error) {
	var result PromoOfferRecord
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("offer_id = ?", offerID).
		Where("app = ?", app).
		Where("environment = ?", env).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *PromoOfferRecordsDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}
