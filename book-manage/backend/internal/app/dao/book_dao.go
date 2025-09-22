package dao

import (
	"context"
	"fmt"
	"time"

	. "github.com/nextsurfer/book-manage-api/internal/app/model"
	"gorm.io/gorm"
)

// BookDAO is a dao service
type BookDAO struct {
	db *gorm.DB
}

// NewBookDAO to create a dao service
func NewBookDAO(db *gorm.DB) *BookDAO {
	return &BookDAO{db: db}
}

// GetTableName get sql table name.获取数据库名字
func (obj *BookDAO) GetTableName() string {
	return "book"
}

// GetFromID 通过id获取内容 Primary key
func (obj *BookDAO) GetFromID(ctx context.Context, id int64) (*Book, error) {
	var result Book
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

func (obj *BookDAO) GetFromName(ctx context.Context, name string) (*Book, error) {
	var result Book
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("name = ?", name).
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

func (obj *BookDAO) GetAll(ctx context.Context) ([]Book, error) {
	result := make([]Book, 0)
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Find(&result)
	err := conn.Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Create to insert, return id and error
func (obj *BookDAO) Create(ctx context.Context, book *Book) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(book).Error
	return err
}

// DeleteByID simple delete function
func (obj *BookDAO) DeleteByID(ctx context.Context, id int64) error {
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
func (obj *BookDAO) Update(ctx context.Context, book *Book) error {
	conn := obj.db.WithContext(ctx).Model(book).
		Where("deleted_at = 0").
		Updates(book)
	if conn.RowsAffected != 1 {
		return fmt.Errorf("update item with id %d failed in table %s ", book.ID, obj.GetTableName())
	}
	return conn.Error
}

func (obj *BookDAO) UpdateDownloadPath(ctx context.Context, bookID int64) error {
	updatedAt := time.Now()
	// fmt.Println("UpdateDownloadPath bookID: ", bookID)
	conn := obj.db.WithContext(ctx).Model(&Book{}).
		Where("deleted_at = 0").
		Where("id = ?", bookID).
		UpdateColumns(Book{UpdatedAt: updatedAt, DownloadURL: "/download/csv/" + fmt.Sprintf("%d", updatedAt.UnixMilli())})
	if conn.RowsAffected != 1 {
		return fmt.Errorf("UpdateDownloadPathwith id %d failed in table %s ", bookID, obj.GetTableName())
	}
	return conn.Error
}
