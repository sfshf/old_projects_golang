package dao

import (
	"github.com/nextsurfer/ground/pkg/dao"
	"gorm.io/gorm"
)

// Manager for dao service
type Manager struct {
	DB                      *gorm.DB
	ServiceDAO              *ServiceDAO
	ApplicationDAO          *ApplicationDAO
	AcmeResourceDAO         *AcmeResourceDAO
	GatewayNodeDAO          *GatewayNodeDAO
	ProtoStatisticDAO       *ProtoStatisticDAO
	ProtoStatisticHourlyDAO *ProtoStatisticHourlyDAO
	AlarmEmailDAO           *AlarmEmailDAO
	RateLimitRuleDAO        *RateLimitRuleDAO
	HostManageDAO           *HostManageDAO
	TimeoutStatisticDAO     *TimeoutStatisticDAO
}

// NewManager create dao manager which contains all dao service
func NewManager(option *dao.OPtion) *Manager {
	return NewManagerWithDB(option.DB)
}

func NewManagerWithDB(db *gorm.DB) *Manager {
	return &Manager{
		DB:                      db,
		ServiceDAO:              NewServiceDAO(db),
		ApplicationDAO:          NewApplicationDAO(db),
		AcmeResourceDAO:         NewAcmeResourceDAO(db),
		GatewayNodeDAO:          NewGatewayNodeDAO(db),
		ProtoStatisticDAO:       NewProtoStatisticDAO(db),
		ProtoStatisticHourlyDAO: NewProtoStatisticHourlyDAO(db),
		AlarmEmailDAO:           NewAlarmEmailDAO(db),
		RateLimitRuleDAO:        NewRateLimitRuleDAO(db),
		HostManageDAO:           NewHostManageDAO(db),
		TimeoutStatisticDAO:     NewTimeoutStatisticDAO(db),
	}
}

// Transaction create transaction and manager using transaction db
func (m *Manager) Transaction() (*gorm.DB, *Manager) {
	tx := m.DB.Begin()
	newManager := NewManagerWithDB(tx)
	return tx, newManager
}
