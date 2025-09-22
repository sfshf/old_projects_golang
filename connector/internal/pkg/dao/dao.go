package dao

import (
	"github.com/nextsurfer/ground/pkg/dao"
	"gorm.io/gorm"
)

// Manager for dao service
type Manager struct {
	DB *gorm.DB

	ApiKeyDAO            *ApiKeyDAO
	AppConfigDAO         *AppConfigDAO
	RelationAppDatumDAO  *RelationAppDatumDAO
	RelationAppKeyDAO    *RelationAppKeyDAO
	ManagePlatformLogDAO *ManagePlatformLogDAO
}

// NewManager create dao manager which contains all dao service
func NewManager(option *dao.OPtion) *Manager {
	return NewManagerWithDB(option.DB)
}

func NewManagerWithDB(db *gorm.DB) *Manager {
	return &Manager{
		DB:                   db,
		ApiKeyDAO:            NewApiKeyDAO(db),
		AppConfigDAO:         NewAppConfigDAO(db),
		RelationAppDatumDAO:  NewRelationAppDatumDAO(db),
		RelationAppKeyDAO:    NewRelationAppKeyDAO(db),
		ManagePlatformLogDAO: NewManagePlatformLogDAO(db),
	}
}

// Transaction create transaction and manager using transaction db
func (m *Manager) Transaction() (*gorm.DB, *Manager) {
	tx := m.DB.Begin()
	newManager := NewManagerWithDB(tx)
	return tx, newManager
}
