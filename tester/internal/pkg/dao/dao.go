package dao

import (
	"github.com/nextsurfer/ground/pkg/dao"
	"gorm.io/gorm"
)

// Manager for dao service
type Manager struct {
	DB *gorm.DB
}

// NewManager create dao manager which contains all dao service
func NewManager(option *dao.OPtion) *Manager {
	return NewManagerWithDB(option.DB)
}

func NewManagerWithDB(db *gorm.DB) *Manager {
	return &Manager{
		DB: db,
	}
}

// Transaction create transaction and manager using transaction db
func (m *Manager) Transaction() (*gorm.DB, *Manager) {
	tx := m.DB.Begin()
	newManager := NewManagerWithDB(tx)
	return tx, newManager
}

type PswdsManager struct {
	DB                     *gorm.DB
	PrivacyEmailAccountDAO *PrivacyEmailAccountDAO
}

func NewPswdsManagerWithDB(db *gorm.DB) *PswdsManager {
	return &PswdsManager{
		DB:                     db,
		PrivacyEmailAccountDAO: NewPrivacyEmailAccountDAO(db),
	}
}
