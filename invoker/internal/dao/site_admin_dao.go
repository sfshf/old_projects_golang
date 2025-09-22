package dao

import (
	"context"

	. "github.com/nextsurfer/invoker/internal/model"
	"gorm.io/gorm"
)

type SiteAdminDAO struct {
	db *gorm.DB
}

func NewSiteAdminDAO(db *gorm.DB) *SiteAdminDAO {
	return &SiteAdminDAO{db: db}
}

func (obj *SiteAdminDAO) GetTableName() string {
	return TableNameSiteAdmin
}

func (obj *SiteAdminDAO) GetAdminsBySiteID(ctx context.Context, siteID int64) ([]int64, error) {
	var result []int64
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Distinct(`user_id`).
		Where(`site_id = ?`, siteID).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return result, nil
	}
}

func (obj *SiteAdminDAO) ValidateSiteAdmin(ctx context.Context, siteID, userID int64) (bool, error) {
	var result int64
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where(`site_id=?`, siteID).
		Where(`user_id=?`, userID).
		Count(&result)
	return result > 0, conn.Error
}

func (obj *SiteAdminDAO) Create(ctx context.Context, one *SiteAdmin) error {
	return obj.db.WithContext(ctx).Create(one).Error
}

func (obj *SiteAdminDAO) DeleteBySiteIDAndUserID(ctx context.Context, siteID, userID int64) error {
	return obj.db.WithContext(ctx).Delete(&SiteAdmin{}, `site_id=? AND user_id=?`, siteID, userID).Error
}

func (obj *SiteAdminDAO) DeleteBySiteID(ctx context.Context, siteID int64) error {
	return obj.db.WithContext(ctx).Delete(&SiteAdmin{}, `site_id=?`, siteID).Error
}
