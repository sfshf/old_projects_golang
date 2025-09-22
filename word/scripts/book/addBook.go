package book

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/nextsurfer/word/internal/pkg/dao"
	. "github.com/nextsurfer/word/internal/pkg/model"
)

func stringToInt64(s string) int64 {
	if len(s) == 0 {
		return 0
	}
	i, error := strconv.ParseInt(s, 10, 64)
	if error != nil {
		fmt.Println("stringToInt64 string: ", s)
		fmt.Println("error: ", error)
		return 0
	} else {
		return i
	}
}

func getPronunciationSSML(ipa string) string {
	return "<speak><phoneme alphabet=\"ipa\" ph=\"" + ipa + "\">a</phoneme></speak>"
}

func getPronunciationByText(text string) string {
	return "<speak>" + text + "</speak>"
}

func getDefinitionForForm(word string, form string) string {
	return form + " of " + word
}

func AddBook(csvPath string, dbPath string, name string, description string) {
	book := &Book{
		Name:        name,
		Description: description,
		DownloadURL: "TODO",
	}

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

	readCSVToDatabase(csvPath, book, db, true)
}

func readCSVToDatabase(csvPath string, book *Book, db *gorm.DB, insertNewBook bool) {
	// use transaction
	tx := db.Begin()
	manager := dao.NewManagerWithDB(tx)
	lines, err := readCSV(csvPath)
	if err == nil {
		if insertNewBook {
			err = addBookToDB(book, tx)
			if err != nil {
				fmt.Println(err)
				err = tx.Rollback().Error
				if err != nil {
					fmt.Println("tx.Rollback().Error")
					fmt.Println(err)
				} else {
					fmt.Println("tx.Rollback() success")
				}
				return
			}
		} else {
			// clean old relations to books.
			err = manager.RelatedDAO.DeleteLinksByBookID(context.TODO(), book.ID)
			if err != nil {
				fmt.Println("RelatedDAO.DeleteLinksByBookID().Error")
				fmt.Println(err)
				return
			}
		}

		err := handleData(book.ID, lines, manager)
		if err != nil {
			fmt.Println(err)
			err = tx.Rollback().Error
			if err != nil {
				fmt.Println("tx.Rollback().Error")
				fmt.Println(err)
			} else {
				fmt.Println("tx.Rollback() success")
			}
			return
		}

		err = tx.Commit().Error
		if err != nil {
			fmt.Println("tx.Commit().Error")
			fmt.Println(err)
			tx.Rollback()
			return
		}

		fmt.Println("Success!!!")
	}

}

func readCSV(csvPath string) (ret []*BookLine, retError error) {
	ret = []*BookLine{}
	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatal(err)
		retError = err
	}

	// remember to close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)

	var columnDescriptions []string
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
			retError = err
		}
		if columnDescriptions == nil {
			columnDescriptions = rec
			// TODO fileEncoding="UTF-8-BOM" https://predictivehacks.com/?all-tips=how-to-remove-the-i-appear-in-the-first-column-in-read-csv-in-r
			columnDescriptions[0] = "string"
		} else {
			line, rErr := readLine(columnDescriptions, rec)
			if rErr != nil {
				log.Fatal(rErr)
				retError = rErr
			} else {
				ret = append(ret, line)
			}
		}
	}

	for i := 0; i < len(ret); i++ {
		if ret[i].sortValue == 0 {
			// if no order, use the order of the Excel file.
			ret[i].sortValue = i * 100
			// fmt.Println("line.order: ", line.order)
		}
	}

	return
}

func addBookToDB(book *Book, db *gorm.DB) error {
	bookDAO := dao.NewBookDAO(db)
	err := bookDAO.Create(context.TODO(), book)
	if err != nil {
		fmt.Println("addBookToDB failed")
	} else {
		fmt.Println("addBookToDB success id: ", book.ID)
	}
	return err
}

type BookLine struct {
	string                string
	stringType            string
	partOfSpeech          string
	specificType          string
	pronunciationIPA      string
	pronunciationIPAWeak  string
	pronunciationIPAOther string
	pronunciationText     string
	cefrLevel             string
	definition            string
	sortValue             int
	example1              string
	wordPositions1        string
	example2              string
	wordPositions2        string
	example3              string
	wordPositions3        string

	// word string id
	stringID     int64
	definitionID int64
	example1ID   int64
	example2ID   int64
	example3ID   int64
}

func (l BookLine) Println() {
	fmt.Printf("%s(%s) \n", l.string, l.stringType)
}

func readLine(columnDescriptions []string, record []string) (*BookLine, error) {
	line := BookLine{}
	for i := 0; i < len(record); i++ {
		description := strings.TrimSpace(columnDescriptions[i])
		record[i] = strings.TrimSpace(record[i])
		switch description {
		case "string":
			line.string = record[i]
		case "stringType":
			line.stringType = record[i]
		case "partOfSpeech":
			line.partOfSpeech = record[i]
		case "specificType":
			line.specificType = record[i]
		case "pronunciationIPA":
			line.pronunciationIPA = record[i]
		case "pronunciationIPAWeak":
			line.pronunciationIPAWeak = record[i]
		case "pronunciationIPAOther":
			line.pronunciationIPAOther = record[i]
		case "pronunciationText":
			line.pronunciationText = record[i]
		case "cefrLevel":
			line.cefrLevel = record[i]
		case "definition":
			line.definition = record[i]
		case "example1":
			line.example1 = record[i]
		case "baseWordPositions1":
			line.wordPositions1 = record[i]
		case "example2":
			line.example2 = record[i]
		case "baseWordPositions2":
			line.wordPositions2 = record[i]
		case "example3":
			line.example3 = record[i]
		case "baseWordPositions3":
			line.wordPositions3 = record[i]
		case "sortValue":
			line.sortValue, _ = strconv.Atoi(record[i])
		// new items:
		case "stringID":
			line.stringID = stringToInt64(record[i])
		case "definitionID":
			line.definitionID = stringToInt64(record[i])
		case "example1ID":
			line.example1ID = stringToInt64(record[i])
		case "example2ID":
			line.example2ID = stringToInt64(record[i])
		case "example3ID":
			line.example3ID = stringToInt64(record[i])
		// TODO
		case "myDefinition":
		case "myDefinitionZH":
		case "TODO":
		case "wordPositions1":
		case "myExample1":
		case "translation1":
		case "wordPositions2":
		case "myExample2":
		case "translation2":
		case "wordPositions3":
		case "myExample3":
		case "translation3":
		default:
			return &line, errors.New("can't handle description: " + description)
		}
	}
	return &line, nil
}

func handleData(bookID int64, words []*BookLine, daoManager *dao.Manager) error {
	ctx := context.TODO()

	relatedForms := []*BookLine{}
	lineNum := 0
	for _, line := range words {
		lineNum++
		if line.stringType == "form" {
			relatedForms = append(relatedForms, line)
		} else {
			forms := relatedForms
			relatedForms = []*BookLine{}
			// if == "", it will be empty
			var err error
			stringID := line.stringID
			if line.stringID == 0 {
				if line.stringType == "word" {
					// query string id
					stringID, err = daoManager.StringDAO.GetIDFromWord(ctx, line.string)
					if err != nil {
						fmt.Println("StringDAO.GetIDFromWord")
						fmt.Println(err)
						return err
					}
					if stringID == 0 {
						// create new stringID
						stringID, err = daoManager.StringDAO.CreateWord(ctx, line.string)
						if err != nil {
							println("StringDAO.CreateWord(ctx, word) error")
							return err
						}
					}
				} else if line.stringType == "phrase" {
					// query string id
					stringID, err = daoManager.StringDAO.GetIDByPhrase(ctx, line.string)
					if err != nil {
						fmt.Println("StringDAO.GetIDByPhrase")
						fmt.Println(err)
						return err
					}
					if stringID == 0 {
						// create new stringID
						stringID, err = daoManager.StringDAO.CreatePhrase(ctx, line.string)
						if err != nil {
							println("StringDAO.CreatePhrase(ctx, word) error")
							return err
						}
					}
				}

			}
			if stringID == 0 {
				println("string id can't be zero")
				// println("l")
				fmt.Printf("line: %v\n", line)
				return errors.New("string id can't be zero")
			}
			// log.Printf("TODO word : %s IPA: %s \n", line.string, line.pronunciationIPA)
			definition, definitionError := createOrUpdateDefinition(line, stringID, daoManager)
			if definitionError != nil {
				fmt.Printf("lineNum: %d\n", lineNum)
				fmt.Println("createOrUpdateDefinition failed")
				fmt.Println(definitionError)
				return definitionError
			}

			// check link
			linkError := createRelation(definition.ID, bookID, line.sortValue, daoManager, ctx)
			if linkError != nil {
				fmt.Printf("lineNum: %d\n", lineNum)
				fmt.Println("createRelation failed")
				fmt.Println(linkError)
				return linkError
			}

			formStringList := []string{}
			formStringList = append(formStringList, line.string)
			if line.stringType == "word" {
				// word
				// related forms:
				err3 := handleRelatedForms(line.string, definition.ID, definition.PartOfSpeech, definition.StringID, forms, ctx, daoManager)
				if err3 != nil {
					fmt.Printf("lineNum: %d\n", lineNum)
					fmt.Println("handleRelatedForms failed")
					fmt.Println(err3)
					return err3
				}

				for _, form := range forms {
					contains := false
					for _, s := range formStringList {
						if form.string == s {
							contains = true
						}
					}
					if !contains {
						// distinct
						formStringList = append(formStringList, form.string)
					}
				}
			}

			// Examples:
			isWord := line.stringType == "word"
			// add 3 examples.
			if line.example1 != "" {
				err = handleExample(bookID, isWord, line.example1, 100, line.wordPositions1, line.example1ID, formStringList, definition, daoManager)
				if err != nil {
					fmt.Printf("lineNum: %d\n", lineNum)
					println("Error handleExample 1")
					return err
				}
			}

			if line.example2 != "" {
				err = handleExample(bookID, isWord, line.example2, 200, line.wordPositions2, line.example2ID, formStringList, definition, daoManager)
				if err != nil {
					fmt.Printf("lineNum: %d\n", lineNum)
					println("Error handleExample 2")
					return err
				}
			}

			if line.example3 != "" {
				err = handleExample(bookID, isWord, line.example3, 300, line.wordPositions3, line.example3ID, formStringList, definition, daoManager)
				if err != nil {
					fmt.Printf("lineNum: %d\n", lineNum)
					println("Error handleExample 3")
					return err
				}
			}

		}
	}
	return nil
}

func createOrUpdateDefinition(line *BookLine, wordID int64, daoManager *dao.Manager) (*Definition, error) {
	definition := &Definition{
		ID:           line.definitionID,
		PartOfSpeech: line.partOfSpeech,
		Definition:   line.definition,
		StringID:     wordID,
	}
	if line.stringType == "phrase" {
		// it's hard to say what the part of speech is.
		definition.PartOfSpeech = "phrase"
	}
	if line.specificType != "" {
		definition.SpecificType = line.specificType
	}

	if line.pronunciationIPA != "" {
		definition.PronunciationIpa = line.pronunciationIPA
		definition.PronunciationIpaWeak = line.pronunciationIPAWeak
		definition.PronunciationIpaOther = line.pronunciationIPAOther
		// definition.PronunciationSsml = getPronunciationSSML(line.pronunciationIPA)
	} else if line.pronunciationText != "" {
		definition.PronunciationText = line.pronunciationText
		// definition.PronunciationSsml = getPronunciationByText(line.pronunciationText)
	}

	if line.cefrLevel != "" {
		definition.CefrLevel = line.cefrLevel
	}
	if definition.ID == 0 {
		// need to create
		err := daoManager.DefinitionDAO.Create(context.TODO(), definition)
		if err != nil {
			fmt.Println("definitionDAO.Create")
			fmt.Println(err)
			return definition, err
		}
	} else {
		// check whether it needs to update
		oldDefinition, error := daoManager.DefinitionDAO.GetFromID(context.TODO(), definition.ID)
		if error != nil {
			fmt.Println("definitionDAO.Create")
			fmt.Println(error)
			return oldDefinition, error
		}

		if oldDefinition.PartOfSpeech != definition.PartOfSpeech ||
			oldDefinition.SpecificType != definition.SpecificType ||
			oldDefinition.PronunciationIpa != definition.PronunciationIpa ||
			oldDefinition.PronunciationText != definition.PronunciationText ||
			oldDefinition.PronunciationIpaWeak != definition.PronunciationIpaWeak ||
			oldDefinition.PronunciationIpaOther != definition.PronunciationIpaOther ||
			oldDefinition.CefrLevel != definition.CefrLevel ||
			oldDefinition.Definition != definition.Definition {
			// need to update
			err := daoManager.DefinitionDAO.Update(context.TODO(), definition)
			if err != nil {
				fmt.Println("definitionDAO.Update")
				fmt.Println(err)
				return definition, err
			}

		}
	}
	return definition, nil
}

// We only need create new relation to books because we clean all relations at the start of update.
func createRelation(definitionID int64, bookID int64, sortValue int, daoManager *dao.Manager, ctx context.Context) error {
	// 	// create new one
	err := daoManager.RelatedDAO.CreateRelationForDefinition(ctx, definitionID, bookID, sortValue)
	if err != nil {
		fmt.Println("relatedDAO.CreateRelationForDefinition")
		fmt.Println(err)
		return err
	}
	return nil
}

func handleRelatedForms(word string, definitionID int64, partOfSpeech string, wordID int64, relatedForms []*BookLine, ctx context.Context, daoManager *dao.Manager) error {
	currentForms, err := daoManager.RelatedDAO.GetRelatedFormsByDefinitionID(ctx, definitionID)
	if err != nil {
		println("GetRelatedFormsByDefinitionID error!!!")
		return err
	}
	list := []*dao.RelatedForm{}
	for _, formLine := range relatedForms {
		form := &dao.RelatedForm{
			String:           formLine.string,
			Form:             formLine.specificType,
			PartOfSpeech:     partOfSpeech,
			Definition:       getDefinitionForForm(word, formLine.specificType),
			BaseStringID:     wordID,
			BaseDefinitionID: definitionID,
		}
		if formLine.pronunciationIPA != "" {
			form.Pronunciation = formLine.pronunciationIPA
			// form.PronunciationSSML = getPronunciationSSML(formLine.pronunciationIPA)
		}
		list = append(list, form)
	}

	if len(currentForms) == 0 {
		// insert new
		if len(list) > 0 {
			for _, form := range list {
				error := daoManager.RelatedDAO.CreateRelatedForm(ctx, form)
				if error != nil {
					println("relatedDAO.CreateForm(ctx, form) error")
					return error
				}
			}
		}
	} else {
		// check and try insert or update or delete
		for _, oldForm := range currentForms {
			needToDelete := true
			for _, newForm := range list {
				if newForm.Form == oldForm.Form {
					needToDelete = false
					if newForm.String != oldForm.String {
						// update string
						oldForm.String = newForm.String
						error := daoManager.RelatedDAO.UpdateRelatedForm(ctx, &oldForm)
						if error != nil {
							println("relatedDAO.UpdateRelatedForm(ctx, oldForm) error")
							return error
						}
					}
					if newForm.PartOfSpeech != oldForm.PartOfSpeech ||
						newForm.Definition != oldForm.Definition ||
						newForm.Pronunciation != oldForm.Pronunciation {
						oldForm.PartOfSpeech = newForm.PartOfSpeech
						oldForm.Definition = newForm.Definition
						oldForm.Pronunciation = newForm.Pronunciation
						// oldForm.PronunciationSSML = newForm.PronunciationSSML

						error := daoManager.RelatedDAO.UpdateRelatedDefinition(ctx, &oldForm)
						if error != nil {
							println("relatedDAO.UpdateRelatedDefinition(ctx, oldForm) error")
							return error
						}
					}
				}
			}
			if needToDelete {
				error := daoManager.RelatedDAO.DeleteRelatedForm(ctx, &oldForm)
				if error != nil {
					println("relatedDAO.DeleteRelatedForm(ctx, oldForm) error")
					return error
				}
			}
		}

		insertList := []*dao.RelatedForm{}
		for _, newForm := range list {
			needToinsert := true
			for _, oldForm := range currentForms {
				if newForm.Form == oldForm.Form {
					needToinsert = false
				}
			}
			if needToinsert {
				insertList = append(insertList, newForm)
			}
		}
		if len(insertList) > 0 {
			for _, newForm := range insertList {
				error := daoManager.RelatedDAO.CreateRelatedForm(ctx, newForm)
				if error != nil {
					println("relatedDAO.CreateRelatedForm(ctx, newForm) error")
					return error
				}
			}
		}
	}
	return nil
}

func handleExample(bookID int64, isWord bool, example string, sortValue int, wordPositions string, exampleID int64, forms []string, definition *Definition, daoManager *dao.Manager) error {
	example1 := &Example{
		ID:           exampleID,
		StringID:     definition.StringID,
		DefinitionID: definition.ID,
		Content:      example,
	}
	if wordPositions == "" {
		str := forms[len(forms)-1]
		// get
		if isWord {
			wordPositions = findWordPositionFromExample(example, forms)
			if wordPositions == "" {
				fmt.Printf("can not find word 【 %s 】 (forms : %v) in the example Example: %s \n", str, forms, example)
				return errors.New("cant find wordPosition")
			}
		} else {
			// try to use find the string of phrase in the example.
			wordPositions = findWordPositionFromExample(example, forms)
			if wordPositions == "" {
				fmt.Printf("can not find phrase or collocation 【 %s 】 in the example Example: %s \n", str, example)
				return errors.New("cant find wordPosition")
			}
			// phrase or collocation have to provide positions manually
			// return errors.New("phrase or collocation have to provide positions manually")
		}
	} else {
		// check word poisitions.
		intStrings := strings.Split(wordPositions, ",")
		if len(intStrings)%2 != 0 || len(intStrings) < 2 {
			fmt.Printf("given wordPositions ( %s ) is wrong in the example Example: %s \n", wordPositions, example)
			return errors.New("cant find wordPosition")
		}
		newWordPositions := ""
		for i := 0; i < len(intStrings); i++ {
			indexOrLength, atoiError := strconv.Atoi(strings.TrimSpace(intStrings[i]))
			if atoiError != nil || indexOrLength < 0 {
				println(atoiError)
				fmt.Printf("given wordPositions ( %s ) is wrong (can't parse int) in the example Example: %s \n", wordPositions, example)
				return errors.New("cant find wordPosition")
			}
			newWordPositions = newWordPositions + strconv.Itoa(indexOrLength) + ","
		}
		wordPositions = newWordPositions[:len(newWordPositions)-1]
	}

	example1.WordPositions = wordPositions

	// create or update
	if example1.ID != 0 {
		// check update
		oldExample, error := daoManager.ExampleDAO.GetFromID(context.TODO(), example1.ID)
		if error != nil {
			fmt.Println("exampleDAO.GetFromID")
			fmt.Println(error)
			return error
		}

		if oldExample.Content != example1.Content ||
			oldExample.WordPositions != example1.WordPositions {
			// update
			if oldExample.StringID != example1.StringID || oldExample.DefinitionID != example1.DefinitionID {
				fmt.Printf("oldExample.WordID != example1.WordID  exampleID %d \n", example1.ID)
				return errors.New("ID error : oldExample.WordID != example1.WordID")
			}
			error = daoManager.ExampleDAO.Update(context.TODO(), example1)
			if error != nil {
				fmt.Println("exampleDAO.Update")
				fmt.Println(error)
				return error
			}
		}
	} else {
		// create
		err := daoManager.ExampleDAO.Create(context.TODO(), example1)
		if err != nil {
			fmt.Println("exampleDAO.Create")
			fmt.Println(err)
			return err
		}

	}

	// create link, all old link will be deleted during updating.
	err := daoManager.RelatedDAO.CreateRelatationForExample(context.TODO(), example1.ID, bookID, sortValue)
	if err != nil {
		fmt.Println("relatedDAO.CreateRelatationForExample")
		fmt.Println(err)
		return err
	}

	return nil
}

func findWordPositionFromExample(example string, wordForms []string) string {
	wordPositions := ""
	for _, form := range wordForms {
		re := regexp.MustCompile(`(?i)(^|\s|[^\w\s])(` + form + `)($|\s|[^\w\s])`)
		if strings.HasPrefix(form, "'") {
			// if starts with ', it may be a contraction, like 's, 're, 'll, 'd, 've, 'm, so we need to remove the space or punctuation before the word.
			re = regexp.MustCompile(`(?i)([\w])(` + form + `)($|\s|[^\w\s])`)
		}
		matched := re.FindAllIndex([]byte(example), -1)

		for i := 0; i < len(matched); i++ {
			positions := matched[i]
			positions[1] -= positions[0]
			if positions[1] == len(form)+1 {
				// remove space or punctuation
				matchedStr := example[positions[0]:(positions[0] + positions[1])]
				index := strings.Index(strings.ToLower(matchedStr), strings.ToLower(form))
				positions[1] -= 1
				if index == 1 {
					positions[0] += 1
				}
			} else if positions[1] == len(form)+2 {
				positions[0] += 1
				positions[1] -= 2
			}
			// else {
			// 	// positions[1] == len(form)
			// }
			// fmt.Println(positions)

			if positions[1] != len(form) {
				fmt.Printf("fatal error in findWordPositionFromExample positions[1] != len(form)  : form: %s , positions[1]: %d ", form, positions[1])
				return ""
			}
			if wordPositions != "" {
				wordPositions += ","
			}
			wordPositions += fmt.Sprintf("%d,%d", positions[0], positions[1])
		}
	}
	return wordPositions
}

// / TODO update test
func UpdateBook(csvPath string, dbPath string, name string, description string, bookID int64) {
	book := &Book{
		ID:          bookID,
		Name:        name,
		Description: description,
	}

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

	readCSVToDatabase(csvPath, book, db, false)
}

func UpdateDownloadURL(bookID int64, downloadURL string, dbPath string) {
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

	bookDAO := dao.NewBookDAO(db)

	book, error := bookDAO.GetFromID(context.TODO(), bookID)
	if error != nil {
		fmt.Println("bookDAO.GetFromID() error")
		log.Fatal(error)
	}
	book.DownloadURL = downloadURL
	error = bookDAO.Update(context.TODO(), book)
	if error != nil {
		fmt.Println("bookDAO.Update() error")
		log.Fatal(error)
	}
	println("success!")
}
