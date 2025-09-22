package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nextsurfer/book-manage-api/internal/app/dao"
	. "github.com/nextsurfer/book-manage-api/internal/app/dto"

	// . "github.com/nextsurfer/book-manage-api/internal/app/model"

	"github.com/wumansgy/goEncrypt"
	// "flag"
)

const publicKey = "ef1cb6e72d149b184cc241037203f60b"
const publicIv = "97893f46e7e13f23"

type BookData struct {
	Book         BookDTO          `json:"book"`
	Definitions  []DefinitionDTO  `json:"definitions"`
	Examples     []ExampleDTO     `json:"examples"`
	RelatedForms []RelatedFormDTO `json:"relatedForms"`
}

func ExportEncryptedBook(bookID int64, outPath string, dbManager *dao.Manager) error {
	err, data := getBookDataFromDB(bookID, dbManager)
	if err != nil {
		fmt.Println("Encrypt Book failed, book ID: ", bookID)
		fmt.Println(err)
		return err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("could not marshal json: %s\n", err)
		fmt.Println(err)
		return err
	}
	fmt.Printf("Get Book (%s), description: (%s) \n", data.Book.Name, data.Book.Description)
	// fmt.Printf(" JSON: \n %s \n", jsonData)

	// 传入明文和自己定义的密钥，密钥为16字节 可以自己传入初始化向量,如果不传就使用默认的初始化向量,16字节
	cryptText, err := goEncrypt.AesCbcEncrypt(jsonData, []byte(publicKey), []byte(publicIv)...)

	if err != nil {
		fmt.Println(err)
		return err
	}
	var file, err2 = os.Create(outPath)
	if err2 != nil {
		fmt.Println("create file error")
		fmt.Println(err2)
		return err2
	}
	defer file.Close()
	l, werr := file.Write(cryptText)
	if werr != nil {
		fmt.Println("write file error")
		fmt.Println(werr)
		return werr
	}
	if l != len(cryptText) {
		fmt.Println("something error happended , writed content != input content")
	}
	fmt.Println("encrypt dictionary file success !")
	return nil
}

func getBookDataFromDB(bookID int64, dbManager *dao.Manager) (error, BookData) {
	// dbPath := "root:waf12KFkwo2@tcp(127.0.0.1:3306)/word?charset=utf8&interpolateParams=True&parseTime=true"

	data := BookData{}

	// db, err := gorm.Open(mysql.Open(dbPath), &gorm.Config{
	// 	PrepareStmt: false,
	// 	NamingStrategy: schema.NamingStrategy{
	// 		SingularTable: true,
	// 	},
	// })
	// if err != nil {
	// 	fmt.Println("connect to db error")
	// 	fmt.Println(err)
	// 	return err, data
	// }

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	fmt.Println("db.DB() error")
	// 	fmt.Println(err)
	// 	return err, data
	// }
	// sqlDB.SetMaxIdleConns(10)
	// sqlDB.SetMaxOpenConns(100)
	// sqlDB.SetConnMaxLifetime(time.Hour)

	// fmt.Println("connected to db")
	manager := dbManager
	// get book
	ctx := context.TODO()
	// get book
	book, bookError := manager.BookDAO.GetFromID(ctx, bookID)
	if bookError != nil {
		fmt.Println("BookDAO.GetFromID : ", bookID)
		fmt.Println(bookError)
		return bookError, data
	}

	data.Book = BookDTO{
		ID:          book.ID,
		UpdatedAt:   book.UpdatedAt.UnixMilli(),
		Name:        book.Name,
		Description: book.Description,
		DownloadURL: book.DownloadURL,
	}

	// get all definitions.

	definitionLinks, relatedError := manager.RelatedDAO.GetDefinitionsByBookID(ctx, bookID)
	if relatedError != nil {
		fmt.Println("RelatedDAO.GetDefinitionsByBookID Error by ID : ", bookID)
		fmt.Println(relatedError)
		return relatedError, data
	}
	definitionDTOs := []DefinitionDTO{}
	formDTOs := []RelatedFormDTO{}
	for _, link := range definitionLinks {
		definitionDO, definitionError := manager.DefinitionDAO.GetFromID(ctx, link.ItemID)
		if definitionError != nil {
			fmt.Println("DefinitionDAO.GetFromID .Error by ID : ", link.ItemID)
			fmt.Println(definitionError)
			return definitionError, data
		}
		definitionDTO := DefinitionDTO{
			ID:        definitionDO.ID,
			UpdatedAt: definitionDO.UpdatedAt.UnixMilli(),
			StringID:  definitionDO.StringID,
			BookID:    link.BookID,

			PartOfSpeech:          definitionDO.PartOfSpeech,
			SpecificType:          definitionDO.SpecificType,
			PronunciationIPA:      definitionDO.PronunciationIpa,
			WeakPronunciationIPA:  definitionDO.PronunciationIpaWeak,
			OtherPronunciationIPA: definitionDO.PronunciationIpaOther,
			PronunciationText:     definitionDO.PronunciationText,
			SortValue:             link.SortValue,
			Level:                 definitionDO.CefrLevel,
			Definition:            definitionDO.Definition,
		}

		// find string
		str, strError := manager.StringDAO.GetFromID(ctx, definitionDO.StringID)
		if strError != nil {
			fmt.Println("StringDAO.GetFromID .Error by ID : ", definitionDO.StringID)
			fmt.Println(strError)
			return strError, data
		}
		definitionDTO.String = str.String

		definitionDTOs = append(definitionDTOs, definitionDTO)

		// relatedForms
		relatedForms, formError := manager.RelatedDAO.GetRelatedFormsByDefinitionID(ctx, definitionDO.ID)
		if formError != nil {
			fmt.Println("RelatedDAO.GetRelatedFormsByDefinitionID .Error by ID : ", definitionDO.ID)
			fmt.Println(formError)
			return formError, data
		}
		for _, form := range relatedForms {
			formDTO := RelatedFormDTO{
				FormStringID:     form.StringID,
				FormDefinitionID: form.DefinitionID,
				WordStringID:     definitionDO.StringID,
				WordDefinitionID: definitionDO.ID,
				Form:             form.Form,
				String:           form.String,
				PronunciationIPA: form.Pronunciation,
				Definition:       form.Definition,
				BookID:           bookID,
			}
			formDTOs = append(formDTOs, formDTO)
		}
	}

	data.Definitions = definitionDTOs
	data.RelatedForms = formDTOs
	// all examples

	// var examples []Example
	exampleDTOs := []ExampleDTO{}
	exampleLinks, relatedError2 := manager.RelatedDAO.GetExamplesByBookID(ctx, bookID)
	if relatedError2 != nil {
		fmt.Println(".RelatedDAO.GetExamplesByBookID .Error by ID : ", bookID)
		fmt.Println(relatedError2)
		return relatedError2, data
	}
	for _, link := range exampleLinks {
		exampleDO, exampleError := manager.ExampleDAO.GetFromID(ctx, link.ItemID)
		if exampleError != nil {
			fmt.Println("ExampleDAO.GetFromID .Error by ID : ", link.ItemID)
			fmt.Println(exampleError)
			return exampleError, data
		}

		exampleDTO := ExampleDTO{
			ID:           exampleDO.ID,
			UpdatedAt:    exampleDO.UpdatedAt.UnixMilli(),
			StringID:     exampleDO.StringID,
			BookID:       link.BookID,
			DefinitionID: exampleDO.DefinitionID,

			Content:       exampleDO.Content,
			WordPositions: exampleDO.WordPositions,
			SortValue:     link.SortValue,
		}
		exampleDTOs = append(exampleDTOs, exampleDTO)
	}
	data.Examples = exampleDTOs

	// data check. do not export book with error.
	// check examples:
	for _, definition := range data.Definitions {
		var hasExample = false
		for _, example := range data.Examples {
			if example.DefinitionID == definition.ID {
				hasExample = true
				break
			}
		}
		if !hasExample {
			fmt.Println("Definition without example: ", definition.ID)
			return fmt.Errorf("definition without example: %d", definition.ID), data
		}
	}

	return nil, data
}
