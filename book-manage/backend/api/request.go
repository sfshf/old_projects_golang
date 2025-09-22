package api

import (
	"time"

	"github.com/nextsurfer/book-manage-api/internal/app/dao"
	"github.com/nextsurfer/book-manage-api/internal/app/dto"
	"github.com/nextsurfer/book-manage-api/internal/app/model"
)

type Pagination struct {
	Page  int64 `json:"page" binding:"required,min=1"`
	Size  int64 `json:"size" binding:"required,min=1"`
	Total int64 `json:"total"`
}

// type AddBookResponseData struct {
// 	BookId int64 `json:"bookID"`
// }

type BookListRequest struct {
	Password string `json:"password" binding:"required"`
}

type BookListResponseData struct {
	Books []dto.BookDTO `json:"books"`
}

type DownloadResponseData struct {
	// only give the path of url
	Path string `json:"path"`
}

type SearchBookItemRequest struct {
	BookID int64 `json:"bookID" binding:"required"`
	Index  int   `json:"index" binding:"required"`
}

type SearchBookItemResponseData struct {
	Item  SearchBookItemResponseItem `json:"item"`
	Index int                        `json:"index"`
	Total int                        `json:"total"`
}

type SearchBookItemResponseItem struct {
	String                 string                   `json:"string"`
	Type                   string                   `json:"type"`
	SortValue              int32                    `json:"sortValue"`
	Definition             model.Definition         `json:"definition"`
	DefinitionTranslations []model.Translation      `json:"definitionTranslations"`
	Examples               []model.Example          `json:"examples"`
	ExampleTranslations    []model.Translation      `json:"exampleTranslations"`
	RelatedForms           []dao.RelatedForm        `json:"relatedForms"`
	DefinitionComment      *model.DefinitionComment `json:"definitionComment"`
}

type SearchStringPaginationResponseData struct {
	Items []*SearchStringPaginationResponseItem `json:"items"`
	Total int64                                 `json:"total"`
}

type SearchStringPaginationResponseItem struct {
	BookID       int64  `json:"bookID"`
	DefinitionID int64  `json:"definitionID"`
	String       string `json:"string"`
	Level        string `json:"level"`
	Type         string `json:"type"`
	PartOfSpeech string `json:"partOfSpeech"`
	Index        int64  `json:"index"`
	Definition   string `json:"definition"`
}

type GetCefrLevelsRequest struct {
	Password string `form:"password" binding:"required"`
	CurLevel string `form:"curLevel" binding:""`
	BookID   int64  `form:"bookID" binding:""`
}

type GetCefrLevelsResponseData struct {
	List []GetCefrLevelsResponseItem `json:"list"`
}

type GetCefrLevelsResponseItem struct {
	BookID int64  `json:"bookID"`
	Level  string `json:"level"`
}

type GetNextSortValueResponseData struct {
	NextSortValue int `json:"nextSortValue"`
}

type UpdateCefrLevelRequest struct {
	Password     string `json:"password" binding:"required"`
	BookID       int64  `json:"bookID" binding:"required"`
	DefinitionID int64  `json:"definitionID" binding:"required"`
	CefrLevel    string `json:"cefrLevel" binding:"required"`
	SortValue    int    `json:"sortValue" binding:""`
}

type GetDefinitionInfoRequest struct {
	Password     string `form:"password" binding:"required"`
	DefinitionID int64  `form:"definitionID" binding:"required"`
}

type GetDefinitionInfoResponse struct {
	BookID int64 `json:"bookID"`
	Index  int   `json:"index"`
}

type NewDefinitionRequest struct {
	Password              string `json:"password" binding:"required"`
	BookID                int64  `json:"bookID" binding:"required"`
	CefrLevel             string `json:"cefrLevel" binding:""`
	String                string `json:"string" binding:"required"`
	Type                  string `json:"type" binding:"oneof=word phrase"`
	SortValue             int32  `json:"sortValue" binding:""`
	Definition            string `json:"definition" binding:"required"`
	PartOfSpeech          string `json:"partOfSpeech" binding:""`
	SpecificType          string `json:"specificType" binding:""`
	PronunciationIpa      string `json:"pronunciationIpa" binding:""`
	PronunciationIpaWeak  string `json:"pronunciationIpaWeak" binding:""`
	PronunciationIpaOther string `json:"pronunciationIpaOther" binding:""`
	Forms                 []struct {
		Form          string `json:"form" binding:""`
		FormString    string `json:"formString" binding:""`
		Pronunciation string `json:"pronunciation" binding:""`
	} `json:"forms" binding:""`
	PronunciationText string `json:"pronunciationText" binding:""`
	Example1          string `json:"example1" binding:""`
	Positions1        string `json:"positions1" binding:""`
	Example2          string `json:"example2" binding:""`
	Positions2        string `json:"positions2" binding:""`
	Example3          string `json:"example3" binding:""`
	Positions3        string `json:"positions3" binding:""`
}

type ListDefinitionRequest struct {
	Password  string `form:"password" binding:"required"`
	CefrLevel string `form:"cefrLevel" binding:""`
	PageSize  int    `form:"pageSize" binding:"gt=0"`
	Page      int    `form:"page" binding:"gte=0"`
}

type ListDefinitionResponse struct {
	List  []*ListDefinitionResponseItem `json:"list"`
	Total int64                         `json:"total"`
}

type ListDefinitionResponseItem struct {
	BookID       int64  `json:"bookID"`
	StringID     int64  `json:"stringID"`
	String       string `json:"string"`
	PartOfSpeech string `json:"partOfSpeech"`
	DefinitionID int64  `json:"definitionID"`
	Definition   string `json:"definition"`
	Type         string `json:"type"`
	Index        int64  `json:"index"`
}

type UploadingLogResponseData struct {
	Logs      []string `json:"logs"`
	WordCount int      `json:"wordCount"`
	Progress  int      `json:"progress"`
	Error     string   `json:"error"`
}

type OperateLogPaginationResponseData struct {
	OperateLogs []dto.OperateLogDTO `json:"operateLogs"`
	Total       int64               `json:"total"`
}

type UpdatePreviewRequest struct {
	Password     string `json:"password" binding:"required"`
	DefinitionID int64  `json:"definitionID" binding:"required"`
	BookID       int64  `json:"bookID" binding:"required"`
	Field        string `json:"field" binding:"required"`

	StringID int64  `json:"stringID" binding:"required_if=Field string"`
	String   string `json:"string" binding:"required_if=Field string"`

	Definition string `json:"definition" binding:"required_if=Field definition"`

	PartOfSpeech string `json:"partOfSpeech" binding:"required_if=Field part_of_speech"`

	SpecificType string `json:"specificType" binding:""`

	PronunciationIpa      string `json:"pronunciationIpa" binding:""`
	PronunciationIpaWeak  string `json:"pronunciationIpaWeak" binding:""`
	PronunciationIpaOther string `json:"pronunciationIpaOther" binding:""`
	PronunciationText     string `json:"pronunciationText" binding:""`

	ExampleID int64  `json:"exampleID" binding:""`
	Example   string `json:"example" binding:"required_if=Field example_1,required_if=Field example_2,required_if=Field example_3"`
	Position  string `json:"position" binding:"required_with=Example"`

	DefinitionCommentID int64  `json:"definitionCommentID" binding:""`
	DefinitionComment   string `json:"definitionComment" binding:""`

	FormStringID  int64  `json:"formStringID" binding:""`
	Form          string `json:"form" binding:"required_if=Field form"`
	FormString    string `json:"formString" binding:"required_if=Field form"`
	Pronunciation string `json:"pronunciation" binding:""`

	SortValue int32 `json:"sortValue" binding:"required_if=Field sort_value,gte=0"`

	TranslationID      int64  `json:"translationID" binding:""`
	TranslationContent string `json:"translationContent" binding:"required_if=Field definition_translation,required_if=Field example_translation"`
	LanguageCode       string `json:"languageCode" binding:"required_if=Field definition_translation,required_if=Field example_translation"`
}

type DeletePreviewRequest struct {
	Password     string `json:"password" binding:"required"`
	DefinitionID int64  `json:"definitionID" binding:"required"`
	BookID       int64  `json:"bookID" binding:"required"`
	Field        string `json:"field" binding:"required"`

	ExampleID int64 `json:"exampleID" binding:"required_if=Field example_1,required_if=Field example_2,required_if=Field example_3,required_if=Field example_translation"`

	DefinitionCommentID int64 `json:"definitionCommentID" binding:"required_if=Field definition_comment"`

	FormStringID int64 `json:"formStringID" binding:"required_if=Field form"`

	TranslationId int64 `json:"translationID" binding:"required_if=Field definition_translation,required_if=Field example_translation"`
}

type GetWorkingEthicsResponseData struct {
	List          []WorkingEthicData `json:"list"`
	TotalDuration int                `json:"totalDuration"`
}

type WorkingEthicData struct {
	OperateLogs []dto.OperateLogDTO `json:"operateLogs"`
	Duration    int                 `json:"duration"` // in seconds
}

type MakeBackupRequest struct {
	Password string `json:"password" binding:"required"`
	Book     string `json:"book" binding:"required"`
}

type RegainBackupRequest struct {
	Password string `json:"password" binding:"required"`
	BookID   int64  `json:"bookID" binding:"required"`
	BackupID int64  `json:"backupID" binding:"required"`
}

type GetCronStatusResponseData struct {
	Started           bool      `json:"started"`
	StartedOrStopedAt time.Time `json:"startedOrStopedAt"`
	NextTime          time.Time `json:"nextTime"`
	ScheduleSpec      string    `json:"scheduleSpec"`
	LastExecError     string    `json:"lastExecError"`
}

type SetCronJobRequest struct {
	Password     string `json:"password" binding:"required"`
	StartCron    string `json:"startCron" binding:"omitempty,oneof=start stop"`
	ScheduleSpec string `json:"scheduleSpec" binding:"omitempty,cron"`
}
