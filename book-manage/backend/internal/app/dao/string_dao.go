package dao

import (
	"context"
	"fmt"
	"time"

	. "github.com/nextsurfer/book-manage-api/internal/app/model"
	"gorm.io/gorm"
)

// StringDAO is a dao service
type StringDAO struct {
	db *gorm.DB
}

// NewStringDAO to create a dao service
func NewStringDAO(db *gorm.DB) *StringDAO {
	return &StringDAO{db: db}
}

// GetTableName get sql table name.
func (obj *StringDAO) GetTableName() string {
	return "string"
}

// GetFromID
func (obj *StringDAO) GetFromID(ctx context.Context, id int64) (*String, error) {
	var result String
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

func (obj *StringDAO) GetNullableItemByID(ctx context.Context, id int64) (*String, error) {
	var result String
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

func (obj *StringDAO) GetFromString(ctx context.Context, str string) (*String, error) {
	var result String
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("string = ?", str).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result).Error; err != nil {
		return nil, err
	}
	if result.ID == 0 {
		return nil, nil
	}
	return &result, nil
}

// get base form word ID.
func (obj *StringDAO) GetIDFromWord(ctx context.Context, word string) (int64, error) {
	var result String
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").Where("type = 'word'").
		Where("string = ?", word).
		Find(&result)
	err := conn.Error
	if err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (obj *StringDAO) GetIDByPhrase(ctx context.Context, phrase string) (int64, error) {
	var result String
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").Where("type = 'phrase'").
		Where("string = ?", phrase).
		Find(&result)
	err := conn.Error
	if err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Create to insert
func (obj *StringDAO) Create(ctx context.Context, d *String) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(d).Error
	return err
}

func (obj *StringDAO) CreateWord(ctx context.Context, d string) (int64, error) {
	word := &String{
		String: d,
		Type:   "word",
	}
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(word).Error
	return word.ID, err
}

// typo: word, phrase
func (obj *StringDAO) CreateWithID(ctx context.Context, str string, typo string, id int64) error {
	// delete existing word
	if err := obj.db.Unscoped().WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Delete(&String{}).Error; err != nil {
		return err
	}

	word := &String{
		ID:     id,
		String: str,
		Type:   typo,
	}
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(word).Error
	return err
}

func (obj *StringDAO) CreateForm(ctx context.Context, d string, baseStringID int64) (int64, error) {
	form := &String{
		String:       d,
		Type:         "form",
		BaseStringID: baseStringID,
	}
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(form).Error
	return form.ID, err
}

func (obj *StringDAO) CreatePhrase(ctx context.Context, d string) (int64, error) {
	phrase := &String{
		String: d,
		Type:   "phrase",
		// BaseStringID: baseStringID,
	}
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(phrase).Error
	return phrase.ID, err
}

// DeleteByID simple delete function
func (obj *StringDAO) DeleteByID(ctx context.Context, id int64) error {
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
func (obj *StringDAO) Update(ctx context.Context, d *String) error {
	conn := obj.db.WithContext(ctx).Model(d).
		Where("deleted_at = 0").
		Updates(d)
	if conn.RowsAffected != 1 {
		return fmt.Errorf("update item with id %d failed in table %s ", d.ID, obj.GetTableName())
	}
	return conn.Error
}
