package dto

type ExampleDTO struct {
	ID           int64 `json:"id"`
	UpdatedAt    int64 `json:"updatedAt"`
	StringID     int64 `json:"stringID"`
	BookID       int64 `json:"bookID"`
	DefinitionID int64 `json:"definitionID"`

	Content       string `json:"content"`
	WordPositions string `json:"wordPositions"`
	SortValue     int32  `json:"sortValue"`
}
