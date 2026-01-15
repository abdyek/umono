package repository

import (
	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

type PageRepository struct {
	db *gorm.DB
}

func NewPageRepository(db *gorm.DB) *PageRepository {
	return &PageRepository{
		db: db,
	}
}

func (r *PageRepository) GetBySlug(slug string) models.Page {
	var page models.Page
	r.db.Model(&models.Page{}).Where("slug = ?", slug).First(&page)
	return page
}
