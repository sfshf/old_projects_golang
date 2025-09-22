package dao

import (
	"context"

	. "github.com/nextsurfer/oracle/internal/model"
	"gorm.io/gorm"
)

type GatewayNodeDAO struct {
	db *gorm.DB
}

func NewGatewayNodeDAO(db *gorm.DB) *GatewayNodeDAO {
	return &GatewayNodeDAO{db: db}
}

func (obj *GatewayNodeDAO) GetTableName() string {
	return TableNameGatewayNode
}

func (obj *GatewayNodeDAO) Create(ctx context.Context, one *GatewayNode) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *GatewayNodeDAO) GetAll(ctx context.Context) ([]*GatewayNode, error) {
	var list []*GatewayNode
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName())
	if err := conn.Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// real delete
func (obj *GatewayNodeDAO) RemoveAll(ctx context.Context) error {
	return obj.db.Delete(&GatewayNode{}, "1=1").Error
}

func (obj *GatewayNodeDAO) GetByName(ctx context.Context, name string) (*GatewayNode, error) {
	var result GatewayNode
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Omit("proto_file").
		Where("name = ?", name).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}
