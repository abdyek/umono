package storage

import (
	"github.com/umono-cms/umono/models"
	"github.com/umono-cms/umono/umono"
	"gorm.io/gorm"
)

var Page page

type page struct {
	slugMap map[string]models.Page
}

func InitPageStorage() {
	Page.slugMap = make(map[string]models.Page)
}

func (p page) LoadAll(db *gorm.DB) {

	var allPages []models.Page
	db.Model(&models.Page{}).Where("enabled = ?", true).Find(&allPages)

	for _, pg := range allPages {
		p.Load(pg)
	}
}

func (p page) Load(pg models.Page) {
	pg.Content = umono.Lang.Convert(pg.Content)
	p.slugMap[pg.Slug] = pg
}

func (p page) Remove(pg models.Page) {
	delete(p.slugMap, pg.Slug)
}

func (p page) GetPage(slug string) (models.Page, bool) {
	pg, ok := p.slugMap[slug]
	return pg, ok
}
