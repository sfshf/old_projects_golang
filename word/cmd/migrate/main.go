package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	gdao "github.com/nextsurfer/ground/pkg/dao"
	"github.com/nextsurfer/word/internal/pkg/dao"
	. "github.com/nextsurfer/word/internal/pkg/model"
	"gorm.io/gorm/clause"
)

type ConfigInfo struct {
	BookmanagerMysqlDNS      string  `json:"bookmanagerMysqlDNS"`
	WordMysqlDNSInTest       string  `json:"wordMysqlDNSInTest"`
	WordMysqlDNSInProduction string  `json:"wordMysqlDNSInProduction"`
	BookIDs                  []int64 `json:"bookIDs"`
}

var cnf ConfigInfo

func init() {
	confPath := os.Getenv("MIGRAGE_CONFIG")
	if confPath == "" {
		confPath = "./cmd/migrage/conf.json"
	}
	r, err := os.Open(confPath)
	if err != nil {
		log.Fatalln(err)
	}
	decoder := json.NewDecoder(r)
	err = decoder.Decode(&cnf)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Load config success: %#v\n", cnf)
}

func main() {

	// bookmanager
	bmOpt := gdao.NewOption(cnf.BookmanagerMysqlDNS, "migrate", 13306, 1)
	bmDaoManager := dao.NewManager(bmOpt)

	// word test
	wtOpt := gdao.NewOption(cnf.WordMysqlDNSInTest, "migrate", 13307, 1)
	wtDaoManager := dao.NewManager(wtOpt)

	// auto update table structures
	if err := wtDaoManager.DB.AutoMigrate(
		&Book{},
		&Definition{},
		&Example{},
		&RelatedBook{},
		&RelatedDefinition{},
		&String{},
		&Translation{},
	); err != nil {
		log.Fatalln(err)
	}

	// word production
	wpOpt := gdao.NewOption(cnf.WordMysqlDNSInProduction, "migrate", 13308, 1)
	wpDaoManager := dao.NewManager(wpOpt)

	if err := wpDaoManager.DB.AutoMigrate(
		&Book{},
		&Definition{},
		&Example{},
		&RelatedBook{},
		&RelatedDefinition{},
		&String{},
		&Translation{},
	); err != nil {
		log.Fatalln(err)
	}

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		if err := Migrate(bmDaoManager, wtDaoManager); err != nil {
			fmt.Printf("failed to migrate to word test: %v\n", err)
			return
		}
		fmt.Println("migrate to word test: success")
	}()

	go func() {
		defer wg.Done()

		if err := Migrate(bmDaoManager, wpDaoManager); err != nil {
			fmt.Printf("failed to migrate to word production: %v\n", err)
			return
		}
		fmt.Println("migrate to word production: success")
	}()

	wg.Wait()
}

func Migrate(from, to *dao.Manager) (err error) {
	ctx := context.Background()

	tx, manager := to.Transaction()
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback().Error; rbErr != nil {
				fmt.Printf("tx.Rollback().Error: %s", rbErr.Error())
			} else {
				fmt.Printf("tx.Rollback() success")
			}
		}
	}()

	// Remove all in `to` database.
	if err = RemoveAll(ctx, to); err != nil {
		return err
	}

	for _, bookID := range cnf.BookIDs {
		// book
		var book *Book
		book, err = from.BookDAO.GetFromID(ctx, bookID)
		if err != nil {
			return err
		}
		if err = manager.BookDAO.Create(ctx, book); err != nil {
			return err
		}

		// related_book
		var rows *sql.Rows
		rows, err = from.DB.Model(&RelatedBook{}).
			Where("book_id = ?", bookID).
			Where("deleted_at = 0").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var relatedBook RelatedBook
			if err = from.DB.ScanRows(rows, &relatedBook); err != nil {
				return err
			}
			if err = manager.DB.Table(TableNameRelatedBook).
				Clauses(clause.OnConflict{DoNothing: true}).
				Create(&relatedBook).Error; err != nil {
				return err
			}
			if relatedBook.ItemType == "definition" {
				// definition
				var definition *Definition
				definition, err = from.DefinitionDAO.GetFromID(ctx, relatedBook.ItemID)
				if err != nil {
					return err
				}
				if err = manager.DB.Table(TableNameDefinition).
					Clauses(clause.OnConflict{DoNothing: true}).
					Create(definition).Error; err != nil {
					return err
				}
				// string
				if err = CopyString(ctx, from, manager, definition.StringID); err != nil {
					return err
				}
				// related_definition
				var rows2 *sql.Rows
				rows2, err = from.DB.Model(&RelatedDefinition{}).
					Where("definition_id = ?", definition.ID).
					Where("deleted_at = 0").Rows()
				if err != nil {
					return err
				}
				defer rows2.Close()
				for rows2.Next() {
					var relatedDefinition RelatedDefinition
					if err = from.DB.ScanRows(rows2, &relatedDefinition); err != nil {
						return err
					}
					if err = manager.DB.Table(TableNameRelatedDefinition).
						Clauses(clause.OnConflict{DoNothing: true}).
						Create(&relatedDefinition).Error; err != nil {
						return err
					}
				}
				if err = rows2.Err(); err != nil {
					return err
				}
				// translation
				var rows3 *sql.Rows
				rows3, err = from.DB.Model(&Translation{}).
					Where("item_id = ?", definition.ID).
					Where("item_type = 'definition'").
					Where("deleted_at = 0").Rows()
				if err != nil {
					return err
				}
				defer rows3.Close()
				for rows3.Next() {
					var translation Translation
					if err = from.DB.ScanRows(rows3, &translation); err != nil {
						return err
					}
					if err = manager.DB.Table(TableNameTranslation).
						Clauses(clause.OnConflict{DoNothing: true}).
						Create(&translation).Error; err != nil {
						return err
					}
				}
				if err = rows3.Err(); err != nil {
					return err
				}
			} else if relatedBook.ItemType == "example" {
				var example *Example
				example, err = from.ExampleDAO.GetFromID(ctx, relatedBook.ItemID)
				if err != nil {
					return err
				}
				if err = manager.DB.Table(TableNameExample).
					Clauses(clause.OnConflict{DoNothing: true}).
					Create(example).Error; err != nil {
					return err
				}
				// translation
				var rows2 *sql.Rows
				rows2, err = from.DB.Model(&Translation{}).
					Where("item_id = ?", example.ID).
					Where("item_type = 'example'").
					Where("deleted_at = 0").Rows()
				if err != nil {
					return err
				}
				defer rows2.Close()
				for rows2.Next() {
					var translation Translation
					if err = from.DB.ScanRows(rows2, &translation); err != nil {
						return err
					}
					if err = manager.DB.Table(TableNameTranslation).
						Clauses(clause.OnConflict{DoNothing: true}).
						Create(&translation).Error; err != nil {
						return err
					}
				}
				if err = rows2.Err(); err != nil {
					return err
				}
			}
		}
		if err = rows.Err(); err != nil {
			return err
		}
	}

	if err = tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func CopyString(ctx context.Context, from, to *dao.Manager, stringID int64) error {
	word, err := from.StringDAO.GetFromID(ctx, stringID)
	if err != nil {
		return err
	}
	conn := to.DB.Table(TableNameString).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(word)
	if err := conn.Error; err != nil {
		return err
	}
	if conn.RowsAffected == 1 {
		// recur on its base string id, if not empty
		if word.BaseStringID > 0 {
			return CopyString(ctx, from, to, word.BaseStringID)
		}
	}
	return nil
}

func RemoveAll(ctx context.Context, to *dao.Manager) error {
	var err error
	// book
	err = to.BookDAO.RemoveAll(ctx)
	if err != nil {
		return err
	}
	// definition
	err = to.DefinitionDAO.RemoveAll(ctx)
	if err != nil {
		return err
	}
	// example
	err = to.ExampleDAO.RemoveAll(ctx)
	if err != nil {
		return err
	}
	// related_book
	err = to.RelatedDAO.RemoveAllRelatedBook(ctx)
	if err != nil {
		return err
	}
	// related_definition
	err = to.RelatedDAO.RemoveAllRelatedDefinition(ctx)
	if err != nil {
		return err
	}
	// string
	err = to.StringDAO.RemoveAll(ctx)
	if err != nil {
		return err
	}
	// translation
	err = to.TranslationDAO.RemoveAll(ctx)
	if err != nil {
		return err
	}
	return nil
}
