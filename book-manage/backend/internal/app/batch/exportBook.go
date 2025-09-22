package batch

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"sort"

	"github.com/nextsurfer/book-manage-api/internal/app/dao"
	. "github.com/nextsurfer/book-manage-api/internal/app/model"
)

func int64ToString(i int64) string {
	if i == 0 {
		return ""
	} else {
		return strconv.FormatInt(i, 10)
	}
}

func getBookData(bookID int64, manager *dao.Manager) (error, []BookLine, *Book) {
	lines := []BookLine{}

	// db, err := gorm.Open(mysql.Open(dbPath), &gorm.Config{
	// 	PrepareStmt: false,
	// 	NamingStrategy: schema.NamingStrategy{
	// 		SingularTable: true,
	// 	},
	// })
	// if err != nil {
	// 	fmt.Println("connect to db error")
	// 	log.Println(err)
	// 	return err, nil, nil
	// }

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	fmt.Println("db.DB() error")
	// 	log.Println(err)
	// 	return err, nil, nil
	// }
	// sqlDB.SetMaxIdleConns(10)
	// sqlDB.SetMaxOpenConns(100)
	// sqlDB.SetConnMaxLifetime(time.Hour)

	// fmt.Println("connected to db")
	// manager := dao.NewManagerWithDB(db)
	// get book
	ctx := context.TODO()
	book, bookError := manager.BookDAO.GetFromID(ctx, bookID)
	if bookError != nil {
		fmt.Println("BookDAO.GetFromID : ", bookID)
		log.Println(bookError)
		return bookError, lines, book
	}

	// get all definitions.
	var definitions = []Definition{}
	definitionLinks, relatedError := manager.RelatedDAO.GetDefinitionsByBookID(ctx, bookID)
	if relatedError != nil {
		fmt.Println("RelatedDAO.GetDefinitionsByBookID Error by ID : ", bookID)
		log.Println(relatedError)
		return relatedError, lines, book
	}
	for _, link := range definitionLinks {
		definition, definitionError := manager.DefinitionDAO.GetFromID(ctx, link.ItemID)
		if definitionError != nil {
			fmt.Println("DefinitionDAO.GetFromID .Error by ID : ", link.ItemID)
			log.Println(definitionError)
			return definitionError, lines, book
		}
		definitions = append(definitions, *definition)
	}

	// all examples
	exampleLinks, relatedError2 := manager.RelatedDAO.GetExamplesByBookID(ctx, bookID)
	if relatedError2 != nil {
		fmt.Println(".RelatedDAO.GetExamplesByBookID .Error by ID : ", bookID)
		log.Println(relatedError2)
		return relatedError2, lines, book
	}
	var examples = []Example{}
	for _, link := range exampleLinks {
		example, exampleError := manager.ExampleDAO.GetFromID(ctx, link.ItemID)
		if exampleError != nil {
			fmt.Println("ExampleDAO.GetFromID .Error by ID : ", link.ItemID)
			log.Println(exampleError)
			return exampleError, lines, book
		}
		examples = append(examples, *example)
	}

	for _, definition := range definitions {
		relatedForms, error := manager.RelatedDAO.GetRelatedFormsByDefinitionID(ctx, definition.ID)
		if error != nil {
			fmt.Println("RelatedDAO.GetRelatedFormsByDefinitionID .Error by ID : ", definition.ID)
			log.Println(error)
			return error, lines, book
		}

		str, stringError := manager.StringDAO.GetFromID(ctx, definition.StringID)
		if stringError != nil {
			fmt.Println("StringDAO.GetFromID .Error by ID : ", definition.StringID)
			log.Println(stringError)
			return stringError, lines, book
		}

		var sortValue int
		for _, link := range definitionLinks {
			if link.ItemID == definition.ID {
				sortValue = int(link.SortValue)
			}
		}

		// generate line
		for _, form := range relatedForms {
			line := BookLine{
				string:           form.String,
				stringType:       "form",
				partOfSpeech:     form.PartOfSpeech,
				specificType:     form.Form,
				pronunciationIPA: form.Pronunciation,
				definition:       form.Definition,
				stringID:         form.StringID,
				definitionID:     form.DefinitionID,
			}
			lines = append(lines, line)
		}

		line := BookLine{
			string:     str.String,
			stringType: str.Type,
		}
		// add definitions.
		line.partOfSpeech = definition.PartOfSpeech
		line.definition = definition.Definition
		line.specificType = definition.SpecificType
		line.pronunciationIPA = definition.PronunciationIpa
		line.pronunciationIPAWeak = definition.PronunciationIpaWeak
		line.pronunciationIPAOther = definition.PronunciationIpaOther
		line.pronunciationText = definition.PronunciationText
		line.cefrLevel = definition.CefrLevel

		line.stringID = definition.StringID
		line.definitionID = definition.ID

		line.sortValue = sortValue

		// add examples
		var exampleList = []Example{}
		var sortValueList = []int{}
		for _, example := range examples {
			if example.DefinitionID == definition.ID {
				for _, link := range exampleLinks {
					if link.ItemID == example.ID {
						exampleList = append(exampleList, example)
						sortValueList = append(sortValueList, int(link.SortValue))
					}
				}
			}
		}

		sort.Slice(exampleList, func(i, j int) bool {
			return sortValueList[i] < sortValueList[j]
		})

		for i, example := range exampleList {
			if i == 0 {
				line.example1 = example.Content
				line.wordPositions1 = example.WordPositions
				line.example1ID = example.ID
			} else if i == 1 {
				line.example2 = example.Content
				line.wordPositions2 = example.WordPositions
				line.example2ID = example.ID
			} else if i == 2 {
				line.example3 = example.Content
				line.wordPositions3 = example.WordPositions
				line.example3ID = example.ID
			}
		}
		lines = append(lines, line)
	}

	return nil, lines, book
}

func ExportBook(bookID int64, filePath string, manager *dao.Manager) error {
	file, err := os.Create(filePath)
	if err != nil {
		log.Println("failed to open file", err)
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()
	// Using Write

	// first row: word,originWord,form,partOfSpeech,specificType,pronunciation,cefrLevel,definition,example1,wordPositions1,example2,wordPositions2,example3,wordPositions3
	// TODO "\xEF\xBB\xBF" for Excel. utf-8-BOM.
	firstRow := []string{"\xEF\xBB\xBFstring", "stringType", "partOfSpeech", "specificType", "pronunciationIPA", "pronunciationIPAWeak", "pronunciationIPAOther", "pronunciationText", "cefrLevel", "definition",
		"example1", "baseWordPositions1", "example2", "baseWordPositions2", "example3", "baseWordPositions3", "sortValue",
		"stringID", "definitionID", "example1ID", "example2ID", "example3ID"}
	if err := w.Write(firstRow); err != nil {
		log.Println("error writing record to file", err)
		return err
	}

	err, lines, book := getBookData(bookID, manager)
	if err != nil {
		fmt.Println("ExportBook failed, book ID: ", bookID)
		fmt.Println(err)
		return err
	}
	if len(lines) < 1 {
		fmt.Println("ExportBook failed.... no line found")
		return fmt.Errorf("no line found")
	}
	fmt.Printf("Get Book (%s), description: (%s) , line count: %d \n", book.Name, book.Description, len(lines))

	for _, line := range lines {
		row := []string{line.string, line.stringType, line.partOfSpeech, line.specificType, line.pronunciationIPA, line.pronunciationIPAWeak, line.pronunciationIPAOther, line.pronunciationText, line.cefrLevel, line.definition,
			line.example1, line.wordPositions1, line.example2, line.wordPositions2, line.example3, line.wordPositions3, strconv.Itoa(line.sortValue),
			int64ToString(line.stringID), int64ToString(line.definitionID), int64ToString(line.example1ID), int64ToString(line.example2ID), int64ToString(line.example3ID)}
		if err := w.Write(row); err != nil {
			log.Println("error writing record to file", err)
			return err
		}
	}

	fmt.Println("Export success !!")
	return nil
}
