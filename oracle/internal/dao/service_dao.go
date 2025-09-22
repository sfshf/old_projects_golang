package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/oracle/internal/model"
	"gorm.io/gorm"
)

type ServiceDAO struct {
	db *gorm.DB
}

func NewServiceDAO(db *gorm.DB) *ServiceDAO {
	return &ServiceDAO{db: db}
}

func (obj *ServiceDAO) GetTableName() string {
	return TableNameService
}

func (obj *ServiceDAO) Create(ctx context.Context, one *Service) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *ServiceDAO) GetAllPrerequisites(ctx context.Context) ([]*Service, error) {
	var list []*Service
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Omit("proto_file").
		Where("application_id = 0").
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *ServiceDAO) GetAllServices(ctx context.Context, omitProtoFile, omitFileDescriptor bool) ([]*Service, error) {
	var list []*Service
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName())
	if omitProtoFile {
		conn = conn.Omit("proto_file")
	}
	if omitFileDescriptor {
		conn = conn.Omit("file_descriptor_data")
	}
	if err := conn.Where("application_id > 0").
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *ServiceDAO) DeleteByName(ctx context.Context, name string) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("name = ?", name).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
}

func (obj *ServiceDAO) DeleteAllByApplicationID(ctx context.Context, applicationID int64) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("application_id = ?", applicationID).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
}

func (obj *ServiceDAO) GetProtoFileByName(ctx context.Context, name string) (string, error) {
	var result Service
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Select("proto_file").
		Where("name = ?", name).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return "", conn.Error
	} else {
		return result.ProtoFile, nil
	}
}

func (obj *ServiceDAO) GetByName(ctx context.Context, name string, omitProtoFile, omitDeleted bool) (*Service, error) {
	var result Service
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("name = ?", name)
	if omitProtoFile {
		conn = conn.Omit("proto_file")
	}
	if omitDeleted {
		conn = conn.Where("deleted_at = 0")
	}
	conn.Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *ServiceDAO) GetByID(ctx context.Context, id int64, omitProtoFile, omitFileDescriptor bool) (*Service, error) {
	var result Service
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName())
	if omitProtoFile {
		conn = conn.Omit("proto_file")
	}
	if omitFileDescriptor {
		conn = conn.Omit("file_descriptor_data")
	}
	conn.Where("id = ?", id).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *ServiceDAO) CountByApplicationID(ctx context.Context, applicationID int64) (int64, error) {
	var result int64
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Omit("proto_file").
		Where("application_id = ?", applicationID).
		Where("deleted_at = 0").
		Count(&result).Error; err != nil {
		return 0, err
	}
	return result, nil
}

func (obj *ServiceDAO) GetByPathPrefix(ctx context.Context, pathPrefix string) (*Service, error) {
	var result Service
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Omit("proto_file").
		Where("path_prefix = ?", pathPrefix).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *ServiceDAO) Update(ctx context.Context, one *Service) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", one.ID).
		Where("deleted_at = 0").
		Updates(one).Error
}
