package services

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nextsurfer/book-manage-api/api"
	"github.com/nextsurfer/book-manage-api/api/code"
	"github.com/nextsurfer/book-manage-api/internal/app/dao"
	"github.com/nextsurfer/book-manage-api/internal/app/dto"
	"github.com/nextsurfer/book-manage-api/internal/tools"
)

type OperateLogService struct {
	dao *dao.Manager
}

func NewOperateLogService() *OperateLogService {
	return &OperateLogService{
		dao: dao.NewManagerWithDB(tools.MysqlDB()),
	}
}

func (s *OperateLogService) GetPagination(c *gin.Context) (*api.OperateLogPaginationResponseData, int32, string) {
	password := c.Query("password")

	// password check
	if !tools.CheckAdminPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}
	var params dao.OperateLogPaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		return nil, code.ErrWrongParam, err.Error()
	}

	data := &api.OperateLogPaginationResponseData{}
	operateLogs, total, err := s.dao.OperateLogDAO.GetPagination(context.Background(), &params)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	} else {
		l := len(operateLogs)
		operateDTOs := make([]dto.OperateLogDTO, 0, l)
		for i := 0; i < l; i++ {
			operateDTOs = append(operateDTOs, dto.OperateLogDTO{
				ID:                 operateLogs[i].ID,
				Operator:           operateLogs[i].CreatedBy,
				OperateTime:        tools.Time{Time: operateLogs[i].CreatedAt},
				OperateStatus:      operateLogs[i].OperateStatus,
				OperateType:        operateLogs[i].OperateType,
				BookID:             operateLogs[i].BookID,
				DefinitionID:       operateLogs[i].DefinitionID,
				OtherOperateParams: operateLogs[i].OtherOperateParams,
				Error:              operateLogs[i].Error,
			})
		}
		data.OperateLogs = operateDTOs
		data.Total = total
		return data, code.Ok, ""
	}
}

func (s *OperateLogService) GetWorkingEthics(c *gin.Context) (*api.GetWorkingEthicsResponseData, int32, string) {
	password := c.Query("password")

	// password check
	if !tools.CheckAdminPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}
	var params dao.GetWorkingEthicsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		return nil, code.ErrWrongParam, err.Error()
	}

	data := &api.GetWorkingEthicsResponseData{}
	workingEthics, err := s.dao.OperateLogDAO.GetWorkingEthics(context.Background(), &params)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	} else {
		var list []api.WorkingEthicData
		var totalDuration int
		for i := 0; i < len(workingEthics); i++ {
			var workingEthic api.WorkingEthicData
			var operateLogs []dto.OperateLogDTO
			for idx, log := range workingEthics[i] {
				operateTime := log.CreatedAt
				logDTO := dto.OperateLogDTO{
					ID:                 log.ID,
					OperateTime:        tools.Time{Time: operateTime},
					Operator:           log.CreatedBy,
					OperateStatus:      log.OperateStatus,
					OperateType:        log.OperateType,
					BookID:             log.BookID,
					DefinitionID:       log.DefinitionID,
					OtherOperateParams: log.OtherOperateParams,
					Error:              log.Error,
				}
				if idx == 0 {
					logDTO.OperateTime = tools.Time{Time: operateTime.Add(time.Duration(1) * time.Minute)}
				}
				operateLogs = append(operateLogs, logDTO)
				if len(workingEthics[i]) == 1 {
					logDTO.OperateTime = tools.Time{Time: operateTime}
					operateLogs = append(operateLogs, logDTO)
				}
			}
			workingEthic.OperateLogs = operateLogs
			workingEthic.Duration = int(operateLogs[0].OperateTime.Time.Sub(operateLogs[len(operateLogs)-1].OperateTime.Time) / time.Second)
			totalDuration += workingEthic.Duration
			list = append(list, workingEthic)
		}
		data.List = list
		data.TotalDuration = totalDuration
		return data, code.Ok, ""
	}
}

func (s *OperateLogService) GetPreviewLatestLogs(c *gin.Context) (*api.OperateLogPaginationResponseData, int32, string) {
	password := c.Query("password")

	// password check
	if !tools.CheckAdminPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}
	var params dao.OperateLogPaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		return nil, code.ErrWrongParam, err.Error()
	}
	if params.DefinitionID <= 0 {
		return nil, code.ErrWrongParam, "invalid DefinitionID"
	}
	params = dao.OperateLogPaginationParams{
		DefinitionID: params.DefinitionID,
		Page:         0,
		PageSize:     5,
		Order:        "desc",
		OrderBy:      "created_at",
	}
	data := &api.OperateLogPaginationResponseData{}
	operateLogs, total, err := s.dao.OperateLogDAO.GetPagination(context.Background(), &params)
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	} else {
		l := len(operateLogs)
		operateDTOs := make([]dto.OperateLogDTO, 0, l)
		for i := 0; i < l; i++ {
			operateDTOs = append(operateDTOs, dto.OperateLogDTO{
				ID:            operateLogs[i].ID,
				Operator:      operateLogs[i].CreatedBy,
				OperateTime:   tools.Time{Time: operateLogs[i].CreatedAt},
				OperateStatus: operateLogs[i].OperateStatus,
				OperateType:   operateLogs[i].OperateType,
			})
		}
		data.OperateLogs = operateDTOs
		data.Total = total
		return data, code.Ok, ""
	}
}
