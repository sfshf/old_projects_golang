package dao

import (
	"context"
	"fmt"
	"net/url"
	"time"

	. "github.com/nextsurfer/book-manage-api/internal/app/model"
	"gorm.io/gorm"
)

// DefinitionDAO is a dao service
type DefinitionDAO struct {
	db *gorm.DB
}

// NewDefinitionDAO to create a dao service
func NewDefinitionDAO(db *gorm.DB) *DefinitionDAO {
	return &DefinitionDAO{db: db}
}

// GetTableName get sql table name.
func (obj *DefinitionDAO) GetTableName() string {
	return "definition"
}

// GetFromID
func (obj *DefinitionDAO) GetFromID(ctx context.Context, id int64) (*Definition, error) {
	var result Definition
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

func (obj *DefinitionDAO) GetNullableItemByID(ctx context.Context, id int64) (*Definition, error) {
	var result Definition
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
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

type DefinitionWithStringParams struct {
	SearchText    string `form:"searchText" json:"searchText" binding:""`
	Type          string `form:"type" json:"type" binding:""`
	PageSize      int    `form:"pageSize" json:"pageSize" binding:""`
	Page          int    `form:"page" json:"page" binding:""`
	BookIDs       []int64
	OrderByString bool `form:"-" json:"-" binding:""`
}

type DefinitionWithString struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	PartOfSpeech string `gorm:"column:part_of_speech;not null" json:"part_of_speech"`
	CefrLevel    string `gorm:"column:cefr_level" json:"cefr_level"`
	Definition   string `gorm:"column:definition;not null" json:"definition"`
	Type         string `gorm:"column:type" json:"type"`
	StringID     int64  `gorm:"column:string_id" json:"string_id"`
	String       string `gorm:"column:string" json:"string"`
	BookID       int64  `gorm:"column:book_id" json:"book_id"`
	Idx          int64  `gorm:"column:idx" json:"idx"`
}

func (obj *DefinitionDAO) SearchDefinitionWithString(ctx context.Context, queries *DefinitionWithStringParams) ([]*DefinitionWithString, int64, error) {
	var res []*DefinitionWithString
	// sub query
	tb := obj.db.WithContext(ctx).Table(TableNameString)
	var err error
	var searchText string
	if queries.SearchText != "" {
		searchText, err = url.QueryUnescape(queries.SearchText)
		if err != nil {
			return nil, 0, err
		}
		tb = tb.Where(`string LIKE ?`, "%"+searchText+"%")
	}
	if queries.Type != "" {
		tb = tb.Where(`type = ?`, queries.Type)
	} else {
		tb = tb.Where(`type = ? OR type = ?`, "word", "phrase")
	}
	tb = tb.Where("deleted_at = 0")
	// joins
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("string_id IN (?)", tb.Distinct("id")).
		Where("definition.deleted_at = 0").
		Distinct("definition.id AS id", "part_of_speech", "cefr_level", "definition", "type", "string_id", "string", "book_id", "idx").
		Joins(`LEFT JOIN string ON string.id = definition.string_id`).
		Joins(`LEFT JOIN (?) AS tb2 ON tb2.item_id = definition.id`,
			obj.db.WithContext(ctx).Table(TableNameRelatedBook).
				Select("book_id", "item_id", `ROW_NUMBER() OVER (PARTITION BY book_id ORDER BY sort_value ASC) AS idx`).
				Where("book_id IN (?)", queries.BookIDs).
				Where("item_type = 'definition'").
				Where("deleted_at = 0")).
		Where("book_id IN (?)", queries.BookIDs)
	if queries.OrderByString {
		if searchText != "" {
			conn = conn.Order(fmt.Sprintf(`CASE WHEN TRIM(string.string) = '%s' THEN 0 ELSE 1 END, string.string ASC`, searchText))
		} else {
			conn = conn.Order(`string.string ASC`)
		}
	}
	var total int64
	if err := conn.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := conn.Offset(queries.Page * queries.PageSize).
		Limit(queries.PageSize).Find(&res).Error; err != nil {
		return nil, 0, err
	}
	return res, total, nil
}

func (obj *DefinitionDAO) GetFromStringIDExclude(ctx context.Context, string_id, definition_id int64) (*Definition, error) {
	var result Definition
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("string_id = ?", string_id).
		Where("id != ?", definition_id).
		Where("deleted_at = 0").
		Find(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

// Create to insert
func (obj *DefinitionDAO) Create(ctx context.Context, d *Definition) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(d).Error
	return err
}

// CreateWithID to create with a given id
func (obj *DefinitionDAO) CreateWithID(ctx context.Context, d *Definition) error {
	// delete existing item
	if err := obj.db.Unscoped().WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", d.ID).
		Delete(&Definition{}).Error; err != nil {
		return err
	}

	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(d).Error
	return err
}

// DeleteByID simple delete function
func (obj *DefinitionDAO) DeleteByID(ctx context.Context, id int64) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli())
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete item with id %d failed in table %s ", id, obj.GetTableName())
	}
	return conn.Error
}

// Update simple Update function
// book with id. If fields are not nil, they will be updated.
func (obj *DefinitionDAO) Update(ctx context.Context, d *Definition, selects ...interface{}) error {
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

func (obj *DefinitionDAO) DeleteFieldsByID(ctx context.Context, id int64, fields ...string) error {
	updates := make(map[string]interface{}, len(fields))
	for _, field := range fields {
		updates[field] = gorm.Expr("NULL")
	}
	conn := obj.db.WithContext(ctx).Model(&Definition{ID: id})
	conn.Where("deleted_at = 0").Updates(updates)
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete field %v with id %d failed in table %s ", fields, id, obj.GetTableName())
	}
	return conn.Error
}
