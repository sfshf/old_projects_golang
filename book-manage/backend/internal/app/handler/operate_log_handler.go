package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nextsurfer/book-manage-api/api"
	"github.com/nextsurfer/book-manage-api/internal/app/services"
)

type OperateLogHandler struct {
	operateLogService *services.OperateLogService
}

func NewOperateLogHandler() *OperateLogHandler {
	return &OperateLogHandler{
		operateLogService: services.NewOperateLogService(),
	}
}

func (h *OperateLogHandler) GetPagination(c *gin.Context) {
	data, errCode, errMsg := h.operateLogService.GetPagination(c)
	if errCode > 0 {
		fmt.Println("GetPagination failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *OperateLogHandler) GetWorkingEthics(c *gin.Context) {
	data, errCode, errMsg := h.operateLogService.GetWorkingEthics(c)
	if errCode > 0 {
		fmt.Println("GetWorkingEthics failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *OperateLogHandler) GetPreviewLatestLogs(c *gin.Context) {
	data, errCode, errMsg := h.operateLogService.GetPreviewLatestLogs(c)
	if errCode > 0 {
		fmt.Println("GetPreviewLatestLogs failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}
