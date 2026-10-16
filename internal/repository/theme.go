package repository

import (
	"encoding/json"

	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

// activeThemesKey is the option that holds the active themes.
const activeThemesKey = "active_themes"

type ThemeRepository struct {
	db *gorm.DB
}

func NewThemeRepository(db *gorm.DB) *ThemeRepository {
	return &ThemeRepository{db: db}
}

// Activate adds the contract record c when it is new (ID 0) and puts t into
// the active_themes option in place of the element of its group, or appends
// it, in one transaction. On success c has its ID.
func (r *ThemeRepository) Activate(c *models.Contract, t models.ThemeOption) error {
	ref, err := models.ParseContractRef(t.ContractID)
	if err != nil {
		return err
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if c.ID == 0 {
			if err := tx.Create(c).Error; err != nil {
				return err
			}
		}

		var rows []models.Option
		if err := tx.Where("key = ?", activeThemesKey).Limit(1).Find(&rows).Error; err != nil {
			return err
		}
		themes := []models.ThemeOption{}
		if len(rows) > 0 {
			if err := json.Unmarshal([]byte(rows[0].Value), &themes); err != nil {
				return err
			}
		}

		replaced := false
		for i, active := range themes {
			activeRef, err := models.ParseContractRef(active.ContractID)
			if err != nil {
				return err
			}
			if activeRef.Group() == ref.Group() {
				themes[i] = t
				replaced = true
				break
			}
		}
		if !replaced {
			themes = append(themes, t)
		}

		value, err := json.Marshal(themes)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return tx.Create(&models.Option{Key: activeThemesKey, Value: string(value)}).Error
		}
		return tx.Model(&models.Option{}).Where("id = ?", rows[0].ID).Update("value", string(value)).Error
	})
}
