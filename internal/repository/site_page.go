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
	r.db.Model(&models.SitePage{}).Find(&all)
	return all
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
