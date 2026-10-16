package migration

import (
	"encoding/json"
	"time"

	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

type notFoundPageOption struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

const notFoundPageOptionKey = "not_found_page_option"

// NotFoundPage moves the 404 page of v0.8, the not_found_page_option option,
// to the not_found system page and deletes the option. Without the option
// the system page is empty. It returns the ID of the created page, or 0 when
// the system page already exists and nothing changed.
func NotFoundPage(db *gorm.DB) (uint, error) {
	var id uint
	err := db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.SitePage{}).Where("system_key = ?", models.NotFoundSystemKey).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}

		var opts []models.Option
		if err := tx.Where("key = ?", notFoundPageOptionKey).Limit(1).Find(&opts).Error; err != nil {
			return err
		}

		opt := notFoundPageOption{}
		var found *models.Option
		if len(opts) > 0 {
			found = &opts[0]
			if err := json.Unmarshal([]byte(found.Value), &opt); err != nil {
				return err
			}
		}

		now := time.Now()
		key := models.NotFoundSystemKey
		page := models.SitePage{
			Name:           opt.Title,
			Slug:           models.NotFoundSlug,
			Content:        opt.Content,
			Enabled:        true,
			LastModifiedAt: &now,
			SystemKey:      &key,
		}
		if err := tx.Create(&page).Error; err != nil {
			return err
		}

		if found != nil {
			if err := tx.Delete(&models.Option{}, found.ID).Error; err != nil {
				return err
			}
		}

		id = page.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}
