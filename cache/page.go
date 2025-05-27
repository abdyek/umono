package cache

import (
	ul "github.com/umono-cms/umono-lang"
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

	p.load404Page()
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

func (p page) load404Page() {

	name := "Page Not Found"

	if g404 := umono.Lang.GetGlobalComponent("404"); g404 != nil {

		titleParam := ul.ParameterByName(g404.Parameters(), "title")

		if titleParam != nil && titleParam.Type() == "string" {
			val, ok := titleParam.Default().(string)
			if ok {
				name = val
			}
		}
	}

	p.slugMap["_404"] = models.Page{
		Name:    name,
		Content: umono.Lang.Convert("{{ 404 }}"),
	}
}
