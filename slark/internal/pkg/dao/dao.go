package dao

import (
	"github.com/nextsurfer/ground/pkg/dao"
	"gorm.io/gorm"
)

// Manager for dao service
type Manager struct {
	db *gorm.DB

	UserDAO       *UserDAO
	SessionDAO    *SessionDAO
	ThirdPartyDAO *ThirdPartyDAO
	DeviceDAO     *DeviceDAO
}

// NewManager create dao manager which contains all dao service
func NewManager(option *dao.OPtion) *Manager {
	return ManagerWithDB(option.DB)
}

func ManagerWithDB(db *gorm.DB) *Manager {
	userDAO := NewUserDAO(db)
	sessionDAO := NewSessionDAO(db)
	thirdPartyDAO := NewThirdPartyDAO(db)
	deviceDAO := NewDeviceDAO(db)
	return &Manager{
		db: db,

		UserDAO:       userDAO,
		SessionDAO:    sessionDAO,
		ThirdPartyDAO: thirdPartyDAO,
		DeviceDAO:     deviceDAO,
	}
}

// Transaction create transaction and manager using transaction db
func (m *Manager) Transaction() (*gorm.DB, *Manager) {
	tx := m.db.Begin()
	return tx, ManagerWithDB(tx)
}
