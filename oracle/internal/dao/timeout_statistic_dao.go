package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/oracle/internal/model"
	"gorm.io/gorm"
)

type TimeoutStatisticDAO struct {
	db *gorm.DB
}

func NewTimeoutStatisticDAO(db *gorm.DB) *TimeoutStatisticDAO {
	return &TimeoutStatisticDAO{db: db}
}

func (obj *TimeoutStatisticDAO) GetTableName() string {
	return TableNameTimeoutStatistic
}

func (obj *TimeoutStatisticDAO) GetByDateAndPath(ctx context.Context, date time.Time, path string) (*TimeoutStatistic, error) {
	var result TimeoutStatistic
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("date = ?", date.Format("2006-01-02")).
		Where("path = ?", path).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *TimeoutStatisticDAO) Create(ctx context.Context, one *TimeoutStatistic) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *TimeoutStatisticDAO) UpdateCountByID(ctx context.Context, id int64, count int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("count", count).Error
	return err
}

func (obj *TimeoutStatisticDAO) GetPaginationByConditions(ctx context.Context, conditions map[string]interface{}, pageSize, pageNumber int) ([]*TimeoutStatistic, int64, error) {
	var list []*TimeoutStatistic
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).Where("deleted_at = 0")
	for k, v := range conditions {
		conn = conn.Where(k, v)
	}
	var total int64
	if err := conn.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := conn.Order("updated_at DESC").
		Offset(pageSize * pageNumber).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, total, err
	}
	return list, total, nil
}
