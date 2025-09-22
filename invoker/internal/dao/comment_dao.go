package dao

import (
	"context"

	. "github.com/nextsurfer/invoker/internal/model"
	"gorm.io/gorm"
)

type CommentDAO struct {
	db *gorm.DB
}

func NewCommentDAO(db *gorm.DB) *CommentDAO {
	return &CommentDAO{db: db}
}

func (obj *CommentDAO) GetTableName() string {
	return TableNameComment
}

func (obj *CommentDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *CommentDAO) GetByID(ctx context.Context, id int64) (*Comment, error) {
	var result Comment
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

func (obj *CommentDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *CommentDAO) UpdateByID(ctx context.Context, id int64, one *Comment) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id = ?", id).
		Updates(one).Error
}

func (obj *CommentDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).Delete(&Comment{ID: id}).Error
}

func (obj *CommentDAO) DeleteBySiteID(ctx context.Context, siteID int64) error {
	return obj.db.WithContext(ctx).Delete(&Comment{}, `site_id=?`, siteID).Error
}

type CountIndex struct {
	ID        int64 `gorm:"column:id;"`
	RowNumber int64 `gorm:"column:rowNumber;"`
}

func (obj *CommentDAO) GetFirstLevelCountAndIndex(ctx context.Context, postID, commentID int64) (*CountIndex, error) {
	var result CountIndex
	subquery := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Order(`updated_at ASC`).
		Select(`id, ROW_NUMBER() OVER (ORDER BY updated_at ASC) AS rowNumber`).
		Where(`post_id=? AND root_comment_id=0`, postID)
	conn := obj.db.WithContext(ctx).
		Select(`id, rowNumber`).
		Table(`(?) AS sub_table`, subquery).
		Where(`id=?`, commentID).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *CommentDAO) GetSecondLevelCountAndIndex(ctx context.Context, rootCommentID, commentID int64) (*CountIndex, error) {
	var result CountIndex
	subquery := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Order(`updated_at ASC`).
		Select(`id, ROW_NUMBER() OVER (ORDER BY updated_at ASC) AS rowNumber`).
		Where(`root_comment_id=?`, rootCommentID)
	conn := obj.db.WithContext(ctx).
		Select(`id, rowNumber`).
		Table(`(?) AS sub_table`, subquery).
		Where(`id=?`, commentID).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}
