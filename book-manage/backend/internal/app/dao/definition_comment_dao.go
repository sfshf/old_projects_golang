package dao

import (
	"context"
	"fmt"
	"time"

	. "github.com/nextsurfer/book-manage-api/internal/app/model"
	"gorm.io/gorm"
)

type DefinitionCommentDAO struct {
	db *gorm.DB
}

func NewDefinitionCommentDAO(db *gorm.DB) *DefinitionCommentDAO {
	return &DefinitionCommentDAO{db: db}
}

func (obj *DefinitionCommentDAO) GetTableName() string {
	return "definition_comment"
}

func (obj *DefinitionCommentDAO) Create(ctx context.Context, d *DefinitionComment) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(d).Error
	return err
}

func (obj *DefinitionCommentDAO) Update(ctx context.Context, d *DefinitionComment, selects ...interface{}) error {
	conn := obj.db.WithContext(ctx).Model(d)
	if len(selects) > 0 {
		conn = conn.Select(selects[0], selects[1:]...)
	}
	conn.Where("deleted_at = 0").Updates(d)
	if conn.RowsAffected != 1 {
		return fmt.Errorf("update item with id %d failed in table %s ", d.ID, obj.GetTableName())
	}
	return conn.Error
}

func (obj *DefinitionCommentDAO) GetFromID(ctx context.Context, id int64) (*DefinitionComment, error) {
	var result DefinitionComment
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	if result.ID == 0 {
		return nil, fmt.Errorf("can't find item with id %d in table %s", id, obj.GetTableName())
	}
	return &result, nil
}

func (obj *DefinitionCommentDAO) GetFromDefinitionID(ctx context.Context, id int64) (*DefinitionComment, error) {
	var result DefinitionComment
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("definition_id = ?", id).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteByID simple delete function
func (obj *DefinitionCommentDAO) DeleteByID(ctx context.Context, id int64) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli())
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete item with id %d failed in table %s ", id, obj.GetTableName())
	}
	return conn.Error
}

func (obj *DefinitionCommentDAO) DeleteByDefinitionID(ctx context.Context, definitionID int64) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("definition_id = ?", definitionID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
}

func (obj *DefinitionCommentDAO) DeleteFieldsByID(ctx context.Context, id int64, fields ...string) error {
	updates := make(map[string]interface{}, len(fields))
	for _, field := range fields {
		updates[field] = gorm.Expr("NULL")
	}
	conn := obj.db.WithContext(ctx).Model(&DefinitionComment{ID: id})
	conn.Where("deleted_at = 0").Updates(updates)
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete field %v with id %d failed in table %s ", fields, id, obj.GetTableName())
	}
	return conn.Error
}
