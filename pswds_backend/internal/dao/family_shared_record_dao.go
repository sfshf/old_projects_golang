package dao

import (
	"context"
	"encoding/json"
	"fmt"

	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/gorm"
)

type FamilySharedRecordDAO struct {
	db *gorm.DB
}

func NewFamilySharedRecordDAO(db *gorm.DB) *FamilySharedRecordDAO {
	return &FamilySharedRecordDAO{db: db}
}

func (obj *FamilySharedRecordDAO) GetTableName() string {
	return TableNameFamilySharedRecord
}

func (obj *FamilySharedRecordDAO) Table(ctx context.Context) *gorm.DB {
	return obj.db.WithContext(ctx).Table(obj.GetTableName())
}

const (
	FamilySharedRecordShared      = 1
	FamilySharedRecordSharedToAll = 2
)

func (obj *FamilySharedRecordDAO) DeleteByFamilyID(ctx context.Context, familyID string) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Delete(&FamilySharedRecord{}).Error
}

func (obj *FamilySharedRecordDAO) DeleteByUserIDAndDataID(ctx context.Context, userID int64, dataID string) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("shared_by = ?", userID).
		Where("data_id = ?", dataID).
		Delete(&FamilySharedRecord{}).Error
}

func (obj *FamilySharedRecordDAO) DeleteMemberByFamilyIDAndUserID(ctx context.Context, familyID string, userID int64) error {
	// 1. userID分享出去的数据
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Where("shared_by = ?", userID).
		Delete(&FamilySharedRecord{}).Error; err != nil {
		return err
	}
	// 2. 其他成员分享给userID的数据
	var result []*FamilySharedRecord
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where(`family_id = ? AND sharing_members LIKE ?`, familyID, fmt.Sprintf("%%%d%%", userID)).
		Find(&result).Error; err != nil {
		return err
	}
	for _, record := range result {
		members := make([]int64, 0)
		if err := json.Unmarshal([]byte(record.SharingMembers), &members); err != nil {
			return err
		}
		var newMembers []int64
		for _, member := range members {
			if member != userID {
				newMembers = append(newMembers, member)
			}
		}
		newSharingMembers, err := json.Marshal(newMembers)
		if err != nil {
			return err
		}
		if err := obj.db.WithContext(ctx).
			Table(obj.GetTableName()).
			Where("id = ?", record.ID).
			Updates(&FamilySharedRecord{
				SharingMembers: string(newSharingMembers),
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (obj *FamilySharedRecordDAO) GetByUserIDAndDataID(ctx context.Context, userID int64, dataID string) (*FamilySharedRecord, error) {
	var result FamilySharedRecord
	conn := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("shared_by = ?", userID).
		Where("data_id = ?", dataID).
		Limit(1).
		Find(&result)
	if conn.RowsAffected == 0 {
		return nil, conn.Error
	} else {
		return &result, nil
	}
}

func (obj *FamilySharedRecordDAO) Create(ctx context.Context, records interface{}) error {
	return obj.db.WithContext(ctx).Create(records).Error
}

func (obj *FamilySharedRecordDAO) UpdateByUserIDAndDataID(ctx context.Context, userID int64, dataID string, one *FamilySharedRecord) error {
	return obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("shared_by = ?", userID).
		Where("data_id = ?", dataID).
		Updates(one).Error
}

func (obj *FamilySharedRecordDAO) GetSharedDataChecksumByFamilyIDAndUserID(ctx context.Context, familyID string, userID int64) (int64, error) {
	var result struct {
		Checksum int64 `gorm:"column:checksum" json:"checksum"`
	}
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Select("SUM(updated_at) AS checksum").
		Where(`family_id = ? AND (sharing_members LIKE ? OR shared_by = ? OR shared_to_all = ?)`,
			familyID,
			fmt.Sprintf("%%%d%%", userID), // 其他成员分享给userID的数据
			userID,                        // userID分享出去的数据
			FamilySharedRecordSharedToAll, // 家庭全员（包括后加入者）共享的数据
		).
		Find(&result).Error; err != nil {
		return -1, err
	}
	return result.Checksum, nil
}

func (obj *FamilySharedRecordDAO) GetByFamilyIDAndUserID(ctx context.Context, familyID string, userID int64) ([]*FamilySharedRecord, error) {
	var result []*FamilySharedRecord
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where(`family_id = ? AND (sharing_members LIKE ? OR shared_by = ? OR shared_to_all = ?)`,
			familyID,
			fmt.Sprintf("%%%d%%", userID), // 其他成员分享给userID的数据
			userID,                        // userID分享出去的数据
			FamilySharedRecordSharedToAll, // 家庭全员（包括后加入者）共享的数据
		).
		Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (obj *FamilySharedRecordDAO) CountByFamilyID(ctx context.Context, familyID string) (int64, error) {
	var result int64
	if err := obj.db.WithContext(ctx).
		Table(obj.GetTableName()).
		Where("family_id = ?", familyID).
		Count(&result).Error; err != nil {
		return 0, err
	}
	return result, nil
}
