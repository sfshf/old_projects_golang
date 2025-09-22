package dao

import (
	"context"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type PrivacyEmailDAO struct {
	db *gorm.DB
}

func NewPrivacyEmailDAO(db *gorm.DB) *PrivacyEmailDAO {
	return &PrivacyEmailDAO{db: db}
}

func (obj *PrivacyEmailDAO) GetTableName() string {
	return TableNamePrivacyEmail
}

func (obj *PrivacyEmailDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

func (obj *PrivacyEmailDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *PrivacyEmailDAO) GetByHeaders(ctx context.Context, emailAccount, mailbox string, uid uint32, sentAt int64, sentBy, subject string) (*PrivacyEmail, error) {
	var result PrivacyEmail
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("email_account=?", emailAccount).
		Where("mailbox=?", mailbox).
		Where("uid=?", uid).
		Where("sent_at=?", sentAt).
		Where("sent_by=?", sentBy).
		Where("subject=?", subject).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *PrivacyEmailDAO) GetByMailbox(ctx context.Context, emailAccount, mailbox string) ([]*PrivacyEmail, error) {
	var result []*PrivacyEmail
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("email_account = ?", emailAccount).
		Where("mailbox = ?", mailbox).
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *PrivacyEmailDAO) GetByMailboxAndUid(ctx context.Context, mailbox string, uid uint32) (*PrivacyEmail, error) {
	var result PrivacyEmail
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("mailbox=?", mailbox).
		Where("uid=?", uid).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *PrivacyEmailDAO) DeleteByMailbox(ctx context.Context, emailAccount, mailbox string) error {
	return obj.db.WithContext(ctx).Delete(&PrivacyEmail{}, "email_account=? AND mailbox=?", emailAccount, mailbox).Error
}

func (obj *PrivacyEmailDAO) GetByID(ctx context.Context, id int64) (*PrivacyEmail, error) {
	var result PrivacyEmail
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("id=?", id).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *PrivacyEmailDAO) DeleteByID(ctx context.Context, id int64) error {
	return obj.db.WithContext(ctx).Delete(&PrivacyEmail{}, "id=?", id).Error
}

func (obj *PrivacyEmailDAO) ReorderUids(ctx context.Context, emailAccount, mailbox string, deletedUid int64) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("email_account=?", emailAccount).
		Where("mailbox=?", mailbox).
		Where("uid>?", deletedUid).
		UpdateColumn("uid", gorm.Expr("uid-?", 1)).Error
}

func (obj *PrivacyEmailDAO) DeleteExpiredEmails(ctx context.Context, expiredAt int64) error {
	return obj.db.WithContext(ctx).Delete(&PrivacyEmail{}, "sent_at<?", expiredAt).Error
}
