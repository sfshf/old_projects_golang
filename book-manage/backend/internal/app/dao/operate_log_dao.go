package dao

import (
	"context"
	"fmt"
	"time"

	. "github.com/nextsurfer/book-manage-api/internal/app/model"
	"gorm.io/gorm"
)

// OperateLogDAO is a dao service
type OperateLogDAO struct {
	db *gorm.DB
}

// NewBookDAO to create a dao service
func NewOperateLogDAO(db *gorm.DB) *OperateLogDAO {
	return &OperateLogDAO{db: db}
}

// GetTableName get sql table name.
func (obj *OperateLogDAO) GetTableName() string {
	return TableNameOperateLog
}

type OperateLogPaginationParams struct {
	Operator      string `form:"operator" json:"operator" binding:""`           // 根据操作员过滤
	BookID        int64  `form:"bookID" json:"bookID" binding:""`               // 根据bookID过滤
	OperateStatus int32  `form:"operateStatus" json:"operateStatus" binding:""` // 根据是否成功过滤
	DefinitionID  int64  `form:"definitionID" json:"definitionID" binding:""`   // 根据definitionID过滤
	Page          int    `form:"page" json:"page" binding:""`                   // 分页-当前页号
	PageSize      int    `form:"pageSize" json:"pageSize" binding:""`           // 分页-每页数量
	Order         string `form:"order" json:"order" binding:""`                 // 排序顺序-asc/desc
	OrderBy       string `form:"orderBy" json:"orderBy" binding:""`             // 排序字段
}

func (obj *OperateLogDAO) GetPagination(ctx context.Context, queries *OperateLogPaginationParams) ([]OperateLog, int64, error) {
	list := make([]OperateLog, 0)
	var total int64
	tb := obj.db.WithContext(ctx).Table(obj.GetTableName())
	if queries.Operator != "" {
		tb = tb.Where(`created_by = ?`, queries.Operator)
	}
	if queries.BookID > 0 {
		tb = tb.Where(`book_id = ?`, queries.BookID)
	}
	if queries.OperateStatus > 0 {
		tb = tb.Where(`operate_status = ?`, queries.OperateStatus)
	}
	if queries.DefinitionID > 0 {
		tb = tb.Where(`definition_id = ?`, queries.DefinitionID)
	}
	if err := tb.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if queries.PageSize > 100 {
		queries.PageSize = 100
	}
	if queries.Order != "" && queries.OrderBy != "" {
		tb = tb.Order(fmt.Sprintf("%s %s", queries.OrderBy, queries.Order))
	}
	if err := tb.Offset(queries.Page * queries.PageSize).
		Limit(queries.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

type GetWorkingEthicsParams struct {
	Operator string `form:"operator" json:"operator" binding:"required"`             // 根据操作员过滤
	Date     string `form:"date" json:"date" binding:"required,datetime=2006-01-02"` // date
}

func (obj *OperateLogDAO) GetWorkingEthics(ctx context.Context, queries *GetWorkingEthicsParams) ([][]OperateLog, error) {
	dailyLogs := make([]OperateLog, 0)

	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where(`created_by = ?`, queries.Operator).
		Where(`DATE(created_at) = ?`, queries.Date).
		Order(`created_at desc`).
		Find(&dailyLogs).Error; err != nil {
		return nil, err
	}

	ethics := make([][]OperateLog, 0)
	var dur []OperateLog
	for idx, log := range dailyLogs {
		if idx == 0 {
			dur = append(dur, log)
			continue
		}
		tp := dailyLogs[idx-1].CreatedAt.Add(time.Duration(-1) * time.Minute)
		if log.CreatedAt.Equal(tp) || log.CreatedAt.After(tp) {
			dur = append(dur, log)
		} else {
			ethics = append(ethics, dur)
			dur = make([]OperateLog, 0)
			dur = append(dur, log)
		}
	}
	if len(dur) > 0 {
		ethics = append(ethics, dur)
	}

	return ethics, nil
}

func (obj *OperateLogDAO) Create(ctx context.Context, operateLog *OperateLog) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(operateLog).Error
}

func (obj *OperateLogDAO) Update(ctx context.Context, operateLog *OperateLog) error {
	conn := obj.db.WithContext(ctx).Model(operateLog).
		Where(`deleted_at = 0`).
		Updates(operateLog)
	if conn.RowsAffected != 1 {
		return fmt.Errorf("update item with id %d failed in table %s ", operateLog.ID, obj.GetTableName())
	}
	return conn.Error
}
