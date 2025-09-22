package dao

import (
	"context"

	. "github.com/nextsurfer/oracle/internal/model"
	"gorm.io/gorm"
)

type ProtoStatisticDAO struct {
	db *gorm.DB
}

func NewProtoStatisticDAO(db *gorm.DB) *ProtoStatisticDAO {
	return &ProtoStatisticDAO{db: db}
}

func (obj *ProtoStatisticDAO) GetTableName() string {
	return TableNameProtoStatistic
}

func (obj *ProtoStatisticDAO) GetByDateAndPath(ctx context.Context, date string, path string) (*ProtoStatistic, error) {
	var result ProtoStatistic
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("date = ?", date).
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

func (obj *ProtoStatisticDAO) Create(ctx context.Context, one *ProtoStatistic) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *ProtoStatisticDAO) Update(ctx context.Context, one *ProtoStatistic) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", one.ID).
		Where("deleted_at = 0").
		Updates(one).Error
	return err
}

func (obj *ProtoStatisticDAO) GetPaginationByConditions(ctx context.Context, conditions map[string]interface{}, pageSize, pageNumber int, aggregate bool) ([]*ProtoStatistic, int64, error) {
	var list []*ProtoStatistic
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).Where("deleted_at = 0")
	for k, v := range conditions {
		conn = conn.Where(k, v)
	}
	if aggregate {
		subquery := conn.
			Select(
				`path,
				SUM(hit) AS hit,
				SUM(success_hit) AS success_hit,
				SUM(proxy_success_hit) AS proxy_success_hit,
				ROUND(SUM(hit * duration_average)/SUM(hit)) AS duration_average,
				MIN(NULLIF(duration_min,0)) AS duration_min,
				MAX(duration_max) AS duration_max,
				ROUND(SUM(hit * service_duration_average)/SUM(hit)) AS service_duration_average,
				MIN(NULLIF(service_duration_min,0)) AS service_duration_min,
				MAX(service_duration_max) AS service_duration_max`,
			).Group("path")
		conn = obj.db.WithContext(ctx).Table("(?) AS t", subquery).
			Joins(`LEFT JOIN proto_statistic ON proto_statistic.path=t.path`).
			Select(`
				DISTINCT 
				proto_statistic.application_id, 
				proto_statistic.service_id, 
				t.path, 
				t.hit, 
				t.success_hit, 
				t.proxy_success_hit, 
				t.duration_average, 
				t.duration_min, 
				t.duration_max, 
				t.service_duration_average, 
				t.service_duration_min, 
				t.service_duration_max`,
			)
	} else {
		conn = conn.Order("updated_at DESC")
	}
	var total int64
	if err := conn.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := conn.
		Offset(pageSize * pageNumber).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, total, err
	}
	return list, total, nil
}
