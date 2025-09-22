package dto

type BookDTO struct {
	ID          int64  `json:"id"`
	UpdatedAt   int64  `json:"updatedAt"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DownloadURL string `json:"downloadURL"`
}
