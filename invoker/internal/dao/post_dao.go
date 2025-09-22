package dao

import (
	"context"

	. "github.com/nextsurfer/invoker/internal/model"
	"gorm.io/gorm"
)

type PostDAO struct {
	db *gorm.DB
}

func NewPostDAO(db *gorm.DB) *PostDAO {
	return &PostDAO{db: db}
}

func (obj *PostDAO) GetTableName() string {
	return TableNamePost
}

func (obj *PostDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

const (
	PostState_Posted  = 1
	PostState_Deleted = 2
)

func (obj *PostDAO) CountByCategoryID(ctx context.Context, categoryID int64) (int64, error) {
	var result int64
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("category_id=? AND state=?", categoryID, PostState_Posted).
		Count(&result)
	return result, conn.Error
}

func (obj *PostDAO) Create(ctx context.Context, one *Post) error {
	return obj.db.WithContext(ctx).Create(one).Error
}

func (obj *PostDAO) GetByID(ctx context.Context, id int64) (*Post, error) {
	var result Post
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("state=?", PostState_Posted).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *PostDAO) UpdateByID(ctx context.Context, id int64, one *Post) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("state = ?", PostState_Posted).
		Updates(one).Error
}

func (obj *PostDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where(`id=?`, id).
		Where(`state=?`, PostState_Posted).
		Update("state", PostState_Deleted).Error
}

func (obj *PostDAO) DeleteBySiteID(ctx context.Context, siteID int64) error {
	return obj.db.WithContext(ctx).Delete(&Post{}, `site_id=?`, siteID).Error
}
