package dao

import (
	"context"

	. "github.com/nextsurfer/oracle/internal/model"
	"gorm.io/gorm"
)

type AcmeResourceDAO struct {
	db *gorm.DB
}

func NewAcmeResourceDAO(db *gorm.DB) *AcmeResourceDAO {
	return &AcmeResourceDAO{db: db}
}

func (obj *AcmeResourceDAO) GetTableName() string {
	return TableNameAcmeResource
}

func (obj *AcmeResourceDAO) GetAll(ctx context.Context) ([]*AcmeResource, error) {
	var list []*AcmeResource
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *AcmeResourceDAO) GetByDomain(ctx context.Context, domain string) (*AcmeResource, error) {
	var result AcmeResource
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

func (obj *AcmeResourceDAO) GetByToken(ctx context.Context, token string) (*AcmeResource, error) {
	var result AcmeResource
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("token = ?", token).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *AcmeResourceDAO) GetByID(ctx context.Context, id int64) (*AcmeResource, error) {
	var result AcmeResource
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

func (obj *AcmeResourceDAO) Create(ctx context.Context, one *AcmeResource) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *AcmeResourceDAO) DeleteByDomain(ctx context.Context, domain string) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Delete(&AcmeResource{}, "domain = ?", domain).Error
	return err
}

func (obj *AcmeResourceDAO) Update(ctx context.Context, one *AcmeResource) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", one.ID).
		Where("deleted_at = 0").
		Updates(one).Error
	return err
}
