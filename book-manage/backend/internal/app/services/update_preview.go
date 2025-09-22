package services

import (
	"context"
	"errors"
	"strings"

	"github.com/nextsurfer/book-manage-api/api"
	"github.com/nextsurfer/book-manage-api/internal/app/batch"
	"github.com/nextsurfer/book-manage-api/internal/app/dao"
	"github.com/nextsurfer/book-manage-api/internal/app/model"
	"gorm.io/gorm"
)

func (s *BookService) UpdatePreviewString(req *api.UpdatePreviewRequest, manager *dao.Manager, user string) (int64, error) {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "UpdatePreviewString", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "UpdatePreviewString", req.BookID, req.DefinitionID)
		}
	}()

	// rule checking
	stringTxt := strings.TrimSpace(req.String)
	if err = CheckEmpty(stringTxt); err != nil {
		return 0, err
	}

	// get definition
	definition, err := manager.DefinitionDAO.GetFromID(context.TODO(), req.DefinitionID)
	if err != nil {
		return 0, err
	}
	// check string id
	if definition.StringID != req.StringID {
		err = errors.New("invalid string id")
		return 0, err
	}
	// get string
	oldStr, err := manager.StringDAO.GetFromID(context.TODO(), definition.StringID)
	if err != nil {
		return 0, err
	}
	// no need to update if request's string is equal to that in the database
	if oldStr.String == stringTxt {
		return req.StringID, nil
	}

	var stringID int64

	reqStr, err := manager.StringDAO.GetFromString(context.TODO(), stringTxt)
	if err != nil {
		return 0, err
	}
	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)
		// 如果修改为已有的string，则只需要修改stringID即可。
		if reqStr.ID > 0 {
			if err := manager.DefinitionDAO.Update(context.TODO(), &model.Definition{
				ID:       definition.ID,
				StringID: reqStr.ID,
			}); err != nil {
				return err
			}
			stringID = reqStr.ID
		} else {
			another, err := manager.DefinitionDAO.GetFromStringIDExclude(context.TODO(), req.StringID, definition.ID)
			if err != nil {
				return err
			}
			// 如果有多个词条对应string， 则创建一个新的string
			if another.ID > 0 {
				newStr := &model.String{
					String:       stringTxt,
					Type:         oldStr.Type,
					BaseStringID: oldStr.BaseStringID,
				}
				if err := manager.StringDAO.Create(context.TODO(), newStr); err != nil {
					return err
				}
				if err := manager.DefinitionDAO.Update(context.TODO(), &model.Definition{
					ID:       definition.ID,
					StringID: newStr.ID,
				}); err != nil {
					return err
				}
				stringID = newStr.ID
			} else { // 如果当前只有一个词条对应 string， 则直接修改string
				if err := manager.StringDAO.Update(context.TODO(), &model.String{
					ID:     req.StringID,
					String: stringTxt,
				}); err != nil {
					return err
				}
				stringID = req.StringID
			}
		}
		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return stringID, nil
}

func (s *BookService) UpdatePreviewDefinition(req *api.UpdatePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "UpdatePreviewDefinition", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "UpdatePreviewDefinition", req.BookID, req.DefinitionID)
		}
	}()

	// rule checking
	definitionTxt := strings.TrimSpace(req.Definition)
	if err = CheckEmpty(definitionTxt); err != nil {
		return err
	}
	if err = CheckInvalidCharacters(definitionTxt); err != nil {
		return err
	}

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)
		if err := manager.DefinitionDAO.Update(context.TODO(), &model.Definition{
			ID:         req.DefinitionID,
			Definition: definitionTxt,
		}); err != nil {
			return err
		}
		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) UpdatePreviewPartOfSpeech(req *api.UpdatePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "UpdatePreviewPartOfSpeech", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "UpdatePreviewPartOfSpeech", req.BookID, req.DefinitionID)
		}
	}()

	// rule checking
	partOfSpeech := strings.TrimSpace(req.PartOfSpeech)
	if err = CheckEmpty(partOfSpeech); err != nil {
		return err
	}

	definition, err := manager.DefinitionDAO.GetFromID(context.TODO(), req.DefinitionID)
	if err != nil {
		return err
	}
	str, err := manager.StringDAO.GetFromID(context.TODO(), definition.StringID)
	if err != nil {
		return err
	}
	if err = CheckPartOfSpeech(str.Type, partOfSpeech); err != nil {
		return err
	}

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)

		if err := manager.DefinitionDAO.Update(context.TODO(), &model.Definition{
			ID:           req.DefinitionID,
			PartOfSpeech: partOfSpeech,
		}); err != nil {
			return err
		}

		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) UpdatePreviewSpecificType(req *api.UpdatePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "UpdatePreviewSpecificType", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "UpdatePreviewSpecificType", req.BookID, req.DefinitionID)
		}
	}()

	// rule checking
	specificType := strings.TrimSpace(req.SpecificType)

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)
		if specificType != "" {
			if err := manager.DefinitionDAO.Update(
				context.TODO(),
				&model.Definition{
					ID:           req.DefinitionID,
					SpecificType: specificType,
				},
				"specific_type",
			); err != nil {
				return err
			}
		} else {
			if err := manager.DefinitionDAO.DeleteFieldsByID(
				context.TODO(),
				req.DefinitionID,
				"specific_type",
			); err != nil {
				return err
			}
		}

		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) UpdatePreviewPronunciation(req *api.UpdatePreviewRequest, manager *dao.Manager, user string) error {
	var operateType string

	switch req.Field {
	case "pronunciation_ipa":
		operateType = "UpdatePreviewPronunciationIpa"
	case "pronunciation_ipa_weak":
		operateType = "UpdatePreviewPronunciationIpaWeak"
	case "pronunciation_ipa_other":
		operateType = "UpdatePreviewPronunciationIpaOther"
	case "pronunciation_text":
		operateType = "UpdatePreviewPronunciationText"
	}

	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, operateType, req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, operateType, req.BookID, req.DefinitionID)
		}
	}()

	// rule checking
	definition, err := manager.DefinitionDAO.GetFromID(context.TODO(), req.DefinitionID)
	if err != nil {
		return err
	}
	str, err := manager.StringDAO.GetFromID(context.TODO(), definition.StringID)
	if err != nil {
		return err
	}

	var deletedField string
	var selects []interface{}
	update := &model.Definition{ID: req.DefinitionID}
	switch operateType {
	case "UpdatePreviewPronunciationIpa":
		pronunciationIpa := strings.TrimSpace(req.PronunciationIpa)
		if err = CheckPronunciationIpa(pronunciationIpa, str, definition); err != nil {
			return err
		}
		if pronunciationIpa == "" {
			deletedField = "pronunciation_ipa"
			break
		}
		update.PronunciationIpa = pronunciationIpa
		selects = append(selects, "pronunciation_ipa")
	case "UpdatePreviewPronunciationIpaWeak":
		pronunciationIpaWeak := strings.TrimSpace(req.PronunciationIpaWeak)
		if err = CheckPronunciationIpaWeak(pronunciationIpaWeak, str, definition); err != nil {
			return err
		}
		if pronunciationIpaWeak == "" {
			deletedField = "pronunciation_ipa_weak"
			break
		}
		update.PronunciationIpaWeak = pronunciationIpaWeak
		selects = append(selects, "pronunciation_ipa_weak")
	case "UpdatePreviewPronunciationIpaOther":
		pronunciationIpaOther := strings.TrimSpace(req.PronunciationIpaOther)
		if err = CheckPronunciationIpaOther(pronunciationIpaOther, str, definition); err != nil {
			return err
		}
		if pronunciationIpaOther == "" {
			deletedField = "pronunciation_ipa_other"
			break
		}
		update.PronunciationIpaOther = pronunciationIpaOther
		selects = append(selects, "pronunciation_ipa_other")
	case "UpdatePreviewPronunciationText":
		pronunciationText := strings.TrimSpace(req.PronunciationText)
		if pronunciationText == "" {
			deletedField = "pronunciation_text"
			break
		}
		update.PronunciationText = pronunciationText
		selects = append(selects, "pronunciation_text")
	}

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)
		if deletedField != "" {
			if err := manager.DefinitionDAO.DeleteFieldsByID(
				context.TODO(),
				req.DefinitionID,
				deletedField,
			); err != nil {
				return err
			}
		} else {
			if err := manager.DefinitionDAO.Update(
				context.TODO(),
				update,
				selects...,
			); err != nil {
				return err
			}
		}
		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) UpdatePreviewExamples(req *api.UpdatePreviewRequest, manager *dao.Manager, user string) (*model.Example, error) {
	var operateType string

	switch req.Field {
	case "example_1":
		operateType = "UpdatePreviewExample1"
	case "example_2":
		operateType = "UpdatePreviewExample2"
	case "example_3":
		operateType = "UpdatePreviewExample3"
	}

	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, operateType, req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, operateType, req.BookID, req.DefinitionID)
		}
	}()

	examplesInDB, err := manager.ExampleDAO.GetFromDefinitionID(context.TODO(), req.DefinitionID)
	if err != nil {
		return nil, err
	}
	if len(examplesInDB) >= 3 && req.ExampleID == 0 {
		err = errors.New("example id required")
		return nil, err
	}
	valid := -1
	if req.ExampleID > 0 {
		for idx, exa := range examplesInDB {
			if exa.ID == req.ExampleID {
				valid = idx
			}
		}
		if valid < 0 {
			err = errors.New("invalid example id")
			return nil, err
		}
	}

	// rule checking
	definition, err := manager.DefinitionDAO.GetFromID(context.TODO(), req.DefinitionID)
	if err != nil {
		return nil, err
	}
	str, err := manager.StringDAO.GetFromID(context.TODO(), definition.StringID)
	if err != nil {
		return nil, err
	}

	var selects []interface{}
	update := &model.Example{}

	example := strings.TrimSpace(req.Example)
	if err = CheckEmpty(example); err != nil {
		return nil, err
	}
	if err = CheckInvalidCharacters(example); err != nil {
		return nil, err
	}

	stringData, err := s.dao.StringDAO.GetFromID(context.TODO(), definition.StringID)
	if err != nil {
		return nil, errors.New("can not find string")
	}

	wordForms := []string{stringData.String}
	relatedForms, err := s.dao.RelatedDAO.GetRelatedFormsByDefinitionID(context.TODO(), req.DefinitionID)
	if err != nil {
		return nil, err
	}
	for _, form := range relatedForms {
		wordForms = append(wordForms, form.String)
	}

	update.Content = example
	update.WordPositions = batch.FindWordPositionFromExample(example, wordForms)

	position := strings.TrimSpace(req.Position)
	contentLength := int64(len(example))
	if err = CheckPosition(position, contentLength); err != nil {
		return nil, err
	}
	if position != "" {
		update.WordPositions = position
	}

	selects = append(selects, "content", "word_positions")

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)

		if req.ExampleID > 0 {
			update.ID = req.ExampleID
			if err := manager.ExampleDAO.Update(context.TODO(), update, selects...); err != nil {
				return err
			}
		} else {
			update.StringID = str.ID
			update.DefinitionID = definition.ID
			if err := manager.ExampleDAO.Create(context.TODO(), update); err != nil {
				return err
			}
			if err := manager.RelatedDAO.CreateRelatationForExample(
				context.TODO(),
				update.ID,
				req.BookID,
				100*(len(examplesInDB)+1),
			); err != nil {
				return err
			}
		}

		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return update, nil
}

func (s *BookService) UpdatePreviewDefinitionComment(req *api.UpdatePreviewRequest, manager *dao.Manager, user string) (int64, error) {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "UpdatePreviewDefinitionComment", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "UpdatePreviewDefinitionComment", req.BookID, req.DefinitionID)
		}
	}()

	// rule checking
	definitionComment, err := manager.DefinitionCommentDAO.GetFromDefinitionID(context.TODO(), req.DefinitionID)
	if err != nil {
		return 0, err
	}
	if definitionComment.ID > 0 {
		if req.DefinitionCommentID != definitionComment.ID {
			err = errors.New("invalid definition comment id")
			return 0, err
		}
	}

	update := &model.DefinitionComment{
		ID:           req.DefinitionCommentID,
		DefinitionID: req.DefinitionID,
		Content:      req.DefinitionComment,
	}

	if update.ID > 0 {
		if strings.TrimSpace(req.DefinitionComment) == "" {
			if err = manager.DefinitionCommentDAO.DeleteFieldsByID(context.TODO(), req.DefinitionCommentID, "content"); err != nil {
				return 0, err
			}
		} else {
			if err = manager.DefinitionCommentDAO.Update(context.TODO(), update, "content"); err != nil {
				return 0, err
			}
		}
	} else {
		if err = manager.DefinitionCommentDAO.Create(context.TODO(), update); err != nil {
			return 0, err
		}
	}

	return update.ID, nil
}

func (s *BookService) UpdatePreviewForm(req *api.UpdatePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "UpdatePreviewForm", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "UpdatePreviewForm", req.BookID, req.DefinitionID)
		}
	}()

	// get definition
	definition, err := s.dao.DefinitionDAO.GetFromID(context.TODO(), req.DefinitionID)
	if err != nil {
		return err
	}

	// get base string
	baseString, err := s.dao.StringDAO.GetFromID(context.TODO(), definition.StringID)
	if err != nil {
		return err
	}

	// rule checking
	if err = CheckForm(definition.PartOfSpeech, req.Form); err != nil {
		return err
	}

	relatedForm := &dao.RelatedForm{
		String:           req.FormString,
		StringID:         req.FormStringID,
		Form:             req.Form,
		PartOfSpeech:     definition.PartOfSpeech,
		Definition:       req.Form + " of " + baseString.String,
		BaseStringID:     definition.StringID,
		BaseDefinitionID: definition.ID,
		Pronunciation:    req.Pronunciation,
	}

	if err = s.dao.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)
		if relatedForm.StringID > 0 {
			if err := manager.RelatedDAO.UpdateRelatedForm(
				context.TODO(),
				relatedForm,
			); err != nil {
				return err
			}
			relatedForm, err := s.dao.RelatedDAO.GetRelatedFormByDefinitionIDAndStringID(context.TODO(), req.DefinitionID, relatedForm.StringID)
			if err != nil {
				return err
			}
			formDefinition, err := s.dao.DefinitionDAO.GetFromID(context.TODO(), relatedForm.DefinitionID)
			if err != nil {
				return err
			}
			relatedForm.DefinitionID = formDefinition.ID
			// due to UpdateRelatedForm's callback, need to reassign field values
			relatedForm.Pronunciation = req.Pronunciation
			if err := manager.RelatedDAO.UpdateRelatedDefinition(
				context.TODO(),
				relatedForm,
			); err != nil {
				return err
			}
		} else {
			if err := manager.RelatedDAO.CreateRelatedForm(
				context.TODO(),
				relatedForm,
			); err != nil {
				return err
			}
		}

		// update book's updated_at field
		if err := manager.BookDAO.Update(
			context.TODO(),
			&model.Book{
				ID: req.BookID,
			},
		); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) UpdatePreviewSortValue(req *api.UpdatePreviewRequest, manager *dao.Manager, user string, relatedBookID int64) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "UpdatePreviewSortValue", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "UpdatePreviewSortValue", req.BookID, req.DefinitionID)
		}
	}()

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)

		if err := manager.RelatedDAO.UpdateBookLink(
			context.TODO(),
			&model.RelatedBook{
				ID:        relatedBookID,
				SortValue: req.SortValue,
			},
		); err != nil {
			return err
		}

		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *BookService) UpdatePreviewTranslation(req *api.UpdatePreviewRequest, manager *dao.Manager, user string) (int64, error) {
	var operateType string

	switch req.Field {
	case "definition_translation":
		operateType = "UpdatePreviewDefinitionTranslation"
	case "example_translation":
		operateType = "UpdatePreviewExampleTranslation"
	}

	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, operateType, req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, operateType, req.BookID, req.DefinitionID)
		}
	}()

	var translation *model.Translation

	switch req.Field {
	case "definition_translation":
		if req.TranslationID > 0 {
			translation, err = manager.TranslationDAO.GetFromID(context.TODO(), req.TranslationID)
			if err != nil {
				return 0, err
			}
			if translation.ItemType != "definition" || translation.ItemID != req.DefinitionID {
				return 0, errors.New("invalid translation id")
			}
			translation.Content = req.TranslationContent
			translation.LanguageCode = req.LanguageCode
		} else {
			translation = &model.Translation{
				ItemType:     "definition",
				ItemID:       req.DefinitionID,
				Content:      req.TranslationContent,
				LanguageCode: req.LanguageCode,
			}
		}
	case "example_translation":
		// checking
		example, err := manager.ExampleDAO.GetFromID(context.TODO(), req.ExampleID)
		if err != nil {
			return 0, err
		}
		if example.DefinitionID != req.DefinitionID {
			return 0, errors.New("invalid definition id or example id")
		}
		if req.TranslationID > 0 {
			translation, err = manager.TranslationDAO.GetFromID(context.TODO(), req.TranslationID)
			if err != nil {
				return 0, err
			}
			if translation.ItemType != "example" || translation.ItemID != req.ExampleID {
				return 0, errors.New("invalid translation id")
			}
			translation.Content = req.TranslationContent
			translation.LanguageCode = req.LanguageCode
		} else {
			translation = &model.Translation{
				ItemType:     "example",
				ItemID:       req.ExampleID,
				Content:      req.TranslationContent,
				LanguageCode: req.LanguageCode,
			}
		}
	}

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)

		if translation.ID > 0 {
			if err := manager.TranslationDAO.Update(context.TODO(), translation, "content", "language_code"); err != nil {
				return err
			}
		} else {
			if err := manager.TranslationDAO.Create(context.TODO(), translation); err != nil {
				return err
			}
		}

		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return translation.ID, nil
}

func (s *BookService) DeletePreviewItem(req *api.DeletePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "DeletePreviewItem", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "DeletePreviewItem", req.BookID, req.DefinitionID)
		}
	}()

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)

		// delete relations
		if err := manager.RelatedDAO.DeleteRelatedBook(
			context.TODO(),
			req.BookID,
			"definition",
			req.DefinitionID,
		); err != nil {
			return err
		}
		if err := manager.RelatedDAO.DeleteRelatedDefinition(
			context.TODO(),
			req.DefinitionID,
		); err != nil {
			return err
		}

		// delete definition
		if err := manager.DefinitionDAO.DeleteByID(
			context.TODO(),
			req.DefinitionID,
		); err != nil {
			return err
		}

		// delete examples
		if err := manager.ExampleDAO.DeleteByDefinitionID(context.TODO(), req.DefinitionID); err != nil {
			return err
		}

		// delete definition comment
		if err := manager.DefinitionCommentDAO.DeleteByDefinitionID(context.TODO(), req.DefinitionID); err != nil {
			return err
		}

		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) DeletePreviewSpecificType(req *api.DeletePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "DeletePreviewSpecificType", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "DeletePreviewSpecificType", req.BookID, req.DefinitionID)
		}
	}()

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)
		if err := manager.DefinitionDAO.DeleteFieldsByID(
			context.TODO(),
			req.DefinitionID,
			"specific_type",
		); err != nil {
			return err
		}
		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) DeletePreviewPronunciationIpaWeak(req *api.DeletePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "DeletePreviewPronunciationIpaWeak", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "DeletePreviewPronunciationIpaWeak", req.BookID, req.DefinitionID)
		}
	}()

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)
		if err := manager.DefinitionDAO.DeleteFieldsByID(
			context.TODO(),
			req.DefinitionID,
			"pronunciation_ipa_weak",
		); err != nil {
			return err
		}
		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) DeletePreviewPronunciationIpaOther(req *api.DeletePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "DeletePreviewPronunciationIpaOther", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "DeletePreviewPronunciationIpaOther", req.BookID, req.DefinitionID)
		}
	}()

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)
		if err := manager.DefinitionDAO.DeleteFieldsByID(
			context.TODO(),
			req.DefinitionID,
			"pronunciation_ipa_other",
		); err != nil {
			return err
		}
		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) DeletePreviewPronunciationText(req *api.DeletePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "DeletePreviewPronunciationText", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "DeletePreviewPronunciationText", req.BookID, req.DefinitionID)
		}
	}()

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)
		if err := manager.DefinitionDAO.DeleteFieldsByID(
			context.TODO(),
			req.DefinitionID,
			"pronunciation_text",
		); err != nil {
			return err
		}
		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) DeletePreviewExample(req *api.DeletePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "DeletePreviewExample", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "DeletePreviewExample", req.BookID, req.DefinitionID)
		}
	}()

	// checking
	example, err := manager.ExampleDAO.GetFromID(context.TODO(), req.ExampleID)
	if err != nil {
		return err
	}
	if example.DefinitionID != req.DefinitionID {
		return errors.New("invalid definition id or example id")
	}

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)
		if err := manager.ExampleDAO.DeleteByID(context.TODO(), req.ExampleID); err != nil {
			return err
		}
		// delete related_book
		if err := manager.RelatedDAO.DeleteRelatedBook(context.TODO(), req.BookID, "example", req.ExampleID); err != nil {
			return err
		}

		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) DeletePreviewDefinitionComment(req *api.DeletePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "DeletePreviewDefinitionComment", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "DeletePreviewDefinitionComment", req.BookID, req.DefinitionID)
		}
	}()

	// checking
	comment, err := manager.DefinitionCommentDAO.GetFromID(context.TODO(), req.DefinitionCommentID)
	if err != nil {
		return err
	}
	if comment.DefinitionID != req.DefinitionID {
		return errors.New("invalid definition id or comment id")
	}

	if err = manager.DefinitionCommentDAO.DeleteByID(context.TODO(), req.DefinitionCommentID); err != nil {
		return err
	}

	return nil
}

func (s *BookService) DeletePreviewForm(req *api.DeletePreviewRequest, manager *dao.Manager, user string) error {
	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, "DeletePreviewForm", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "DeletePreviewForm", req.BookID, req.DefinitionID)
		}
	}()

	// get relatedForms
	relatedForms, err := s.dao.RelatedDAO.GetRelatedFormsByDefinitionID(context.TODO(), req.DefinitionID)
	if err != nil {
		return err
	}

	var relatedForm *dao.RelatedForm
	for idx, form := range relatedForms {
		if form.StringID == req.FormStringID {
			relatedForm = &relatedForms[idx]
			break
		}
	}

	if relatedForm == nil {
		return errors.New("invalid form string id")
	}

	if err = s.dao.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)

		if err := manager.RelatedDAO.DeleteRelatedForm(
			context.TODO(),
			relatedForm,
		); err != nil {
			return err
		}

		// update book's updated_at field
		if err := manager.BookDAO.Update(
			context.TODO(),
			&model.Book{
				ID: req.BookID,
			},
		); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BookService) DeletePreviewTranslation(req *api.DeletePreviewRequest, manager *dao.Manager, user string) error {
	var operateType string

	switch req.Field {
	case "definition_translation":
		operateType = "DeletePreviewDefinitionTranslation"
	case "example_translation":
		operateType = "DeletePreviewExampleTranslation"
	}

	var err error

	defer func() {
		if err != nil {
			s.Complete(err, user, operateType, req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, operateType, req.BookID, req.DefinitionID)
		}
	}()

	// checking
	translation, err := manager.TranslationDAO.GetFromID(context.TODO(), req.TranslationId)
	if err != nil {
		return err
	}
	switch req.Field {
	case "definition_translation":
		if translation.ItemType != "definition" || translation.ItemID != req.DefinitionID {
			return errors.New("invalid translation id or definition id")
		}
	case "example_translation":
		// checking
		example, err := manager.ExampleDAO.GetFromID(context.TODO(), req.ExampleID)
		if err != nil {
			return err
		}
		if example.DefinitionID != req.DefinitionID {
			return errors.New("invalid definition id or example id")
		}
		if translation.ItemType != "example" || translation.ItemID != req.ExampleID {
			return errors.New("invalid translation id or example id")
		}
	}

	if err = manager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)
		if err := manager.TranslationDAO.DeleteByID(context.TODO(), req.TranslationId); err != nil {
			return err
		}

		// update book's updated_at field
		if err := manager.BookDAO.Update(context.TODO(), &model.Book{ID: req.BookID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
