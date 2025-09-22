package dao

import (
	"github.com/nextsurfer/ground/pkg/dao"
	"gorm.io/gorm"
)

type Manager struct {
	DB *gorm.DB

	BackupDAO                *BackupDAO
	PasswordRecordDAO        *PasswordRecordDAO
	TrustedContactDAO        *TrustedContactDAO
	PrivacyEmailDAO          *PrivacyEmailDAO
	PrivacyEmailContentDAO   *PrivacyEmailContentDAO
	PrivacyEmailAccountDAO   *PrivacyEmailAccountDAO
	NonPasswordRecordDAO     *NonPasswordRecordDAO
	FamilyDAO                *FamilyDAO
	FamilyInvitationDAO      *FamilyInvitationDAO
	FamilyMemberDAO          *FamilyMemberDAO
	FamilySharedRecordDAO    *FamilySharedRecordDAO
	FamilyMessageDAO         *FamilyMessageDAO
	FamilyBackupDAO          *FamilyBackupDAO
	FamilyRecoverDAO         *FamilyRecoverDAO
	UnlockPasswordRecoverDAO *UnlockPasswordRecoverDAO
}

func NewManager(option *dao.OPtion) *Manager {
	return ManagerWithDB(option.DB)
}

func ManagerWithDB(db *gorm.DB) *Manager {
	return &Manager{
		DB: db,

		BackupDAO:                NewBackupDAO(db),
		PasswordRecordDAO:        NewPasswordRecordDAO(db),
		TrustedContactDAO:        NewTrustedContactDAO(db),
		PrivacyEmailDAO:          NewPrivacyEmailDAO(db),
		PrivacyEmailContentDAO:   NewPrivacyEmailContentDAO(db),
		PrivacyEmailAccountDAO:   NewPrivacyEmailAccountDAO(db),
		NonPasswordRecordDAO:     NewNonPasswordRecordDAO(db),
		FamilyDAO:                NewFamilyDAO(db),
		FamilyInvitationDAO:      NewFamilyInvitationDAO(db),
		FamilyMemberDAO:          NewFamilyMember(db),
		FamilySharedRecordDAO:    NewFamilySharedRecordDAO(db),
		FamilyMessageDAO:         NewFamilyMessageDAO(db),
		FamilyBackupDAO:          NewFamilyBackupDAO(db),
		FamilyRecoverDAO:         NewFamilyRecoverDAO(db),
		UnlockPasswordRecoverDAO: NewUnlockPasswordRecoverDAO(db),
	}
}

func (m *Manager) Transaction() (*gorm.DB, *Manager) {
	tx := m.DB.Begin()
	return tx, ManagerWithDB(tx)
}

func (m *Manager) TransFunc(fc func(tx *gorm.DB) error) error {
	return m.DB.Transaction(fc)
}
