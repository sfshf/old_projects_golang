package dao

import (
	"github.com/nextsurfer/ground/pkg/dao"
	"gorm.io/gorm"
)

type Manager struct {
	DB *gorm.DB

	SiteDAO      *SiteDAO
	SiteAdminDAO *SiteAdminDAO
	CategoryDAO  *CategoryDAO
	PostDAO      *PostDAO
	CommentDAO   *CommentDAO
	ThumbupDAO   *ThumbupDAO
}

func NewManager(option *dao.OPtion) *Manager {
	return ManagerWithDB(option.DB)
}

func ManagerWithDB(db *gorm.DB) *Manager {
	return &Manager{
		DB:           db,
		SiteDAO:      NewSiteDAO(db),
		SiteAdminDAO: NewSiteAdminDAO(db),
		CategoryDAO:  NewCategoryDAO(db),
		PostDAO:      NewPostDAO(db),
		CommentDAO:   NewCommentDAO(db),
		ThumbupDAO:   NewThumbupDAO(db),
	}
}

func (m *Manager) Transaction() (*gorm.DB, *Manager) {
	tx := m.DB.Begin()
	return tx, ManagerWithDB(tx)
}

func (m *Manager) TransFunc(fc func(tx *gorm.DB) error) error {
	return m.DB.Transaction(fc)
}
