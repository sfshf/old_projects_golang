package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nextsurfer/book-manage-api/api"
	"github.com/nextsurfer/book-manage-api/internal/app/services"
)

type SystemHandler struct {
	systemService *services.SystemService
}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		systemService: services.NewSystemService(),
	}
}

func (h *SystemHandler) CheckAdmin(c *gin.Context) {
	data, errCode, errMsg := h.systemService.CheckAdmin(c)
	if errCode > 0 {
		fmt.Println("CheckAdmin failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *SystemHandler) ListStaff(c *gin.Context) {
	data, errCode, errMsg := h.systemService.ListStaff(c)
	if errCode > 0 {
		fmt.Println("ListStaff failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}
