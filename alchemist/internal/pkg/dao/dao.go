package dao

import (
	"github.com/nextsurfer/ground/pkg/dao"
	"gorm.io/gorm"
)

// Manager for dao service
type Manager struct {
	DB *gorm.DB

	SlarkUserDAO                 *SlarkUserDAO
	RawTransactionsDAO           *RawTransactionsDAO
	TransactionsProdDAO          *TransactionsProdDAO
	TransactionsTestDAO          *TransactionsTestDAO
	SubscriptionStateProdDAO     *SubscriptionStateProdDAO
	SubscriptionStateTestDAO     *SubscriptionStateTestDAO
	ReferralCodeDAO              *ReferralCodeDAO
	ReferralNewUserDAO           *ReferralNewUserDAO
	ReferralLogDAO               *ReferralLogDAO
	ReferralPointDAO             *ReferralPointDAO
	NewUserDiscountStateDAO      *NewUserDiscountStateDAO
	PromoOfferRecordsDAO         *PromoOfferRecordsDAO
	FreeTrialStateDAO            *FreeTrialStateDAO
	UserRegisteredOnOldDeviceDAO *UserRegisteredOnOldDeviceDAO
	AppConfigDAO                 *AppConfigDAO
	SubscriptionCountDAO         *SubscriptionCountDAO
}

// NewManager create dao manager which contains all dao service
func NewManager(option *dao.OPtion) *Manager {
	return NewManagerWithDB(option.DB)
}

func NewManagerWithDB(db *gorm.DB) *Manager {
	return &Manager{
		DB:                           db,
		SlarkUserDAO:                 NewSlarkUserDAO(db),
		RawTransactionsDAO:           NewRawTransactionsDAO(db),
		TransactionsProdDAO:          NewTransactionsProdDAO(db),
		TransactionsTestDAO:          NewTransactionsTestDAO(db),
		SubscriptionStateProdDAO:     NewSubscriptionStateProdDAO(db),
		SubscriptionStateTestDAO:     NewSubscriptionStateTestDAO(db),
		ReferralCodeDAO:              NewReferralCodeDAO(db),
		ReferralNewUserDAO:           NewReferralNewUserDAO(db),
		ReferralLogDAO:               NewReferralLogDAO(db),
		ReferralPointDAO:             NewReferralPointDAO(db),
		NewUserDiscountStateDAO:      NewNewUserDiscountStateDAO(db),
		PromoOfferRecordsDAO:         NewPromoOfferRecordsDAO(db),
		FreeTrialStateDAO:            NewFreeTrialStateDAO(db),
		UserRegisteredOnOldDeviceDAO: NewUserRegisteredOnOldDeviceDAO(db),
		AppConfigDAO:                 NewAppConfigDAO(db),
		SubscriptionCountDAO:         NewSubscriptionCountDAO(db),
	}
}

// Transaction create transaction and manager using transaction db
func (m *Manager) Transaction() (*gorm.DB, *Manager) {
	tx := m.DB.Begin()
	newManager := NewManagerWithDB(tx)
	return tx, newManager
}
