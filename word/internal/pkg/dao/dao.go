package dao

import (
	"github.com/nextsurfer/ground/pkg/dao"
	"gorm.io/gorm"
)

// Manager for dao service
type Manager struct {
	DB *gorm.DB

	BookDAO               *BookDAO
	StringDAO             *StringDAO
	DefinitionDAO         *DefinitionDAO
	ExampleDAO            *ExampleDAO
	RelatedDAO            *RelatedDAO
	FavoriteDefinitionDAO *FavoriteDefinitionDAO
	ProgressBackupDAO     *ProgressBackupDAO
	TranslationDAO        *TranslationDAO
}

// NewManager create dao manager which contains all dao service
func NewManager(option *dao.OPtion) *Manager {
	return NewManagerWithDB(option.DB)
}

func NewManagerWithDB(db *gorm.DB) *Manager {
	return &Manager{
		DB:                    db,
		BookDAO:               NewBookDAO(db),
		StringDAO:             NewStringDAO(db),
		DefinitionDAO:         NewDefinitionDAO(db),
		ExampleDAO:            NewExampleDAO(db),
		RelatedDAO:            NewRelatedDAO(db),
		FavoriteDefinitionDAO: NewFavoriteDefinitionDAO(db),
		ProgressBackupDAO:     NewProgressBackupDAO(db),
		TranslationDAO:        NewTranslationDAO(db),
	}
}

// Transaction create transaction and manager using transaction db
func (m *Manager) Transaction() (*gorm.DB, *Manager) {
	tx := m.DB.Begin()
	newManager := NewManagerWithDB(tx)
	return tx, newManager
}
