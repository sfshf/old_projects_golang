package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/nextsurfer/book-manage-api/api"
	"github.com/nextsurfer/book-manage-api/api/code"
	"github.com/nextsurfer/book-manage-api/internal/app/batch"
	"github.com/nextsurfer/book-manage-api/internal/app/dao"
	"github.com/nextsurfer/book-manage-api/internal/app/dto"
	"github.com/nextsurfer/book-manage-api/internal/app/model"
	"github.com/nextsurfer/book-manage-api/internal/tools"
	"gorm.io/gorm"
)

// BookService : service is pure business
type BookService struct {
	dao *dao.Manager
	// dbManager  *dao.Manager
	httpClient *http.Client

	updating    bool
	pendingLogs []string
	wordCount   int
	// percentage of progress:
	// 0-5% is for read csv
	// 5-100% is for word count
	progress    int
	uploadError error
}

// NewBookService is factory function
func NewBookService() *BookService {
	return &BookService{
		dao: dao.NewManagerWithDB(tools.MysqlDB()),
		// dbManager:  dao.NewManagerWithDB(tools.MysqlDB()),
		httpClient: &http.Client{},
		updating:   false,
	}
}

func (s *BookService) GetCSV(c *gin.Context) (*api.DownloadResponseData, int32, string) {
	password := c.Query("password")
	// password check
	if !tools.CheckAdminPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}

	bookIDString := c.Query("book")

	fmt.Println("export csv, password: ", password, " bookID: ", bookIDString)

	if len(bookIDString) == 0 {
		return nil, code.ErrWrongParam, "wrong book ID"
	}

	bookID, err := strconv.ParseInt(bookIDString, 10, 64)
	if err != nil {
		return nil, code.ErrWrongParam, "bookID wrong"
	}
	// bookId check
	bookData, err := s.dao.BookDAO.GetFromID(context.Background(), bookID)
	if err != nil {
		return nil, code.ErrWrongParam, "bookID not exist"
	}
	// fmt.Println("book time", bookData.UpdatedAt.UnixMilli())

	// check if book is exported
	// check dir exist
	dirPath := tools.Config().DownloadPath + "/csv/" + bookIDString
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Println("dirPath not exist")
		// if not exist, create folder
		err := os.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			fmt.Println("Mkdir failed:", err)
			return nil, code.ErrInternal, err.Error()
		}
	}
	// check file exist
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println("ReadDir failed:", err)
		return nil, code.ErrInternal, err.Error()
	}
	targetFileName := strconv.FormatInt(bookData.UpdatedAt.UnixMilli(), 10) + ".csv"
	data := &api.DownloadResponseData{}
	data.Path = "/download/csv/" + bookIDString + "/" + targetFileName
	for _, file := range files {
		// fmt.Println(file.Name(), file.IsDir())
		if file.Name() == targetFileName {
			// fmt.Println("TODO find file")
			return data, code.Ok, ""
		} else if strings.Contains(file.Name(), ".csv") {
			// delete old file
			// fmt.Println("TODO delete this file", file.Name())
			err := os.Remove(dirPath + "/" + file.Name())
			if err != nil {
				fmt.Println("Remove failed:", err)
				return nil, code.ErrInternal, err.Error()
			}
		}
	}

	// export book
	start := time.Now().UnixMilli()
	err = batch.ExportBook(bookID, dirPath+"/"+targetFileName, s.dao)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	end := time.Now().UnixMilli()
	fmt.Println("export book time : ", end-start, " ms")
	return data, code.Ok, ""
}

func (s *BookService) GetBundle(c *gin.Context) (*api.DownloadResponseData, int32, string) {
	password := c.Query("password")
	// password check
	if !tools.CheckAdminPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}

	bookIDString := c.Query("book")

	fmt.Println("export bundle, password: ", password, " bookID: ", bookIDString)

	if len(bookIDString) == 0 {
		return nil, code.ErrWrongParam, "wrong book ID"
	}

	bookID, err := strconv.ParseInt(bookIDString, 10, 64)
	if err != nil {
		return nil, code.ErrWrongParam, "bookID wrong"
	}
	// bookId check
	bookData, err := s.dao.BookDAO.GetFromID(context.Background(), bookID)
	if err != nil {
		return nil, code.ErrWrongParam, "bookID not exist"
	}
	// fmt.Println("book time", bookData.UpdatedAt.UnixMilli())

	// check if book is exported
	// check dir exist
	dirPath := tools.Config().DownloadPath + "/bundle/" + bookIDString
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Println("dirPath not exist")
		// if not exist, create folder
		err := os.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			fmt.Println("Mkdir failed:", err)
			return nil, code.ErrInternal, err.Error()
		}
	}
	// check file exist
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println("ReadDir failed:", err)
		return nil, code.ErrInternal, err.Error()
	}
	targetFileName := strconv.FormatInt(bookData.UpdatedAt.UnixMilli(), 10) + ".bundle"
	data := &api.DownloadResponseData{}
	data.Path = "/download/bundle/" + bookIDString + "/" + targetFileName
	for _, file := range files {
		// fmt.Println(file.Name(), file.IsDir())
		if file.Name() == targetFileName {
			// fmt.Println("TODO find file")
			return data, code.Ok, ""
		} else if strings.Contains(file.Name(), ".bundle") {
			// delete old file
			// fmt.Println("TODO delete this file", file.Name())
			err := os.Remove(dirPath + "/" + file.Name())
			if err != nil {
				fmt.Println("Remove failed:", err)
				return nil, code.ErrInternal, err.Error()
			}
		}
	}

	// export book
	start := time.Now().UnixMilli()
	// err = batch.ExportBook(bookID, dirPath+"/"+targetFileName, tools.Config().Mysql)
	err = batch.ExportEncryptedBook(bookID, dirPath+"/"+targetFileName, s.dao)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	end := time.Now().UnixMilli()
	fmt.Println("export book time : ", end-start, " ms")
	return data, code.Ok, ""
}

func (s *BookService) GetAllBooks(c *gin.Context) (*api.BookListResponseData, int32, string) {
	password := c.Query("password")

	// password check
	if !tools.CheckPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}

	data := &api.BookListResponseData{}
	books, err := s.dao.BookDAO.GetAll(context.Background())
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	} else {
		bookDTOs := make([]dto.BookDTO, len(books))
		for i := 0; i < len(books); i++ {
			bookDTOs[i] = dto.BookDTO{
				ID:          books[i].ID,
				UpdatedAt:   books[i].UpdatedAt.UnixMilli(),
				Name:        books[i].Name,
				Description: books[i].Description,
				DownloadURL: books[i].DownloadURL,
			}
		}
		data.Books = bookDTOs
		return data, code.Ok, ""
	}
}

func (s *BookService) AddBook(c *gin.Context) (int32, string) {
	formData, err := c.MultipartForm()
	if err != nil {
		return code.ErrInternal, err.Error()
	}
	passwordList := formData.Value["password"]
	if len(passwordList) == 0 || len(passwordList[0]) == 0 {
		return code.ErrWrongParam, "password required"
	}
	// password check
	password := passwordList[0]
	if !tools.CheckAdminPassword(password) {
		return code.ErrPassword, "wrong password"
	}
	user := tools.GetUserNameByPassword(password)

	names := formData.Value["name"]
	descriptions := formData.Value["description"]
	files := formData.File["file"]

	// param check
	if len(names) == 0 || len(names[0]) == 0 {
		return code.ErrWrongParam, "name required"
	}
	if len(descriptions) == 0 || len(descriptions[0]) == 0 {
		return code.ErrWrongParam, "description required"
	}
	if len(files) == 0 {
		return code.ErrWrongParam, "csv file required"
	}

	if s.updating {
		return code.ErrWrongParam, "A file is uploading, please wait"
	}

	name := strings.TrimSpace(names[0])
	description := strings.TrimSpace(descriptions[0])

	existed, err := s.dao.BookDAO.GetFromName(context.Background(), name)
	if err != nil {
		fmt.Println("GetFromName error: ", err)
		return code.ErrInternal, err.Error()
	}
	if existed != nil {
		return code.ErrWrongParam, "book name existed"
	}

	fmt.Println("AddBook, name: ", name, " description: ", description)

	s.start()
	go batch.AddBook(files[0], name, description, s, s.dao, user)
	return code.Ok, ""
}

func (s *BookService) UpdateBook(c *gin.Context) (int32, string) {
	formData, err := c.MultipartForm()
	if err != nil {
		return code.ErrInternal, err.Error()
	}
	passwordList := formData.Value["password"]
	if len(passwordList) == 0 || len(passwordList[0]) == 0 {
		return code.ErrWrongParam, "password required"
	}
	// password check
	password := passwordList[0]
	if !tools.CheckAdminPassword(password) {
		return code.ErrPassword, "wrong password"
	}
	user := tools.GetUserNameByPassword(password)

	names := formData.Value["name"]
	descriptions := formData.Value["description"]
	updateInfos := formData.Value["updateInfo"]
	strictModeData := formData.Value["strictMode"]
	bookIds := formData.Value["bookID"]
	files := formData.File["file"]

	// param check
	if len(bookIds) == 0 || len(bookIds[0]) == 0 {
		return code.ErrWrongParam, "bookId required"
	}
	if len(updateInfos) == 0 || len(updateInfos[0]) == 0 {
		return code.ErrWrongParam, "updateInfo required"
	}
	bookID, err := strconv.ParseInt(bookIds[0], 10, 64)
	if err != nil {
		return code.ErrWrongParam, err.Error()
	}

	// bookId check
	bookData, err := s.dao.BookDAO.GetFromID(context.TODO(), bookID)
	if err != nil {
		return code.ErrWrongParam, "bookID not exist"
	}
	updateInfo := updateInfos[0] == "true"
	if updateInfo {
		fmt.Println("update book info, bookID : ", bookID)
		if len(names) == 0 || len(names[0]) == 0 {
			return code.ErrWrongParam, "name required"
		}
		if len(descriptions) == 0 || len(descriptions[0]) == 0 {
			return code.ErrWrongParam, "description required"
		}
		name := strings.TrimSpace(names[0])
		description := strings.TrimSpace(descriptions[0])

		existed, err := s.dao.BookDAO.GetFromName(context.Background(), name)
		if err != nil {
			fmt.Println("GetFromName error: ", err)
			return code.ErrInternal, err.Error()
		}
		if existed != nil && existed.ID != bookID {
			return code.ErrWrongParam, "book name existed"
		}

		bookData.Name = name
		bookData.Description = description
		err = s.dao.BookDAO.Update(context.Background(), bookData)
		if err != nil {
			return code.ErrInternal, err.Error()
		}
	} else {
		fmt.Println("update book data, bookID : ", bookID)

		if len(strictModeData) == 0 || len(strictModeData[0]) == 0 {
			return code.ErrWrongParam, "strictMode required"
		}
		strictMode := strictModeData[0] == "true"
		if len(files) == 0 {
			return code.ErrWrongParam, "csv file required"
		}
		if s.updating {
			return code.ErrWrongParam, "A file is uploading, please wait"
		}
		s.start()
		go batch.UpdateBook(files[0], bookID, s, s.dao, user, strictMode)
	}
	return code.Ok, ""
}

func (s *BookService) SearchBookItem(c *gin.Context) (*api.SearchBookItemResponseData, int32, string) {
	password := c.Query("password")
	// password check
	if !tools.CheckPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}
	bookIDStr := c.Query("bookID")
	indexStr := c.Query("index")
	withCommentStr := c.Query("withComment")

	// fmt.Println("SearchBookItem, bookID: ", bookIDStr, " index: ", indexStr)

	bookID, err := strconv.ParseInt(bookIDStr, 10, 64)
	if err != nil {
		return nil, code.ErrWrongParam, "bookID wrong"
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return nil, code.ErrWrongParam, "index wrong"
	}
	withComment, err := strconv.ParseBool(withCommentStr)
	if err != nil {
		return nil, code.ErrWrongParam, "withComment wrong"
	}

	data := &api.SearchBookItemResponseData{}

	// bookId check
	_, err = s.dao.BookDAO.GetFromID(context.TODO(), bookID)
	if err != nil {
		return data, code.ErrWrongParam, "invalid bookId"
	}

	// get related_book
	record, count, err := s.dao.RelatedDAO.GetDefinitionByBookIDAndSortNumberWithComment(c, bookID, index, withComment)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	if record.ID == 0 {
		return data, code.ErrWrongParam, "invalid number"
	}

	// get definition
	definition, err := s.dao.DefinitionDAO.GetFromID(c, record.ItemID)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}

	// get examples
	examples, err := s.dao.ExampleDAO.GetFromDefinitionID(c, definition.ID)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}

	// get string
	str, err := s.dao.StringDAO.GetFromID(c, definition.StringID)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}

	// get related forms
	relatedForms, err := s.dao.RelatedDAO.GetRelatedFormsByDefinitionID(c, definition.ID)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}

	// get defnition comment
	comment, err := s.dao.DefinitionCommentDAO.GetFromDefinitionID(c, definition.ID)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}

	// get definition translation
	definitionTranslations, err := s.dao.TranslationDAO.GetTranslationsByDefinitionID(context.TODO(), definition.ID)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}

	// get example translations
	var exampleIDs []int64
	for _, example := range examples {
		exampleIDs = append(exampleIDs, example.ID)
	}
	exampleTranslations, err := s.dao.TranslationDAO.GetTranslationsByExampleIDs(context.TODO(), exampleIDs)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}

	// data.items
	data.Item = api.SearchBookItemResponseItem{
		String:                 str.String,
		Type:                   str.Type,
		SortValue:              record.SortValue,
		Definition:             *definition,
		DefinitionTranslations: definitionTranslations,
		Examples:               examples,
		RelatedForms:           relatedForms,
		DefinitionComment:      comment,
		ExampleTranslations:    exampleTranslations,
	}
	data.Total = count
	data.Index = index
	if data.Index >= count {
		data.Index = count - 1
	}

	return data, code.Ok, ""
}

func (s *BookService) TextToSpeach(c *gin.Context) (*api.DownloadResponseData, int32, string) {
	password := c.Query("password")
	// password check
	if !tools.CheckPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}
	ipa := strings.TrimSpace(c.Query("ipa"))
	text := strings.TrimSpace(c.Query("text"))

	if ipa == "" && text == "" {
		return nil, code.ErrWrongParam, "ipa or text required"
	}
	ssml := ""
	if ipa != "" {
		ssml = fmt.Sprintf("<speak><phoneme alphabet=\"ipa\" ph=\"%s\">a</phoneme></speak>", ipa)
	} else {
		ssml = fmt.Sprintf("<speak>%s</speak>", text)
	}

	postBody, _ := json.Marshal(map[string]string{
		"ssml":   ssml,
		"accent": "us",
		"voice":  "Joanna",
		"apiKey": tools.Config().TTSAPIKey,
	})
	responseBody := bytes.NewBuffer(postBody)
	// s.httpClient.Get("https://api.n1xt.net/word/audio/getAudioURL/v1" + ssml)
	resp, err := s.httpClient.Post("https://api.n1xt.net/word/audio/getAudioURL/v1", "application/json", responseBody)
	if err != nil {
		fmt.Println("http post error: ", err)
		return nil, code.ErrInternal, err.Error()
	}
	defer resp.Body.Close()

	// fmt.Println("response Status:", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("io.ReadAll error: ", err)
		return nil, code.ErrInternal, err.Error()
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil, code.ErrInternal, err.Error()
	}

	_code := int(data["code"].(float64))
	if _code != 0 {
		message := data["message"].(string)
		fmt.Println("code error: ", _code, "  msg: ", message)
		return nil, code.ErrInternal, message
	}

	emData := data["data"].(map[string]interface{})
	audioURL := emData["audioURL"].(string)
	// fmt.Println("success, audioURL :", audioURL)

	return &api.DownloadResponseData{
		Path: audioURL,
	}, code.Ok, ""
}

func (s *BookService) UploadingLog(c *gin.Context) (*api.UploadingLogResponseData, int32, string) {
	password := c.Query("password")

	// password check
	if !tools.CheckPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}

	// if !s.updating {
	// 	return nil, code.ErrWrongParam, "not updating"
	// }

	data := &api.UploadingLogResponseData{
		Logs:      s.pendingLogs,
		WordCount: s.wordCount,
		Progress:  s.progress,
		Error:     "",
	}
	if s.uploadError != nil {
		data.Error = s.uploadError.Error()
	}
	s.pendingLogs = []string{}

	return data, code.Ok, ""
}

func (s *BookService) UpdatePreview(c *gin.Context) (interface{}, int32, string) {
	var req api.UpdatePreviewRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	if !tools.CheckPassword(req.Password) {
		return nil, code.ErrPassword, "wrong password"
	}
	user := tools.GetUserNameByPassword(req.Password)

	// check relation of bookID and definitionID
	relatedBook, err := s.dao.RelatedDAO.GetRelatedBookForDefinition(c, req.DefinitionID, req.BookID)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	if relatedBook == nil {
		return nil, code.ErrWrongParam, "invalid definition or book"
	}

	var res interface{}

	switch req.Field {
	case "string":
		res, err = s.UpdatePreviewString(&req, s.dao, user)
	case "type":
	case "definition":
		err = s.UpdatePreviewDefinition(&req, s.dao, user)
	case "part_of_speech":
		err = s.UpdatePreviewPartOfSpeech(&req, s.dao, user)
	case "specific_type":
		err = s.UpdatePreviewSpecificType(&req, s.dao, user)
	case "pronunciation_ipa", "pronunciation_ipa_weak",
		"pronunciation_ipa_other", "pronunciation_text":
		err = s.UpdatePreviewPronunciation(&req, s.dao, user)
	case "example_1", "example_2", "example_3":
		res, err = s.UpdatePreviewExamples(&req, s.dao, user)
	case "definition_comment":
		res, err = s.UpdatePreviewDefinitionComment(&req, s.dao, user)
	case "form":
		err = s.UpdatePreviewForm(&req, s.dao, user)
	case "sort_value":
		err = s.UpdatePreviewSortValue(&req, s.dao, user, relatedBook.ID)
	case "definition_translation", "example_translation":
		res, err = s.UpdatePreviewTranslation(&req, s.dao, user)
	}

	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}

	return res, code.Ok, ""
}

func (s *BookService) GetExamplePosition(c *gin.Context) (string, int32, string) {
	password := c.Query("password")
	// password check
	if !tools.CheckPassword(password) {
		return "", code.ErrPassword, "wrong password"
	}

	example := c.Query("example")

	definitionIDString := strings.TrimSpace(c.Query("definitionID"))
	definitionID, err := strconv.ParseInt(definitionIDString, 10, 64)
	if err != nil {
		return "", code.ErrWrongParam, "definitionID wrong"
	}

	definition, err := s.dao.DefinitionDAO.GetFromID(context.TODO(), definitionID)
	if err != nil {
		return "", code.ErrWrongParam, "definitionID wrong"
	}
	stringData, err := s.dao.StringDAO.GetFromID(context.TODO(), definition.StringID)
	if err != nil {
		return "", code.ErrWrongParam, "can not find string"
	}
	// related forms don't include the base string.
	var wordForms []string = []string{stringData.String}
	relatedForms, err := s.dao.RelatedDAO.GetRelatedFormsByDefinitionID(context.TODO(), definitionID)
	if err != nil {
		return "", code.ErrInternal, err.Error()
	}
	for _, form := range relatedForms {
		wordForms = append(wordForms, form.String)
	}

	return batch.FindWordPositionFromExample(example, wordForms), code.Ok, ""
}

func (s *BookService) DeletePreview(c *gin.Context) (int32, string) {
	var req api.DeletePreviewRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		return code.ErrInternal, err.Error()
	}
	if !tools.CheckPassword(req.Password) {
		return code.ErrPassword, "wrong password"
	}
	user := tools.GetUserNameByPassword(req.Password)

	// check relation of bookID and definitionID
	relatedBook, err := s.dao.RelatedDAO.GetRelatedBookForDefinition(c, req.DefinitionID, req.BookID)
	if err != nil {
		return code.ErrInternal, err.Error()
	}
	if relatedBook == nil {
		return code.ErrWrongParam, "invalid definition or book"
	}

	switch req.Field {
	case "preview_item":
		if !tools.CheckAdminPassword(req.Password) {
			return code.ErrPassword, "只有管理员可以执⾏该操作"
		}
		err = s.DeletePreviewItem(&req, s.dao, user)
	case "specific_type":
		err = s.DeletePreviewSpecificType(&req, s.dao, user)
	case "pronunciation_ipa_weak":
		err = s.DeletePreviewPronunciationIpaWeak(&req, s.dao, user)
	case "pronunciation_ipa_other":
		err = s.DeletePreviewPronunciationIpaOther(&req, s.dao, user)
	case "pronunciation_text":
		err = s.DeletePreviewPronunciationText(&req, s.dao, user)
	case "example_1", "example_2", "example_3":
		err = s.DeletePreviewExample(&req, s.dao, user)
	case "definition_comment":
		err = s.DeletePreviewDefinitionComment(&req, s.dao, user)
	case "form":
		err = s.DeletePreviewForm(&req, s.dao, user)
	case "definition_translation", "example_translation":
		err = s.DeletePreviewTranslation(&req, s.dao, user)
	}

	if err != nil {
		return code.ErrInternal, err.Error()
	}

	return code.Ok, ""
}

func (s *BookService) SearchStringPagination(c *gin.Context) (*api.SearchStringPaginationResponseData, int32, string) {
	ctx := context.TODO()
	password := c.Query("password")
	// password check
	if !tools.CheckPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}

	var err error
	var bookIDs []int64
	for _, item := range tools.DefaultBookLevelConfig.Items {
		bookIDs = append(bookIDs, item.BookID)
	}

	var params dao.DefinitionWithStringParams
	if err := c.ShouldBindQuery(&params); err != nil {
		return nil, code.ErrWrongParam, err.Error()
	}
	params.OrderByString = true
	params.BookIDs = bookIDs
	definitionWithStrings, total, err := s.dao.DefinitionDAO.SearchDefinitionWithString(ctx, &params)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	var respItems []*api.SearchStringPaginationResponseItem
	for _, definitionWithString := range definitionWithStrings {
		respItems = append(respItems, &api.SearchStringPaginationResponseItem{
			BookID:       definitionWithString.BookID,
			DefinitionID: definitionWithString.ID,
			String:       definitionWithString.String,
			Level:        definitionWithString.CefrLevel,
			Type:         definitionWithString.Type,
			PartOfSpeech: definitionWithString.PartOfSpeech,
			Index:        definitionWithString.Idx - 1,
			Definition:   definitionWithString.Definition,
		})
	}
	return &api.SearchStringPaginationResponseData{
		Items: respItems,
		Total: total,
	}, 0, ""
}

func (s *BookService) GetCefrLevels(c *gin.Context) (*api.GetCefrLevelsResponseData, int32, string) {
	var req api.GetCefrLevelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	if !tools.CheckPassword(req.Password) {
		return nil, code.ErrPassword, "wrong password"
	}
	// check cur_level
	var list []api.GetCefrLevelsResponseItem
	// check book_id
	if req.BookID > 0 {
		for _, item := range tools.DefaultBookLevelConfig.Items {
			if req.BookID == item.BookID {
				list = append(list, api.GetCefrLevelsResponseItem{
					BookID: item.BookID,
					Level:  item.Level,
				})
				break
			}
		}
		return &api.GetCefrLevelsResponseData{
			List: list,
		}, 0, ""
	}
	for _, item := range tools.DefaultBookLevelConfig.Items {
		if req.CurLevel != "" && req.CurLevel == item.Level {
			continue
		}
		list = append(list, api.GetCefrLevelsResponseItem{
			BookID: item.BookID,
			Level:  item.Level,
		})
	}
	return &api.GetCefrLevelsResponseData{
		List: list,
	}, 0, ""
}

func (s *BookService) GetNextSortValue(c *gin.Context) (*api.GetNextSortValueResponseData, int32, string) {
	password := c.Query("password")
	// password check
	if !tools.CheckPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}
	ctx := c.Request.Context()
	level := c.Query("cur_level")
	var bookID int64
	for _, item := range tools.DefaultBookLevelConfig.Items {
		if level == item.Level {
			bookID = item.BookID
			break
		}
	}
	// get max sort value of the book
	maxSortValue, err := s.dao.RelatedDAO.GetMaxSortValue(ctx, bookID)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	return &api.GetNextSortValueResponseData{
		NextSortValue: maxSortValue + 100,
	}, 0, ""
}

func (s *BookService) UpdateCefrLevel(c *gin.Context) (int32, string) {
	var req api.UpdateCefrLevelRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		return code.ErrInternal, err.Error()
	}
	if !tools.CheckPassword(req.Password) {
		return code.ErrPassword, "wrong password"
	}
	user := tools.GetUserNameByPassword(req.Password)
	var err error
	defer func() {
		if err != nil {
			s.Complete(err, user, "UpdateCefrLevel", req.BookID, req.DefinitionID)
		} else {
			s.Complete(nil, user, "UpdateCefrLevel", req.BookID, req.DefinitionID)
		}
	}()
	ctx := c.Request.Context()
	var bookID int64
	for _, item := range tools.DefaultBookLevelConfig.Items {
		if req.CefrLevel == item.Level {
			bookID = item.BookID
			break
		}
	}
	newModel := &model.RelatedBook{
		BookID:    bookID,
		SortValue: int32(req.SortValue),
	}
	if err := s.dao.RelatedDAO.UpdateBookIDAndSortValue(ctx, req.DefinitionID, req.BookID, newModel); err != nil {
		return code.ErrInternal, err.Error()
	}
	return 0, ""
}

func (s *BookService) GetDefinitionInfo(c *gin.Context) (*api.GetDefinitionInfoResponse, int32, string) {
	var req api.GetDefinitionInfoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	if !tools.CheckPassword(req.Password) {
		return nil, code.ErrPassword, "wrong password"
	}
	var err error
	ctx := c.Request.Context()
	relatedBooks, err := s.dao.RelatedDAO.GetBooksByDefinitionID(ctx, req.DefinitionID)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	if len(relatedBooks) == 0 {
		return nil, 0, ""
	}
	bookID := relatedBooks[0].BookID
	index, err := s.dao.RelatedDAO.GetDefinitionIndexByBookIDAndDefinitionID(ctx, bookID, req.DefinitionID)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	return &api.GetDefinitionInfoResponse{
		BookID: bookID,
		Index:  index,
	}, 0, ""
}

func (s *BookService) NewDefinition(c *gin.Context) (int32, string) {
	var req api.NewDefinitionRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		return code.ErrInternal, err.Error()
	}
	if !tools.CheckPassword(req.Password) {
		return code.ErrPassword, "wrong password"
	}
	user := tools.GetUserNameByPassword(req.Password)
	var err error
	defer func() {
		if err != nil {
			s.Complete(err, user, "NewDefinition", req.BookID)
		} else {
			s.Complete(nil, user, "NewDefinition", req.BookID)
		}
	}()
	ctx := c.Request.Context()
	if err = s.dao.DB.Transaction(func(tx *gorm.DB) error {
		txDao := dao.NewManagerWithDB(tx)
		// bookID
		_, err := txDao.BookDAO.GetFromID(ctx, req.BookID)
		if err != nil {
			return err
		}
		// string, type
		targetString, err := txDao.StringDAO.GetFromString(ctx, req.String)
		if err != nil {
			return err
		}
		if targetString == nil {
			targetString = &model.String{
				String: req.String,
				Type:   req.Type,
			}
			if err := txDao.StringDAO.Create(ctx, targetString); err != nil {
				return err
			}
		}
		// definition, cefrLevel, pronunciation*, partOfSpeech, specificType
		newDefinition := &model.Definition{
			StringID:              targetString.ID,
			PartOfSpeech:          req.PartOfSpeech,
			SpecificType:          req.SpecificType,
			PronunciationIpa:      req.PronunciationIpa,
			PronunciationIpaWeak:  req.PronunciationIpaWeak,
			PronunciationIpaOther: req.PronunciationIpaOther,
			PronunciationText:     req.PronunciationText,
			CefrLevel:             req.CefrLevel,
			Definition:            req.Definition,
		}
		if err := txDao.DefinitionDAO.Create(ctx, newDefinition); err != nil {
			return err
		}
		// forms
		if len(req.Forms) > 0 {
			for _, item := range req.Forms {
				if err := txDao.RelatedDAO.CreateRelatedForm(ctx, &dao.RelatedForm{
					String:           item.FormString,
					Form:             item.Form,
					PartOfSpeech:     newDefinition.PartOfSpeech,
					Definition:       item.Form + " of " + targetString.String,
					BaseStringID:     newDefinition.StringID,
					BaseDefinitionID: newDefinition.ID,
					Pronunciation:    item.Pronunciation,
				}); err != nil {
					return err
				}
			}
		}
		// sortValue
		if err := txDao.RelatedDAO.CreateRelationForDefinition(ctx, newDefinition.ID, req.BookID, int(req.SortValue)); err != nil {
			return err
		}
		// example*, positions*
		if req.Example1 != "" {
			newExample := &model.Example{
				StringID:      targetString.ID,
				DefinitionID:  newDefinition.ID,
				Content:       req.Example1,
				WordPositions: req.Positions1,
			}
			if err := txDao.ExampleDAO.Create(ctx, newExample); err != nil {
				return err
			}
			if err := txDao.RelatedDAO.CreateRelatationForExample(ctx, newExample.ID, req.BookID, 100); err != nil {
				return err
			}
		}
		if req.Example2 != "" {
			newExample := &model.Example{
				StringID:      targetString.ID,
				DefinitionID:  newDefinition.ID,
				Content:       req.Example2,
				WordPositions: req.Positions2,
			}
			if err := txDao.ExampleDAO.Create(ctx, newExample); err != nil {
				return err
			}
			if err := txDao.RelatedDAO.CreateRelatationForExample(ctx, newExample.ID, req.BookID, 200); err != nil {
				return err
			}
		}
		if req.Example3 != "" {
			newExample := &model.Example{
				StringID:      targetString.ID,
				DefinitionID:  newDefinition.ID,
				Content:       req.Example3,
				WordPositions: req.Positions3,
			}
			if err := txDao.ExampleDAO.Create(ctx, newExample); err != nil {
				return err
			}
			if err := txDao.RelatedDAO.CreateRelatationForExample(ctx, newExample.ID, req.BookID, 300); err != nil {
				return err
			}
		}
		// update book's updated_at field
		if err := txDao.BookDAO.Update(
			ctx,
			&model.Book{
				ID: req.BookID,
			},
		); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return code.ErrInternal, err.Error()
	}
	return 0, ""
}

func (s *BookService) ListDefinition(c *gin.Context) (*api.ListDefinitionResponse, int32, string) {
	var req api.ListDefinitionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	if !tools.CheckPassword(req.Password) {
		return nil, code.ErrPassword, "wrong password"
	}
	// get bookID
	ctx := c.Request.Context()
	var bookIDs []int64
	for _, item := range tools.DefaultBookLevelConfig.Items {
		if req.CefrLevel != "" {
			if req.CefrLevel == item.Level {
				bookIDs = append(bookIDs, item.BookID)
				break
			}
		} else {
			bookIDs = append(bookIDs, item.BookID)
		}
	}
	params := &dao.DefinitionWithStringParams{
		BookIDs:  bookIDs,
		PageSize: req.PageSize,
		Page:     req.Page,
	}
	definitionWithStrings, total, err := s.dao.DefinitionDAO.SearchDefinitionWithString(ctx, params)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}
	var list []*api.ListDefinitionResponseItem
	for _, definitionWithString := range definitionWithStrings {
		list = append(list, &api.ListDefinitionResponseItem{
			BookID:       definitionWithString.BookID,
			DefinitionID: definitionWithString.ID,
			StringID:     definitionWithString.StringID,
			String:       definitionWithString.String,
			Type:         definitionWithString.Type,
			PartOfSpeech: definitionWithString.PartOfSpeech,
			Index:        definitionWithString.Idx - 1,
			Definition:   definitionWithString.Definition,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Index < list[j].Index
	})
	return &api.ListDefinitionResponse{
		List:  list,
		Total: total,
	}, 0, ""
}

// logger

func (s *BookService) Complete(err error, params ...any) {
	operateLog := &model.OperateLog{
		CreatedBy:          "", // operator or admin
		OperateStatus:      1,  // default is failure
		OperateType:        "",
		OtherOperateParams: "{}",
		BookID:             0,
	}
	if len(params) >= 1 {
		operateLog.CreatedBy = params[0].(string)
	}
	if len(params) >= 2 {
		operateLog.OperateType = params[1].(string)
	}
	if len(params) >= 3 {
		operateLog.BookID = params[2].(int64)
	}
	if len(params) >= 4 {
		operateLog.DefinitionID = params[3].(int64)
	}
	if err != nil {
		s.updating = false
		s.uploadError = err
		operateLog.Error = err.Error()
	} else {
		s.progress = 100
		s.updating = false
		operateLog.OperateStatus = 2 // success
	}
	if err := s.dao.OperateLogDAO.Create(context.TODO(), operateLog); err != nil {
		fmt.Println(err)
	}
}

func (s *BookService) VerbosePrint(v ...any) {
	if s.updating {
		s.pendingLogs = append(s.pendingLogs, fmt.Sprintln(v...))
	}
}

func (s *BookService) InfoPrint(v ...any) {
	if s.updating {
		fmt.Println(v...)
		s.pendingLogs = append(s.pendingLogs, fmt.Sprintln(v...))
	}
}

func (s *BookService) Progress(line int, total int) {
	if s.updating {
		s.progress = 5 + line*95/total
	}
}

func (s *BookService) CountDefinition(count int) {
	if s.updating {
		s.wordCount = count
	}
}

func (s *BookService) start() {
	s.updating = true
	s.pendingLogs = []string{}
	s.wordCount = 0
	s.progress = 0
	s.uploadError = nil
}
