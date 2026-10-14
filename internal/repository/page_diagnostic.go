package repository

import (
	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

type PageDiagnosticRepository struct {
	db *gorm.DB
}

func NewPageDiagnosticRepository(db *gorm.DB) *PageDiagnosticRepository {
	return &PageDiagnosticRepository{
		db: db,
	}
}

func (r *PageDiagnosticRepository) Replace(pageID uint, items []models.PageDiagnostic) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("page_id = ?", pageID).Delete(&models.PageDiagnostic{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].PageID = pageID
			items[i].ID = 0
		}
		return tx.Create(&items).Error
	})
}

func (r *PageDiagnosticRepository) DeleteByPage(pageID uint) error {
	return r.db.Where("page_id = ?", pageID).Delete(&models.PageDiagnostic{}).Error
}

func (r *PageDiagnosticRepository) GetPublished() []models.PublishedDiagnostic {
	var items []models.PublishedDiagnostic
	r.db.Table("page_diagnostics").
		Select("page_diagnostics.page_id, pages.name AS page_name, page_diagnostics.code, page_diagnostics.message, page_diagnostics.source").
		Joins("JOIN pages ON pages.id = page_diagnostics.page_id").
		Where("pages.enabled = ?", true).
		Order("pages.name ASC, pages.id ASC, page_diagnostics.id ASC").
		Scan(&items)
	if items == nil {
		items = []models.PublishedDiagnostic{}
	}
	return items
}
