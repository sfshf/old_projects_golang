package dao

import (
	"context"

	. "github.com/nextsurfer/oracle/internal/model"
	"gorm.io/gorm"
)

type ProtoStatisticHourlyDAO struct {
	db *gorm.DB
}

func NewProtoStatisticHourlyDAO(db *gorm.DB) *ProtoStatisticHourlyDAO {
	return &ProtoStatisticHourlyDAO{db: db}
}

func (obj *ProtoStatisticHourlyDAO) GetTableName() string {
	return TableNameProtoStatisticHourly
}

func (obj *ProtoStatisticHourlyDAO) Create(ctx context.Context, value interface{}) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(value).Error
	return err
}
func (obj *ProtoStatisticHourlyDAO) Delete(ctx context.Context, conds ...interface{}) error {
	return obj.db.Delete(&ProtoStatisticHourly{}, conds...).Error
}
