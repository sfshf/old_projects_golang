package book

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nextsurfer/word/internal/pkg/dao"
	. "github.com/nextsurfer/word/internal/pkg/dto"
	"github.com/wumansgy/goEncrypt/aes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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

func ExportEncryptedBook(bookID int64, dbPath string, outPath string) {
	err, data := getBookDataFromDB(bookID, dbPath)
	if err != nil {
		fmt.Println("Encrypt Book failed, book ID: ", bookID)
		fmt.Println(err)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("could not marshal json: %s\n", err)
		return
	}
	fmt.Printf("Get Book (%s), description: (%s) \n", data.Book.Name, data.Book.Description)

	// 传入明文和自己定义的密钥，密钥为16字节 可以自己传入初始化向量,如果不传就使用默认的初始化向量,16字节
	cryptText, err := aes.AesCbcEncrypt(jsonData, []byte(publicKey), []byte(publicIv))

	if err != nil {
		fmt.Println(err)
	}

	var file, err2 = os.Create(outPath)
	if err2 != nil {
		panic(err2)
	}
	defer file.Close()
	l, werr := file.Write(cryptText)
	if werr != nil {
		panic(werr)
	}
	if l != len(cryptText) {
		log.Fatal("something error happended , writed content != input content")
	}
	fmt.Print("encrypt dictionary file success !")
}

func getBookDataFromDB(bookID int64, dbPath string) (error, BookData) {
	// dbPath := "root:waf12KFkwo2@tcp(127.0.0.1:3306)/word?charset=utf8&interpolateParams=True&parseTime=true"

	data := BookData{}

	db, err := gorm.Open(mysql.Open(dbPath), &gorm.Config{
		PrepareStmt: false,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		fmt.Println("connect to db error")
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("db.DB() error")
		log.Fatal(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("connected to db")
	manager := dao.NewManagerWithDB(db)
	// get book
	ctx := context.TODO()
	// get book
	book, bookError := manager.BookDAO.GetFromID(ctx, bookID)
	if bookError != nil {
		fmt.Println("BookDAO.GetFromID : ", bookID)
		log.Fatal(bookError)
		return bookError, data
	}

	data.Book = BookDTO{
		ID:          book.ID,
		UpdatedAt:   book.UpdatedAt.UnixMicro(),
		Name:        book.Name,
		Description: book.Description,
		DownloadURL: book.DownloadURL,
	}

	// get all definitions.

	definitionLinks, relatedError := manager.RelatedDAO.GetDefinitionsByBookID(ctx, bookID)
	if relatedError != nil {
		fmt.Println("RelatedDAO.GetDefinitionsByBookID Error by ID : ", bookID)
		log.Fatal(relatedError)
		return relatedError, data
	}
	definitionDTOs := []DefinitionDTO{}
	formDTOs := []RelatedFormDTO{}
	for _, link := range definitionLinks {
		definitionDO, definitionError := manager.DefinitionDAO.GetFromID(ctx, link.ItemID)
		if definitionError != nil {
			fmt.Println("DefinitionDAO.GetFromID .Error by ID : ", link.ItemID)
			log.Fatal(definitionError)
			return definitionError, data
		}
		definitionDTO := DefinitionDTO{
			ID:        definitionDO.ID,
			UpdatedAt: definitionDO.UpdatedAt.UnixMicro(),
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
			log.Fatal(strError)
			return strError, data
		}
		definitionDTO.String = str.String

		definitionDTOs = append(definitionDTOs, definitionDTO)

		// relatedForms
		relatedForms, formError := manager.RelatedDAO.GetRelatedFormsByDefinitionID(ctx, definitionDO.ID)
		if formError != nil {
			fmt.Println("RelatedDAO.GetRelatedFormsByDefinitionID .Error by ID : ", definitionDO.ID)
			log.Fatal(formError)
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
		log.Fatal(relatedError2)
		return relatedError2, data
	}
	for _, link := range exampleLinks {
		exampleDO, exampleError := manager.ExampleDAO.GetFromID(ctx, link.ItemID)
		if exampleError != nil {
			fmt.Println("ExampleDAO.GetFromID .Error by ID : ", link.ItemID)
			log.Fatal(exampleError)
			return exampleError, data
		}

		exampleDTO := ExampleDTO{
			ID:           exampleDO.ID,
			UpdatedAt:    exampleDO.UpdatedAt.UnixMicro(),
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
	return nil, data
}
