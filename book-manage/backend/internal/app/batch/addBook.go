package batch

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"regexp"
	"strconv"
	"strings"

	"github.com/nextsurfer/book-manage-api/internal/app/dao"
	. "github.com/nextsurfer/book-manage-api/internal/app/model"
	"gorm.io/gorm"
)

var logger Logger

func stringToInt64(s string) int64 {
	if len(s) == 0 {
		return 0
	}
	i, error := strconv.ParseInt(s, 10, 64)
	if error != nil {
		logger.InfoPrint("stringToInt64 string: ", s)
		logger.InfoPrint("error: ", error)
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

func deleteUselessDefinitionAndExamples(bookID int64, deletedLinks []RelatedBook, manager *dao.Manager) error {
	logger.InfoPrint("delete Useless Definitions And Examples")
	for index, link := range deletedLinks {

		if link.ItemType == "definition" {
			// check whether it is used now
			newLink, err := manager.RelatedDAO.GetRelatedBookForDefinition(context.TODO(), link.ItemID, bookID)
			if err != nil {
				logger.InfoPrint("deleteUselessDefinitionAndExamples RelatedDAO.GetRelatedBookForDefinition")
				logger.InfoPrint(err)
				return err
			}
			// if not, delete it
			if newLink == nil {
				logger.InfoPrint("delete useless definition, ID: ", link.ItemID)
				err := manager.DefinitionDAO.DeleteByID(context.TODO(), link.ItemID)
				if err != nil {
					logger.InfoPrint("deleteUselessDefinitionAndExamples DefinitionDAO.Delete")
					logger.InfoPrint(err)
					return err
				}
			}
		} else if link.ItemType == "example" {
			// check whether it is used now
			newLink, err := manager.RelatedDAO.GetRelatedBookForExample(context.TODO(), link.ItemID, bookID)
			if err != nil {
				logger.InfoPrint("deleteUselessDefinitionAndExamples RelatedDAO.GetRelatedBookForExample")
				logger.InfoPrint(err)
				return err
			}
			// if not, delete it
			if newLink == nil {
				logger.InfoPrint("delete useless example, ID: ", link.ItemID)
				err := manager.ExampleDAO.DeleteByID(context.TODO(), link.ItemID)
				if err != nil {
					logger.InfoPrint("deleteUselessDefinitionAndExamples ExampleDAO.Delete")
					logger.InfoPrint(err)
					return err
				}
			}

		}
		if index%100 == 8 {
			logger.InfoPrint("checking ", index, " / ", len(deletedLinks))
		}
	}
	return nil
}

func readCSVToDatabase(isBackup bool, reader io.Reader, book *Book, dbmanager *dao.Manager, insertNewBook bool, strictMode bool) error {
	// use transaction
	// tx := db.Begin()
	tx, manager := dbmanager.Transaction()

	sortValueStartAt := 0
	if !insertNewBook && !strictMode {
		// get max sort value, new item will be added after the max sort value.
		maxSortValue, err := manager.RelatedDAO.GetMaxSortValue(context.TODO(), book.ID)
		if err != nil {
			logger.InfoPrint("GetMaxSortValue error!!!")
			return err
		}
		sortValueStartAt = maxSortValue + 100
	}

	logger.VerbosePrint("readCSVToDatabase")
	lines, err := readCSV(reader, sortValueStartAt)
	if err == nil {
		if insertNewBook {
			logger.VerbosePrint("add book to db")
			err = addBookToDB(book, tx)
			if err != nil {
				if err := tx.Rollback().Error; err != nil {
					logger.InfoPrint("tx.Rollback().Error: %s", err.Error())
				} else {
					logger.InfoPrint("tx.Rollback() success")
				}

				return err
			}
		} else if strictMode {
			// clean old relations to books. only use the data from the csv file.
			var links []RelatedBook
			err = manager.RelatedDAO.DeleteLinksByBookID(context.TODO(), book.ID, &links)
			if err != nil {
				if err := tx.Rollback().Error; err != nil {
					logger.InfoPrint("tx.Rollback().Error: %s", err.Error())
				} else {
					logger.InfoPrint("tx.Rollback() success")
				}
				return fmt.Errorf("RelatedDAO.DeleteLinksByBookID().Error: %s", err.Error())
			} else {
				logger.InfoPrint("Restrict Mode delete all items in the book. count: ", len(links))
				// remove all definitions and examples
				err = deleteUselessDefinitionAndExamples(book.ID, links, manager)
				if err != nil {
					if err := tx.Rollback().Error; err != nil {
						logger.InfoPrint("tx.Rollback().Error: %s", err.Error())
					} else {
						logger.InfoPrint("tx.Rollback() success")
					}
					return fmt.Errorf("deleteUselessDefinitionAndExamples().Error: %s", err.Error())
				}
			}
		}

		logger.VerbosePrint("handleData")
		err := handleData(isBackup, book.ID, lines, manager)
		if err != nil {
			logger.InfoPrint(err)
			if err := tx.Rollback().Error; err != nil {
				logger.InfoPrint("tx.Rollback().Error: %s", err.Error())
			} else {
				logger.InfoPrint("tx.Rollback() success")
			}
			return err
		}
		if !insertNewBook {
			// update time
			// fmt.Println("UpdateDownloadPath ")
			err = manager.BookDAO.UpdateDownloadPath(context.TODO(), book.ID)
			if err != nil {
				logger.InfoPrint(err)
				if err := tx.Rollback().Error; err != nil {
					logger.InfoPrint("tx.Rollback().Error: %s", err.Error())
				} else {
					logger.InfoPrint("tx.Rollback() success")
				}
				return err
			}
		}

		err = tx.Commit().Error
		// fmt.Println("tx.Commit()")
		if err != nil {
			logger.InfoPrint("tx.Commit().Error:", err.Error())
			tx.Rollback()
			return err
		}

		logger.InfoPrint("Success!!!")
	} else {
		return err
	}

	return nil
}

func readCSV(reader io.Reader, sortValueStartAt int) (ret []*BookLine, retError error) {
	ret = []*BookLine{}

	// read csv values using csv.Reader
	csvReader := csv.NewReader(reader)

	var columnDescriptions []string
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.InfoPrint(err)
			retError = err
			return
		}
		if columnDescriptions == nil {
			columnDescriptions = rec
			// TODO fileEncoding="UTF-8-BOM" https://predictivehacks.com/?all-tips=how-to-remove-the-i-appear-in-the-first-column-in-read-csv-in-r
			columnDescriptions[0] = "string"
		} else {
			line, rErr := readLine(columnDescriptions, rec)
			if rErr != nil {
				logger.InfoPrint(rErr)
				retError = rErr
				return
			} else {
				ret = append(ret, line)
			}
		}
	}

	for i := 0; i < len(ret); i++ {
		if ret[i].sortValue == 0 {
			// if no order, use the order of the Excel file.
			ret[i].sortValue = i*100 + sortValueStartAt
			// fmt.Println("line.order: ", line.order)
		}
	}
	logger.InfoPrint("read csv success, total lines: ", len(ret))
	return
}

func addBookToDB(book *Book, db *gorm.DB) error {
	bookDAO := dao.NewBookDAO(db)
	err := bookDAO.Create(context.TODO(), book)
	if err != nil {
		logger.InfoPrint("addBookToDB failed")
	} else {
		logger.InfoPrint("addBookToDB success BookID: ", book.ID)
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

func handleData(isBackup bool, bookID int64, words []*BookLine, daoManager *dao.Manager) error {
	ctx := context.TODO()

	relatedForms := []*BookLine{}
	lineNum := 0
	wordCount := 0
	for _, line := range words {
		logger.Progress(lineNum, len(words))
		lineNum++
		if line.stringType == "form" {
			relatedForms = append(relatedForms, line)
		} else {
			forms := relatedForms
			relatedForms = []*BookLine{}
			// if == "", it will be empty
			var err error
			stringID := line.stringID
			if line.stringID != 0 {
				// check whether the string is correct.
				stringDO, err := daoManager.StringDAO.GetNullableItemByID(ctx, line.stringID)
				if err != nil {
					logger.InfoPrint("daoManager.StringDAO.GetFromID(ctx, line.stringID) error, cannot find string by string ID: ", line.stringID, ", at Line: ", lineNum+1)
					logger.InfoPrint(err)
					return err
				} else if stringDO == nil {
					// data is deleted, we need to create a new one.
					err := daoManager.StringDAO.CreateWithID(ctx, line.string, line.stringType, line.stringID)
					if err != nil {
						logger.InfoPrint("daoManager.StringDAO.CreateWithID(ctx, line.string, line.stringType, line.stringID) error, cannot create string by string ID: ", line.stringID, ", at Line: ", lineNum+1)
						logger.InfoPrint(err)
						return err
					}
				} else if stringDO.String != line.string {
					if isBackup {
						stringDO.String = line.string
						if err := daoManager.StringDAO.Update(context.TODO(), stringDO); err != nil {
							return err
						}
					} else {
						logger.InfoPrint("can't change an existing string to another string (You need to make sure string id is empty), at Line: ", lineNum+1)
						return errors.New("can't change an existing string")
					}
				}
			} else {
				if line.stringType == "word" {
					// query string id
					stringID, err = daoManager.StringDAO.GetIDFromWord(ctx, line.string)
					if err != nil {
						logger.InfoPrint("StringDAO.GetIDFromWord")
						logger.InfoPrint(err)
						return err
					}
					if stringID == 0 {
						// create new stringID
						stringID, err = daoManager.StringDAO.CreateWord(ctx, line.string)
						if err != nil {
							logger.InfoPrint("StringDAO.CreateWord(ctx, word) error")
							logger.InfoPrint(err)
							return err
						}
					}
				} else if line.stringType == "phrase" {
					// query string id
					stringID, err = daoManager.StringDAO.GetIDByPhrase(ctx, line.string)
					if err != nil {
						logger.InfoPrint("StringDAO.GetIDByPhrase")
						logger.InfoPrint(err)
						return err
					}
					if stringID == 0 {
						// create new stringID
						stringID, err = daoManager.StringDAO.CreatePhrase(ctx, line.string)
						if err != nil {
							logger.InfoPrint("StringDAO.CreatePhrase(ctx, word) error")
							logger.InfoPrint(err)
							return err
						}
					}
				}
			}
			if stringID == 0 {
				logger.InfoPrint("Problem happend at Line: ", lineNum+1)
				return errors.New("string id can't be zero")
			}

			definition, definitionError := createOrUpdateDefinition(line, stringID, daoManager)
			if definitionError != nil {
				logger.InfoPrint("Problem happend at Line: ", lineNum+1)
				logger.InfoPrint("createOrUpdateDefinition failed")
				logger.InfoPrint(definitionError)
				return definitionError
			}

			// check link
			linkError := createOrUpdateRelation(definition.ID, bookID, line.sortValue, daoManager, ctx)
			if linkError != nil {
				logger.InfoPrint("Problem happend at Line: ", lineNum+1)
				logger.InfoPrint("createRelation failed")
				logger.InfoPrint(linkError)
				return linkError
			}

			formStringList := []string{}
			formStringList = append(formStringList, line.string)
			if line.stringType == "word" {
				// word
				// related forms:
				err3 := handleRelatedForms(line.string, definition.ID, definition.PartOfSpeech, definition.StringID, forms, ctx, daoManager)
				if err3 != nil {
					logger.InfoPrint("Problem happend at Line: ", lineNum+1)
					logger.InfoPrint("handleRelatedForms failed")
					logger.InfoPrint(err3)
					return err3
				}

				for _, form := range forms {
					contains := false
					for _, s := range formStringList {

						if form.string == s ||
							// ToLower: "I" and "i" are the same.
							// https://dictionary.cambridge.org/us/dictionary/english/champagne "champagne" and "Champagne" are the same.
							// strings.ToLower(form.string) == strings.ToLower(s)
							strings.EqualFold(form.string, s) {
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
					logger.InfoPrint("Problem happend at Line: ", lineNum+1)
					logger.InfoPrint("Error handleExample 1")
					return err
				}
			} else {
				logger.InfoPrint("Every item must have at least one example. Line number: ", lineNum+1)
				return errors.New("every item must have at least one example")
			}

			if line.example2 != "" {
				err = handleExample(bookID, isWord, line.example2, 200, line.wordPositions2, line.example2ID, formStringList, definition, daoManager)
				if err != nil {
					logger.InfoPrint("Problem happend at Line: ", lineNum+1)
					logger.InfoPrint("Error handleExample 2")
					return err
				}
			}

			if line.example3 != "" {
				err = handleExample(bookID, isWord, line.example3, 300, line.wordPositions3, line.example3ID, formStringList, definition, daoManager)
				if err != nil {
					logger.InfoPrint("Problem happend at Line: ", lineNum+1)
					logger.InfoPrint("Error handleExample 3")
					return err
				}
			}

			wordCount++
			logger.CountDefinition(wordCount)
			logger.VerbosePrint("finish word ", line.string, " ; word count: ", wordCount)
		}
	}
	logger.InfoPrint("finish handleData, total words: ", wordCount)
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

	if definition.ID != 0 {
		// check whether it needs to update
		oldDefinition, error := daoManager.DefinitionDAO.GetNullableItemByID(context.TODO(), definition.ID)
		if error != nil {
			logger.InfoPrint("definitionDAO.GetFromID")
			logger.InfoPrint(error)
			return nil, error
		} else if oldDefinition == nil {
			// not exist, need to create
			err := daoManager.DefinitionDAO.CreateWithID(context.TODO(), definition)
			if err != nil {
				logger.InfoPrint("definitionDAO.CreateWithID")
				logger.InfoPrint(err)
				return definition, err
			}
		} else if oldDefinition.PartOfSpeech != definition.PartOfSpeech ||
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
				logger.InfoPrint("definitionDAO.Update")
				logger.InfoPrint(err)
				return definition, err
			}
		}
	} else {
		// need to create
		err := daoManager.DefinitionDAO.Create(context.TODO(), definition)
		if err != nil {
			logger.InfoPrint("definitionDAO.Create")
			logger.InfoPrint(err)
			return definition, err
		}
	}
	return definition, nil
}

func createOrUpdateRelation(definitionID int64, bookID int64, sortValue int, daoManager *dao.Manager, ctx context.Context) error {
	relation, err := daoManager.RelatedDAO.GetRelatedBookForDefinition(ctx, definitionID, bookID)
	if err != nil {
		logger.InfoPrint("relatedDAO.GetRelatedBookForDefinition")
		logger.InfoPrint(err)
		return err
	}
	if relation != nil {
		// update
		if relation.SortValue != int32(sortValue) {
			relation.SortValue = int32(sortValue)
			err := daoManager.RelatedDAO.UpdateRelatedBook(ctx, relation)
			if err != nil {
				logger.InfoPrint("relatedDAO.Update")
				logger.InfoPrint(err)
				return err
			}
		}
	} else {
		// create new one
		err := daoManager.RelatedDAO.CreateRelationForDefinition(ctx, definitionID, bookID, sortValue)
		if err != nil {
			logger.InfoPrint("relatedDAO.CreateRelationForDefinition")
			logger.InfoPrint(err)
			return err
		}
	}
	return nil
}

func createOrUpdateExampleBookRelation(exampleID int64, bookID int64, sortValue int, daoManager *dao.Manager, ctx context.Context) error {
	relation, err := daoManager.RelatedDAO.GetRelatedBookForExample(ctx, exampleID, bookID)
	if err != nil {
		logger.InfoPrint("relatedDAO.GetRelatedBookForDefinition")
		logger.InfoPrint(err)
		return err
	}
	if relation != nil {
		// update
		if relation.SortValue != int32(sortValue) {
			relation.SortValue = int32(sortValue)
			err := daoManager.RelatedDAO.UpdateRelatedBook(ctx, relation)
			if err != nil {
				logger.InfoPrint("relatedDAO.Update")
				logger.InfoPrint(err)
				return err
			}
		}
	} else {
		// create new one
		err := daoManager.RelatedDAO.CreateRelatationForExample(ctx, exampleID, bookID, sortValue)
		if err != nil {
			logger.InfoPrint("relatedDAO.CreateRelationForDefinition")
			logger.InfoPrint(err)
			return err
		}
	}
	return nil
}

func handleRelatedForms(word string, definitionID int64, partOfSpeech string, wordID int64, relatedForms []*BookLine, ctx context.Context, daoManager *dao.Manager) error {
	currentForms, err := daoManager.RelatedDAO.GetRelatedFormsByDefinitionID(ctx, definitionID)
	if err != nil {
		logger.InfoPrint("GetRelatedFormsByDefinitionID error!!!")
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
					logger.InfoPrint("relatedDAO.CreateForm(ctx, form) error")
					return error
				}
			}
		}
	} else {
		// delete repeated forms from currentForms
		uniqueFormMap := make(map[string]bool)
		uniqueForms := []dao.RelatedForm{}
		for _, form := range currentForms {
			if uniqueFormMap[form.Form+form.String] {
				// delete repeated forms
				error := daoManager.RelatedDAO.DeleteRelatedForm(ctx, &form)
				if error != nil {
					logger.InfoPrint("delete repeated forms: relatedDAO.DeleteRelatedForm(ctx, form) error")
					return error
				}
			} else {
				uniqueFormMap[form.Form+form.String] = true
				uniqueForms = append(uniqueForms, form)
			}
		}
		currentForms = uniqueForms

		// check and try insert or update or delete
		for _, oldForm := range currentForms {
			needToDelete := true
			for _, newForm := range list {
				// be: simple present: are, is ,am,
				if newForm.Form == oldForm.Form && newForm.String == oldForm.String {
					needToDelete = false
					if newForm.PartOfSpeech != oldForm.PartOfSpeech ||
						newForm.Definition != oldForm.Definition ||
						newForm.Pronunciation != oldForm.Pronunciation {
						oldForm.PartOfSpeech = newForm.PartOfSpeech
						oldForm.Definition = newForm.Definition
						oldForm.Pronunciation = newForm.Pronunciation
						// oldForm.PronunciationSSML = newForm.PronunciationSSML

						error := daoManager.RelatedDAO.UpdateRelatedDefinition(ctx, &oldForm)
						if error != nil {
							logger.InfoPrint("relatedDAO.UpdateRelatedDefinition(ctx, oldForm) error")
							return error
						}
					}
				}
			}
			if needToDelete {
				error := daoManager.RelatedDAO.DeleteRelatedForm(ctx, &oldForm)
				if error != nil {
					logger.InfoPrint("relatedDAO.DeleteRelatedForm(ctx, oldForm) error")
					return error
				}
			}
		}

		insertList := []*dao.RelatedForm{}
		for _, newForm := range list {
			needToinsert := true
			for _, oldForm := range currentForms {
				if newForm.Form == oldForm.Form && newForm.String == oldForm.String {
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
					logger.InfoPrint("relatedDAO.CreateRelatedForm(ctx, newForm) error")
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
		fmt.Println(example, forms)
		// get
		if isWord {
			wordPositions = FindWordPositionFromExample(example, forms)
			if wordPositions == "" {
				msg := fmt.Sprintf("can not find word 【 %s 】 (forms : %v) in the example Example: %s \n", str, forms, example)
				logger.InfoPrint(msg)
				return errors.New(msg)
			}
		} else {
			// try to use find the string of phrase in the example.
			wordPositions = FindWordPositionFromExample(example, forms)
			if wordPositions == "" {
				msg := fmt.Sprintf("can not find phrase or collocation 【 %s 】 in the example Example: %s \n", str, example)
				logger.InfoPrint(msg)
				return errors.New(msg)
			}
			// phrase or collocation have to provide positions manually
			// return errors.New("phrase or collocation have to provide positions manually")
		}
	} else {
		// check word poisitions.
		intStrings := strings.Split(wordPositions, ",")
		if len(intStrings)%2 != 0 || len(intStrings) < 2 {
			msg := fmt.Sprintf("given wordPositions ( %s ) is wrong in the example Example: %s \n", wordPositions, example)
			logger.InfoPrint(msg)
			return errors.New(msg)
		}
		newWordPositions := ""
		for i := 0; i < len(intStrings); i++ {
			indexOrLength, atoiError := strconv.Atoi(strings.TrimSpace(intStrings[i]))
			if atoiError != nil || indexOrLength < 0 {
				msg := fmt.Sprintf("given wordPositions ( %s ) is wrong (can't parse int) in the example Example: %s \n", wordPositions, example)
				logger.InfoPrint(msg)
				return errors.New(msg)
			}
			newWordPositions = newWordPositions + strconv.Itoa(indexOrLength) + ","
		}
		wordPositions = newWordPositions[:len(newWordPositions)-1]
	}

	example1.WordPositions = wordPositions

	// create or update
	if example1.ID != 0 {
		// check update
		oldExample, error := daoManager.ExampleDAO.GetNullableItemByID(context.TODO(), example1.ID)
		if error != nil {
			logger.InfoPrint("exampleDAO.GetFromID")
			logger.InfoPrint(error)
			return error
		}
		if oldExample == nil {
			// create a new one
			error := daoManager.ExampleDAO.CreateWithID(context.TODO(), example1)
			if error != nil {
				logger.InfoPrint("exampleDAO.CreateWithID")
				logger.InfoPrint(error)
				return error
			}
		} else if oldExample.Content != example1.Content ||
			oldExample.WordPositions != example1.WordPositions {
			// update
			if oldExample.StringID != example1.StringID || oldExample.DefinitionID != example1.DefinitionID {
				msg := fmt.Sprintf("oldExample.WordID != example1.WordID  exampleID %d \n", example1.ID)
				logger.InfoPrint(msg)
				return errors.New(msg)
			}
			error = daoManager.ExampleDAO.Update(context.TODO(), example1)
			if error != nil {
				logger.InfoPrint("exampleDAO.Update")
				logger.InfoPrint(error)
				return error
			}
		}
	} else {
		// create
		err := daoManager.ExampleDAO.Create(context.TODO(), example1)
		if err != nil {
			logger.InfoPrint("exampleDAO.Create")
			logger.InfoPrint(err)
			return err
		}

	}

	err := createOrUpdateExampleBookRelation(example1.ID, bookID, sortValue, daoManager, context.TODO())
	if err != nil {
		logger.InfoPrint("createOrUpdateExampleBookRelation")
		logger.InfoPrint(err)
		return err
	}

	return nil
}

func FindWordPositionFromExample(example string, wordForms []string) string {
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
				msg := fmt.Sprintf("fatal error in FindWordPositionFromExample positions[1] != len(form)  : form: %s , positions[1]: %d ", form, positions[1])
				if logger != nil {
					logger.InfoPrint(msg)
				}
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

func AddBook(csvPath *multipart.FileHeader, name string, description string, _logger Logger, manager *dao.Manager, user string) (err error) {
	book := &Book{
		Name:        name,
		Description: description,
		DownloadURL: "TODO",
	}
	logger = _logger
	// db := tools.MysqlDB()
	// logger.start()

	f, err := csvPath.Open()
	if err != nil {
		logger.InfoPrint(err)
		return
	}
	// remember to close the file at the end of the program
	defer f.Close()

	err = readCSVToDatabase(false, f, book, manager, true, false)
	if err != nil {
		logger.InfoPrint(err)
		logger.Complete(err, user, "AddBook", book.ID)
		return err
	}
	logger.InfoPrint("Book ID: ", book.ID)
	logger.Complete(nil, user, "AddBook", book.ID)
	return nil
}

func UpdateBook(csvPath *multipart.FileHeader, bookID int64, _logger Logger, manager *dao.Manager, user string, strictMode bool) (err error) {
	book := &Book{
		ID: bookID,
	}

	logger = _logger

	// db := tools.MysqlDB()
	// logger.start()

	f, err := csvPath.Open()
	if err != nil {
		logger.InfoPrint(err)
		return
	}
	// remember to close the file at the end of the program
	defer f.Close()

	err = readCSVToDatabase(false, f, book, manager, false, strictMode)
	if err != nil {
		logger.InfoPrint(err)
		logger.Complete(err, user, "UpdateBook", bookID)
		return err
	}
	logger.Complete(nil, user, "UpdateBook", bookID)
	return nil
}

// func UpdateDownloadURL(bookID int64, downloadURL string, dbPath string) {
// 	db := tools.MysqlDB()

// 	bookDAO := dao.NewBookDAO(db)

// 	book, error := bookDAO.GetFromID(context.TODO(), bookID)
// 	if error != nil {
// 		logger.InfoPrint("bookDAO.GetFromID() error")
// 		log.Println(error)
// 	}
// 	book.DownloadURL = downloadURL
// 	error = bookDAO.Update(context.TODO(), book)
// 	if error != nil {
// 		logger.InfoPrint("bookDAO.Update() error")
// 		log.Println(error)
// 	}
// 	println("success!")
// }
