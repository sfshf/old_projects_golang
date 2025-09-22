package dao

import (
	"context"
	"time"

	. "github.com/nextsurfer/oracle/internal/model"
	"gorm.io/gorm"
)

type RateLimitRuleDAO struct {
	db *gorm.DB
}

func NewRateLimitRuleDAO(db *gorm.DB) *RateLimitRuleDAO {
	return &RateLimitRuleDAO{db: db}
}

func (obj *RateLimitRuleDAO) GetTableName() string {
	return TableNameRateLimitRule
}

func (obj *RateLimitRuleDAO) GetAll(ctx context.Context) ([]*RateLimitRule, error) {
	var list []*RateLimitRule
	if err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("deleted_at = 0").
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (obj *RateLimitRuleDAO) GetServiceRuleByName(ctx context.Context, target string) (*RateLimitRule, error) {
	var result RateLimitRule
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("type = ?", 1).
		Where("target = ?", target).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *RateLimitRuleDAO) GetPathRuleByName(ctx context.Context, target string) (*RateLimitRule, error) {
	var result RateLimitRule
	conn := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("type = ?", 2).
		Where("target = ?", target).
		Where("deleted_at = 0").
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *RateLimitRuleDAO) Create(ctx context.Context, one *RateLimitRule) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).Create(one).Error
	return err
}

func (obj *RateLimitRuleDAO) DeleteByID(ctx context.Context, id int64) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Where("id = ?", id).
		Where("deleted_at = 0").
		Update("deleted_at", time.Now().UnixMilli()).Error
	return err
}

func (obj *RateLimitRuleDAO) Update(ctx context.Context, one *RateLimitRule) error {
	err := obj.db.WithContext(ctx).Table(obj.GetTableName()).
		Select("*").
		Omit("created_at").
		Where("id = ?", one.ID).
		Where("deleted_at = 0").
		Updates(one).Error
	return err
}
