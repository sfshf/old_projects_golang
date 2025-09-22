package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/slark/internal/pkg/model"
	"gorm.io/gorm"
)

// ThirdPartyDAO is a dao service
type ThirdPartyDAO struct {
	db *gorm.DB
}

// NewThirdPartyDAO to create a dao service
func NewThirdPartyDAO(db *gorm.DB) *ThirdPartyDAO {
	return &ThirdPartyDAO{db: db}
}

// GetTableName get sql table name
func (obj *ThirdPartyDAO) GetTableName() string {
	return TableNameSlkThirdParty
}

// GetFromOpenID get from open id
func (obj *ThirdPartyDAO) GetFromOpenID(ctx context.Context, application string, openID string, thirdParty string) (*SlkThirdParty, error) {
	var result SlkThirdParty
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("application = ?", application).
		Where("third_party = ?", thirdParty).
		Where("open_id = ?", openID).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

// GetFromUnionID get from union id
func (obj *ThirdPartyDAO) GetFromUnionID(ctx context.Context, application string, unionID string, thirdParty string) (*SlkThirdParty, error) {
	var result SlkThirdParty
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("application = ?", application).
		Where("third_party = ?", thirdParty).
		Where("union_id = ?", unionID).
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
func (obj *ThirdPartyDAO) Create(ctx context.Context, item *SlkThirdParty) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(item).Error
	return err
}

func (obj *ThirdPartyDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("user_id = ?", userID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().Unix()).Error
}
