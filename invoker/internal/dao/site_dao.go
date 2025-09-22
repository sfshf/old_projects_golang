package dao

import (
	"context"

	. "github.com/nextsurfer/invoker/internal/model"
	"gorm.io/gorm"
)

type SiteDAO struct {
	db *gorm.DB
}

func NewSiteDAO(db *gorm.DB) *SiteDAO {
	return &SiteDAO{db: db}
}

func (obj *SiteDAO) GetTableName() string {
	return TableNameSite
}

func (obj *SiteDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *SiteDAO) GetAll(ctx context.Context) ([]*Site, error) {
	var result []*Site
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return result, nil
	}
}

func (obj *SiteDAO) GetByName(ctx context.Context, name string) (*Site, error) {
	var result Site
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("name = ?", name).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *SiteDAO) GetByID(ctx context.Context, id int64) (*Site, error) {
	var result Site
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id = ?", id).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *SiteDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *SiteDAO) UpdateByID(ctx context.Context, id int64, one *Site) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id = ?", id).
		Updates(one).Error
}

func (obj *SiteDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).Delete(&Site{ID: id}).Error
}
