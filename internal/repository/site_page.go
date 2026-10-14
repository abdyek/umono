package repository

import (
	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

type SitePageRepository struct {
	db *gorm.DB
}

func NewSitePageRepository(db *gorm.DB) *SitePageRepository {
	return &SitePageRepository{
		db: db,
	}
}

func (r *SitePageRepository) GetBySlug(slug string) models.SitePage {
	var sp models.SitePage
	r.db.Model(&models.SitePage{}).Where("slug = ?", slug).First(&sp)
	return sp
}

func (r *SitePageRepository) GetByID(ID uint) models.SitePage {
	var sp models.SitePage
	r.db.Model(&models.SitePage{}).First(&sp, ID)
	return sp
}

func (r *SitePageRepository) GetAll() []models.SitePage {
	var all []models.SitePage
	r.db.Model(&models.SitePage{}).Order("enabled DESC, last_modified_at DESC").Find(&all)
	return all
}

func (r *SitePageRepository) GetPage(offset, limit int) []models.SitePage {
	var items []models.SitePage
	r.db.Model(&models.SitePage{}).Order("enabled DESC, last_modified_at DESC, id DESC").Offset(offset).Limit(limit).Find(&items)
	return items
}

func (r *SitePageRepository) GetEnabled() []models.SitePage {
	var items []models.SitePage
	r.db.Model(&models.SitePage{}).Where("enabled = ?", true).Order("slug ASC").Find(&items)
	return items
}

func (r *SitePageRepository) Create(sp models.SitePage) models.SitePage {
	r.db.Create(&sp)
	return sp
}

func (r *SitePageRepository) Update(sp models.SitePage) models.SitePage {
	r.db.Model(&sp).Select("*").Updates(sp)
	return sp
}

func (r *SitePageRepository) Delete(id uint) error {
	return r.db.Delete(&models.SitePage{}, id).Error
}
