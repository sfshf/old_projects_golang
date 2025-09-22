package dao

import (
	"context"

	. "github.com/nextsurfer/invoker/internal/model"
	"gorm.io/gorm"
)

type CategoryDAO struct {
	db *gorm.DB
}

func NewCategoryDAO(db *gorm.DB) *CategoryDAO {
	return &CategoryDAO{db: db}
}

func (obj *CategoryDAO) GetTableName() string {
	return TableNameCategory
}

func (obj *CategoryDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *CategoryDAO) GetBySiteID(ctx context.Context, siteID int64) ([]*Category, error) {
	var result []*Category
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("site_id = ?", siteID).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return result, nil
	}
}

func (obj *CategoryDAO) GetBySiteIDAndName(ctx context.Context, siteID int64, name string) (*Category, error) {
	var result Category
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("site_id = ?", siteID).
		Where(`name = ?`, name).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *CategoryDAO) GetByID(ctx context.Context, id int64) (*Category, error) {
	var result Category
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

func (obj *CategoryDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *CategoryDAO) UpdateByID(ctx context.Context, id int64, one *Category) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id = ?", id).
		Updates(one).Error
}

func (obj *CategoryDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).Delete(&Category{ID: id}).Error
}

func (obj *CategoryDAO) DeleteBySiteID(ctx context.Context, siteID int64) error {
	return obj.db.WithContext(ctx).Delete(&Category{}, `site_id=?`, siteID).Error
}
