package dao

import (
	"context"

	. "github.com/nextsurfer/invoker/internal/model"
	"gorm.io/gorm"
)

type ThumbupDAO struct {
	db *gorm.DB
}

func NewThumbupDAO(db *gorm.DB) *ThumbupDAO {
	return &ThumbupDAO{db: db}
}

func (obj *ThumbupDAO) GetTableName() string {
	return TableNameThumbup
}

func (obj *ThumbupDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

const (
	ThumbupType_Post    = 1
	ThumbupType_Comment = 2
)

var (
	ThumbupTypeString = map[int32]string{
		ThumbupType_Post:    "post",
		ThumbupType_Comment: "comment",
	}
)

func (obj *ThumbupDAO) HasThumbupPost(ctx context.Context, userID, postID int64) (bool, error) {
	var result int64
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("post_id=? AND type=?", postID, ThumbupType_Post).
		Where("posted_by=?", userID).
		Count(&result)
	return result > 0, conn.Error
}

func (obj *ThumbupDAO) HasThumbupComment(ctx context.Context, userID, commentID int64) (bool, error) {
	var result int64
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("comment_id=? AND type=?", commentID, ThumbupType_Comment).
		Where("posted_by=?", userID).
		Count(&result)
	return result > 0, conn.Error
}

func (obj *ThumbupDAO) Create(ctx context.Context, one *Thumbup) error {
	return obj.db.WithContext(ctx).Create(one).Error
}

func (obj *ThumbupDAO) DeletePostThumbup(ctx context.Context, userID, postID int64) error {
	return obj.db.WithContext(ctx).Delete(&Thumbup{}, `type=? AND posted_by=? AND post_id=?`, ThumbupType_Post, userID, postID).Error
}

func (obj *ThumbupDAO) DeleteCommentThumbup(ctx context.Context, userID, commentID int64) error {
	return obj.db.WithContext(ctx).Delete(&Thumbup{}, `type=? AND posted_by=? AND comment_id=?`, ThumbupType_Comment, userID, commentID).Error
}

func (obj *ThumbupDAO) DeleteBySiteID(ctx context.Context, siteID int64) error {
	return obj.db.WithContext(ctx).Delete(&Thumbup{}, `site_id=?`, siteID).Error
}
