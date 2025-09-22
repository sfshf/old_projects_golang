package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nextsurfer/book-manage-api/api"
	"github.com/nextsurfer/book-manage-api/internal/app/services"
)

// BookHandler deposit address handler
type BookHandler struct {
	bookService *services.BookService
}

// NewBookService is factory function
func NewBookHandler() *BookHandler {
	return &BookHandler{
		bookService: services.NewBookService(),
	}
}

func (h *BookHandler) GetAllBooks(c *gin.Context) {
	data, errCode, errMsg := h.bookService.GetAllBooks(c)
	if errCode > 0 {
		fmt.Println("GetAllBooks failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) GetCSV(c *gin.Context) {
	data, errCode, errMsg := h.bookService.GetCSV(c)
	if errCode > 0 {
		fmt.Println("GetCSV failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) GetBundle(c *gin.Context) {
	data, errCode, errMsg := h.bookService.GetBundle(c)
	if errCode > 0 {
		fmt.Println("GetBundle failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) AddBook(c *gin.Context) {
	errCode, errMsg := h.bookService.AddBook(c)
	if errCode > 0 {
		fmt.Println("AddBook failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, nil)
	}
}

func (h *BookHandler) UpdateBook(c *gin.Context) {
	errCode, errMsg := h.bookService.UpdateBook(c)
	if errCode > 0 {
		fmt.Println("UpdateBook failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, nil)
	}
}

func (h *BookHandler) GetBook(c *gin.Context) {
	data, errCode, errMsg := h.bookService.SearchBookItem(c)
	if errCode > 0 {
		fmt.Println("GetBook failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) SearchBookItem(c *gin.Context) {
	data, errCode, errMsg := h.bookService.SearchBookItem(c)
	if errCode > 0 {
		fmt.Println("SearchBookItem failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) TextToSpeach(c *gin.Context) {
	data, errCode, errMsg := h.bookService.TextToSpeach(c)
	if errCode > 0 {
		fmt.Println("TextToSpeach failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) GetUploadingLog(c *gin.Context) {
	data, errCode, errMsg := h.bookService.UploadingLog(c)
	if errCode > 0 {
		fmt.Println("GetUploadingLog failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) UpdatePreview(c *gin.Context) {
	data, errCode, errMsg := h.bookService.UpdatePreview(c)
	if errCode > 0 {
		fmt.Println("UpdatePreview failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) GetExamplePosition(c *gin.Context) {
	data, errCode, errMsg := h.bookService.GetExamplePosition(c)
	if errCode > 0 {
		fmt.Println("GetExamplePosition failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) DeletePreview(c *gin.Context) {
	errCode, errMsg := h.bookService.DeletePreview(c)
	if errCode > 0 {
		fmt.Println("DeletePreview failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, nil)
	}
}

func (h *BookHandler) SearchStringPagination(c *gin.Context) {
	data, errCode, errMsg := h.bookService.SearchStringPagination(c)
	if errCode > 0 {
		fmt.Println("SearchStringPagination failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) GetCefrLevels(c *gin.Context) {
	data, errCode, errMsg := h.bookService.GetCefrLevels(c)
	if errCode > 0 {
		fmt.Println("GetCefrLevels failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) GetNextSortValue(c *gin.Context) {
	data, errCode, errMsg := h.bookService.GetNextSortValue(c)
	if errCode > 0 {
		fmt.Println("GetNextSortValue failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) UpdateCefrLevel(c *gin.Context) {
	errCode, errMsg := h.bookService.UpdateCefrLevel(c)
	if errCode > 0 {
		fmt.Println("UpdateCefrLevel failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, nil)
	}
}

func (h *BookHandler) GetDefinitionInfo(c *gin.Context) {
	data, errCode, errMsg := h.bookService.GetDefinitionInfo(c)
	if errCode > 0 {
		fmt.Println("GetDefinitionInfo failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BookHandler) NewDefinition(c *gin.Context) {
	errCode, errMsg := h.bookService.NewDefinition(c)
	if errCode > 0 {
		fmt.Println("NewDefinition failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, nil)
	}
}

func (h *BookHandler) ListDefinition(c *gin.Context) {
	data, errCode, errMsg := h.bookService.ListDefinition(c)
	if errCode > 0 {
		fmt.Println("ListDefinition failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}
