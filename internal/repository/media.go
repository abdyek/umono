package repository

import (
	"strings"

	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

type MediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) GetByID(id string) models.Media {
	var media models.Media
	r.db.Model(&models.Media{}).Where("id = ?", id).First(&media)
	return media
}

func (r *MediaRepository) GetAll() []models.Media {
	var all []models.Media
	r.db.Model(&models.Media{}).Order("created_at DESC").Find(&all)
	return all
}

func (r *MediaRepository) GetByHash(hash string) models.Media {
	var media models.Media
	r.db.Model(&models.Media{}).Where("hash = ?", hash).First(&media)
	return media
}

func (r *MediaRepository) GetByAlias(alias string) models.Media {
	var media models.Media
	r.db.Model(&models.Media{}).
		Where("lower(json_extract(metadata, '$.alias')) = lower(?)", strings.TrimSpace(alias)).
		First(&media)
	return media
}

func (r *MediaRepository) Create(media models.Media) models.Media {
	r.db.Create(&media)
	return media
}

func (r *MediaRepository) Update(media models.Media) models.Media {
	r.db.Model(&media).Select("*").Updates(media)
	return media
}

func (r *MediaRepository) Delete(id string) error {
	return r.db.Delete(&models.Media{}, "id = ?", id).Error
}
