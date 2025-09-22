package dao

import (
	"context"

	. "github.com/nextsurfer/oracle/internal/model"
	"gorm.io/gorm"
)

type HostManageDAO struct {
	db *gorm.DB
}

func NewHostManageDAO(db *gorm.DB) *HostManageDAO {
	return &HostManageDAO{db: db}
}

func (obj *HostManageDAO) GetTableName() string {
	return TableNameHostManage
}

func (obj *HostManageDAO) Create(ctx context.Context, one *HostManage) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *HostManageDAO) Update(ctx context.Context, one *HostManage) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", one.ID).
		Where("deleted_at = 0").
		Updates(one).Error
	return err
}

func (obj *HostManageDAO) GetAll(ctx context.Context) ([]*HostManage, error) {
	var list []*HostManage
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *HostManageDAO) GetByID(ctx context.Context, id int64) (*HostManage, error) {
	var result HostManage
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *HostManageDAO) GetByDomain(ctx context.Context, domain string) (*HostManage, error) {
	var result HostManage
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("domain = ?", domain).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *HostManageDAO) DeleteByID(ctx context.Context, id int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Delete(&HostManage{ID: id}).Error
	return err
}
