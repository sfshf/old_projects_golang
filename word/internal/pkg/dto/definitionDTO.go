package dto

type DefinitionDTO struct {
	ID        int64 `json:"id"`
	UpdatedAt int64 `json:"updatedAt"`
	StringID  int64 `json:"stringID"`
	BookID    int64 `json:"bookID"`

	String                string `json:"string"`
	PartOfSpeech          string `json:"partOfSpeech"`
	SpecificType          string `json:"specificType,omitempty"`
	PronunciationIPA      string `json:"pronunciationIPA,omitempty"`
	WeakPronunciationIPA  string `json:"weakPronunciationIPA,omitempty"`
	OtherPronunciationIPA string `json:"otherPronunciationIPA,omitempty"`
	PronunciationText     string `json:"pronunciationText,omitempty"`
	SortValue             int32  `json:"sortValue"`
	Level                 string `json:"level,omitempty"`
	Definition            string `json:"definition"`
}

type RelatedFormDTO struct {
	FormStringID     int64 `json:"formStringID"`
	FormDefinitionID int64 `json:"formDefinitionID"`
	// UpdatedAt        int64 `json:"updatedAt"`

	WordStringID     int64 `json:"wordStringID"`
	WordDefinitionID int64 `json:"wordDefinitionID"`
	BookID           int64 `json:"bookID"`

	String           string `json:"string"`
	Form             string `json:"form"`
	PronunciationIPA string `json:"pronunciationIPA,omitempty"`
	Definition       string `json:"definition"`
}
