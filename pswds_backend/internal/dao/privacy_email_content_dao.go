package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type PrivacyEmailContentDAO struct {
	db *gorm.DB
}

func NewPrivacyEmailContentDAO(db *gorm.DB) *PrivacyEmailContentDAO {
	return &PrivacyEmailContentDAO{db: db}
}

func (obj *PrivacyEmailContentDAO) GetTableName() string {
	return TableNamePrivacyEmailContent
}

func (obj *PrivacyEmailContentDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *PrivacyEmailContentDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *PrivacyEmailContentDAO) GetByEmailID(ctx context.Context, emailID int64) ([]*PrivacyEmailContent, error) {
	var result []*PrivacyEmailContent
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("email_id = ?", emailID).
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *PrivacyEmailContentDAO) DeleteByEmailID(ctx context.Context, emailID int64) error {
	return obj.db.WithContext(ctx).Delete(&PrivacyEmailContent{}, "email_id=?", emailID).Error
}

func (obj *PrivacyEmailContentDAO) DeleteByMailbox(ctx context.Context, emailAccount, mailbox string) error {
	subQuery := obj.db.WithContext(ctx).
		Table(TableNamePrivacyEmail).
		Select("id").
		Where("email_account=? AND mailbox=?", emailAccount, mailbox)
	return obj.db.WithContext(ctx).Delete(&PrivacyEmailContent{}, "email_id IN (?)", subQuery).Error
}

func (obj *PrivacyEmailContentDAO) DeleteExpiredEmails(ctx context.Context, expiredAt int64) error {
	subQuery := obj.db.WithContext(ctx).
		Table(TableNamePrivacyEmail).
		Select("id").
		Where("sent_at<?", expiredAt)
	return obj.db.WithContext(ctx).Delete(&PrivacyEmailContent{}, "email_id IN (?)", subQuery).Error
}
