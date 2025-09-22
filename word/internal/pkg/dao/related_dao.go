package dao

import (
	"context"
	"fmt"
	"time"

	. "github.com/nextsurfer/word/internal/pkg/model"
	"gorm.io/gorm"
)

// RelatedDAO is a dao service
type RelatedDAO struct {
	db *gorm.DB
}

// NewRelatedDAO to create a dao service
func NewRelatedDAO(db *gorm.DB) *RelatedDAO {
	return &RelatedDAO{db: db}
}

// GetTableName get sql table name.
func (obj *RelatedDAO) GetRelatedBookTableName() string {
	return "related_book"
}

func (obj *RelatedDAO) GetRelatedDefinitionTableName() string {
	return "related_definition"
}

func (obj *RelatedDAO) GetRelatedBookFromID(ctx context.Context, id int64) (*RelatedBook, error) {
	var result RelatedBook
	conn := obj.db.WithContext(ctx).Table(obj.GetRelatedBookTableName()).
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

func (obj *RelatedDAO) GetRelatedBookForDefinition(ctx context.Context, definitionID int64, bookID int64) (*RelatedBook, error) {
	var result RelatedBook
	conn := obj.db.WithContext(ctx).Table(obj.GetRelatedBookTableName()).
		Where("book_id = ?", bookID).
		Where("item_id = ?", definitionID).
		Where("item_type = 'definition'").
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

func (obj *RelatedDAO) GetRelatedBookForExample(ctx context.Context, exampleID int64, bookID int64) (*RelatedBook, error) {
	var result RelatedBook
	conn := obj.db.WithContext(ctx).Table(obj.GetRelatedBookTableName()).
		Where("book_id = ?", bookID).
		Where("item_id = ?", exampleID).
		Where("item_type = 'example'").
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

func (obj *RelatedDAO) GetDefinitionsByBookID(ctx context.Context, id int64) ([]RelatedBook, error) {
	var result []RelatedBook
	conn := obj.db.WithContext(ctx).Table(obj.GetRelatedBookTableName()).
		Where("book_id = ?", id).
		Where("item_type = 'definition'").
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (obj *RelatedDAO) GetExamplesByBookID(ctx context.Context, id int64) ([]RelatedBook, error) {
	var result []RelatedBook
	conn := obj.db.WithContext(ctx).Table(obj.GetRelatedBookTableName()).
		Where("book_id = ?", id).
		Where("item_type = 'example'").
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (obj *RelatedDAO) GetBooksByDefinitionID(ctx context.Context, id int64) ([]RelatedBook, error) {
	var result []RelatedBook
	conn := obj.db.WithContext(ctx).Table(obj.GetRelatedBookTableName()).
		Where("item_id = ?", id).
		Where("item_type = 'definition'").
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (obj *RelatedDAO) GetBooksByExampleID(ctx context.Context, id int64) ([]RelatedBook, error) {
	var result []RelatedBook
	conn := obj.db.WithContext(ctx).Table(obj.GetRelatedBookTableName()).
		Where("item_id = ?", id).
		Where("item_type = 'example'").
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Create to insert
func (obj *RelatedDAO) CreateRelationForDefinition(ctx context.Context, definitionID int64, bookID int64, sortValue int) error {
	d := &RelatedBook{
		ItemID:    definitionID,
		ItemType:  "definition",
		BookID:    bookID,
		SortValue: int32(sortValue),
	}
	err := obj.db.WithContext(ctx).Table(obj.GetRelatedBookTableName()).Create(d).Error
	return err
}

func (obj *RelatedDAO) CreateRelatationForExample(ctx context.Context, exampleID int64, bookID int64, sortValue int) error {
	d := &RelatedBook{
		ItemID:    exampleID,
		ItemType:  "example",
		BookID:    bookID,
		SortValue: int32(sortValue),
	}
	err := obj.db.WithContext(ctx).Table(obj.GetRelatedBookTableName()).Create(d).Error
	return err
}

// DeleteByID simple delete function
func (obj *RelatedDAO) DeleteLinkByID(ctx context.Context, id int64) error {
	conn := obj.db.WithContext(ctx).Table(obj.GetRelatedBookTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli())
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete item with id %d failed in table %s ", id, obj.GetRelatedBookTableName())
	}
	return conn.Error
}

// db.Where("email LIKE ?", "%jinzhu%").Delete(&Email{})
func (obj *RelatedDAO) DeleteLinksByBookID(ctx context.Context, bookID int64) error {
	conn := obj.db.WithContext(ctx).
		Where("deleted_at = 0").
		Where("book_id = ? ", bookID).
		Update("deleted_at", time.Now().UnixMilli())
	return conn.Error
}

// Update simple Update function
// book with id. If fields are not nil, they will be updated.
func (obj *RelatedDAO) UpdateBookLink(ctx context.Context, d *RelatedBook) error {
	conn := obj.db.WithContext(ctx).Model(d).
		Where("deleted_at = 0").
		Updates(d)
	if conn.RowsAffected != 1 {
		return fmt.Errorf("update item with id %d failed in table %s ", d.ID, obj.GetRelatedBookTableName())
	}
	return conn.Error
}

// related definition

// GetFromID 通过id获取内容 Primary key
func (obj *RelatedDAO) GetRelatedDefinitionFromID(ctx context.Context, id int64) (*RelatedDefinition, error) {
	var result RelatedDefinition
	conn := obj.db.WithContext(ctx).Table(obj.GetRelatedDefinitionTableName()).
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

type RelatedForm struct {
	String        string
	StringID      int64
	PartOfSpeech  string
	Form          string
	Definition    string
	DefinitionID  int64
	Pronunciation string
	// PronunciationSSML string

	BaseStringID     int64
	BaseDefinitionID int64
}

func (obj *RelatedDAO) GetRelatedFormsByDefinitionID(ctx context.Context, id int64) ([]RelatedForm, error) {
	var result []RelatedForm
	conn := obj.db.Table(obj.GetRelatedDefinitionTableName()).Select("string.string, string.id as string_id, definition.part_of_speech, definition.specific_type as form, definition.definition, definition.id as definition_id, definition.pronunciation_ipa as pronunciation, string.base_string_id, related_definition.definition_id as base_definition_id ").
		Joins("JOIN definition ON definition.id = related_definition.related_definition_id AND definition.deleted_at = 0 AND related_definition.deleted_at = 0 AND related_definition.definition_id =?", id).
		Joins("JOIN string ON string.id = definition.string_id AND string.deleted_at = 0").Find(&result)

	err := conn.Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (obj *RelatedDAO) CreateRelatedForm(ctx context.Context, d *RelatedForm) error {
	// string
	formString := String{
		String:       d.String,
		Type:         "form",
		BaseStringID: d.BaseStringID,
	}
	result := obj.db.WithContext(ctx).Create(&formString)
	if result.Error != nil {
		return result.Error
	}
	// definition
	formDefinition := Definition{
		StringID:     formString.ID,
		PartOfSpeech: d.PartOfSpeech,
		SpecificType: d.Form,
		Definition:   d.Definition,
	}
	if d.Pronunciation != "" {
		formDefinition.PronunciationIpa = d.Pronunciation
		// formDefinition.PronunciationSsml = d.PronunciationSSML
	}
	result = obj.db.WithContext(ctx).Create(&formDefinition)
	if result.Error != nil {
		return result.Error
	}
	// relation
	relation := RelatedDefinition{
		DefinitionID:        d.BaseDefinitionID,
		RelatedDefinitionID: formDefinition.ID,
	}
	result = obj.db.WithContext(ctx).Create(&relation)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (obj *RelatedDAO) DeleteRelatedForm(ctx context.Context, d *RelatedForm) error {
	// string
	conn := obj.db.WithContext(ctx).
		Model(&String{}).Where("id = ?", d.StringID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli())
	if conn.Error != nil {
		return conn.Error
	}
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete item with id %d failed in table String", d.StringID)
	}
	// definition
	conn = obj.db.WithContext(ctx).
		Model(&Definition{}).Where("id = ?", d.DefinitionID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli())
	if conn.Error != nil {
		return conn.Error
	}
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete item with id %d failed in table Definition", d.DefinitionID)
	}

	// relation
	conn = obj.db.WithContext(ctx).
		Model(&RelatedDefinition{}).Where("definition_id = ?", d.BaseDefinitionID).Where("related_definition_id = ?", d.DefinitionID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli())
	if conn.Error != nil {
		return conn.Error
	}
	if conn.RowsAffected != 1 {
		return fmt.Errorf("delete item failed in table RelatedDefinition")
	}

	return nil
}

func (obj *RelatedDAO) UpdateRelatedForm(ctx context.Context, d *RelatedForm) error {
	// string
	str := &String{
		ID:           d.StringID,
		String:       d.String,
		Type:         "form",
		BaseStringID: d.BaseStringID,
	}
	conn := obj.db.WithContext(ctx).
		Model(str).
		Where("deleted_at = 0").
		Updates(str)
	if conn.Error != nil {
		return conn.Error
	}
	if conn.RowsAffected != 1 {
		return fmt.Errorf("update item with id %d failed in table String", d.StringID)
	}
	return nil
}

func (obj *RelatedDAO) UpdateRelatedDefinition(ctx context.Context, d *RelatedForm) error {
	definition := &Definition{
		ID:               d.DefinitionID,
		StringID:         d.StringID,
		PartOfSpeech:     d.PartOfSpeech,
		SpecificType:     d.Form,
		PronunciationIpa: d.Pronunciation,
		// PronunciationSsml: d.PronunciationSSML,
		Definition: d.Definition,
	}
	conn := obj.db.WithContext(ctx).
		Model(definition).
		Where("deleted_at = 0").
		Updates(definition)
	if conn.Error != nil {
		return conn.Error
	}
	if conn.RowsAffected != 1 {
		return fmt.Errorf("update item with id %d failed in table Definition", d.DefinitionID)
	}
	return nil
}

func (obj *RelatedDAO) RemoveAllRelatedBook(ctx context.Context) error {
	return obj.db.WithContext(ctx).Delete(&RelatedBook{}, "1=1").Error
}

func (obj *RelatedDAO) RemoveAllRelatedDefinition(ctx context.Context) error {
	return obj.db.WithContext(ctx).Delete(&RelatedDefinition{}, "1=1").Error
}

func (obj *RelatedDAO) CreateRelatedBook(ctx context.Context, one *RelatedBook) error {
	return obj.db.WithContext(ctx).Create(one).Error
}

func (obj *RelatedDAO) CreateRelatedDefinition(ctx context.Context, one *RelatedDefinition) error {
	return obj.db.WithContext(ctx).Create(one).Error
}
