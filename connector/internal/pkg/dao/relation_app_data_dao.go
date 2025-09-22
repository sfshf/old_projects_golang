package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/connector/internal/pkg/model"
	"gorm.io/gorm"
)

type RelationAppDatumDAO struct {
	db *gorm.DB
}

func NewRelationAppDatumDAO(db *gorm.DB) *RelationAppDatumDAO {
	return &RelationAppDatumDAO{db: db}
}

func (obj *RelationAppDatumDAO) GetTableName() string {
	return TableNameRelationAppDatum
}

func (obj *RelationAppDatumDAO) Create(ctx context.Context, one *RelationAppDatum) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *RelationAppDatumDAO) GetByAppWithKeyIDAndDataID(ctx context.Context, app, keyID, dataID string) (*RelationAppDatum, error) {
	var result RelationAppDatum
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("app = ?", app).
		Where("key_id = ?", keyID).
		Where("data_id = ?", dataID).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *RelationAppDatumDAO) DeleteByAppWithKeyIDAndDataID(ctx context.Context, app, keyID, dataID string) error {
	return obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where(`app = ?`, app).
		Where(`key_id = ?`, keyID).
		Where("data_id = ?", dataID).
		Where(`deleted_at = 0`).
		Update("deleted_at", time.Now().UnixMilli()).Error
}
