package repository

import (
	"errors"
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

func (r *MediaRepository) GetByIDWithVariants(id string) models.Media {
	var media models.Media
	r.db.Model(&models.Media{}).Preload("Variants").Where("id = ?", id).Find(&media)
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

func (r *MediaRepository) GetVariantByTarget(mediaID string, width int, mimeType string, configVersion int) models.MediaVariant {
	var variant models.MediaVariant
	r.db.Model(&models.MediaVariant{}).
		Where(
			"media_id = ? AND mime_type = ? AND CAST(json_extract(metadata, '$.width') AS INTEGER) = ? AND CAST(json_extract(metadata, '$.config_version') AS INTEGER) = ?",
			mediaID,
			mimeType,
			width,
			configVersion,
		).
		Find(&variant)
	return variant
}

func (r *MediaRepository) GetVariantByPathKey(pathKey string) models.MediaVariant {
	var variant models.MediaVariant
	r.db.Model(&models.MediaVariant{}).Where("path_key = ?", pathKey).Find(&variant)
	return variant
}

func (r *MediaRepository) CreateVariant(variant models.MediaVariant) (models.MediaVariant, error) {
	err := r.db.Create(&variant).Error
	if err != nil {
		width := jsonMapInt(variant.Metadata, "width")
		configVersion := jsonMapInt(variant.Metadata, "config_version")
		existing := r.GetVariantByTarget(variant.MediaID, width, variant.MimeType, configVersion)
		if existing.ID != "" {
			return existing, nil
		}
	}
	return variant, err
}

func jsonMapInt(metadata models.JSONMap, key string) int {
	if metadata == nil {
		return 0
	}

	switch value := metadata[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func (r *MediaRepository) DeleteVariant(id string) error {
	err := r.db.Delete(&models.MediaVariant{}, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (r *MediaRepository) Delete(id string) error {
	return r.db.Delete(&models.Media{}, "id = ?", id).Error
}
