package dao

import (
	"context"
	"fmt"
	"time"

	. "github.com/nextsurfer/book-manage-api/internal/app/model"
	"gorm.io/gorm"
)

// ExampleDAO is a dao service
type ExampleDAO struct {
	db *gorm.DB
}

// NewExampleDAO to create a dao service
func NewExampleDAO(db *gorm.DB) *ExampleDAO {
	return &ExampleDAO{db: db}
}

// GetTableName get sql table name.
func (obj *ExampleDAO) GetTableName() string {
	return "example"
}

func (obj *ExampleDAO) GetLinkTableName() string {
	return "example_link_book"
}

// GetFromID
func (obj *ExampleDAO) GetFromID(ctx context.Context, id int64) (*Example, error) {
	var result Example
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

func (obj *ExampleDAO) GetNullableItemByID(ctx context.Context, id int64) (*Example, error) {
	var result Example
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

// GetFromIDs
func (obj *ExampleDAO) GetFromDefinitionID(ctx context.Context, id int64) ([]Example, error) {
	result := make([]Example, 0)
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("definition_id = ?", id).
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Create to insert
func (obj *ExampleDAO) Create(ctx context.Context, d *Example) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(d).Error
	return err
}

// CreateWithID to create with a given id
func (obj *ExampleDAO) CreateWithID(ctx context.Context, d *Example) error {
	// delete existing item
	if err := obj.db.Unscoped().WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", d.ID).
		Delete(&Example{}).Error; err != nil {
		return err
	}

	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(d).Error
	return err
}

// DeleteByID simple delete function
func (obj *ExampleDAO) DeleteByID(ctx context.Context, id int64) error {
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
func (obj *ExampleDAO) Update(ctx context.Context, d *Example, selects ...interface{}) error {
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

func (obj *ExampleDAO) DeleteByDefinitionID(ctx context.Context, definitionID int64) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("definition_id = ?", definitionID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
}
