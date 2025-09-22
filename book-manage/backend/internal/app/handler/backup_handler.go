package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nextsurfer/book-manage-api/api"
	"github.com/nextsurfer/book-manage-api/internal/app/services"
)

type BackupHandler struct {
	backupService *services.BackupService
}

func NewBackupHandler() *BackupHandler {
	return &BackupHandler{
		backupService: services.NewBackupService(),
	}
}

func (h *BackupHandler) ListBackups(c *gin.Context) {
	data, errCode, errMsg := h.backupService.ListBackups(c)
	if errCode > 0 {
		fmt.Println("ListBackups failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BackupHandler) MakeBackup(c *gin.Context) {
	errCode, errMsg := h.backupService.MakeBackup(c)
	if errCode > 0 {
		fmt.Println("MakeBackup failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, nil)
	}
}

func (h *BackupHandler) RegainBackup(c *gin.Context) {
	errCode, errMsg := h.backupService.RegainBackup(c)
	if errCode > 0 {
		fmt.Println("RegainBackup failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, nil)
	}
}

func (h *BackupHandler) CheckRegainingLog(c *gin.Context) {
	data, errCode, errMsg := h.backupService.CheckRegainingLog(c)
	if errCode > 0 {
		fmt.Println("CheckRegainingLog failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BackupHandler) GetCronStatus(c *gin.Context) {
	data, errCode, errMsg := h.backupService.GetCronStatus(c)
	if errCode > 0 {
		fmt.Println("GetCronStatus failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}

func (h *BackupHandler) UpdateCronSetting(c *gin.Context) {
	data, errCode, errMsg := h.backupService.UpdateCronSetting(c)
	if errCode > 0 {
		fmt.Println("UpdateCronSetting failed, errCode: ", errCode, ", errMsg: ", errMsg)
		api.ErrorResponse(c, errCode, errMsg)
	} else {
		api.SuccessResponse(c, data)
	}
}
