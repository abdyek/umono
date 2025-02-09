package umono

import (
	ul "github.com/umono-cms/umono-lang"
	"github.com/umono-cms/umono-lang/converters"
	"github.com/umono-cms/umono/models"
	"gorm.io/gorm"
)

var Lang *ul.UmonoLang

func InitLang() {
	Lang = ul.New(converters.NewHTML())
}

func SetGlobalComponents(db *gorm.DB) {
	var comps []models.Component
	db.Model(&models.Component{}).Find(&comps)

	for _, cmp := range comps {
		// NOTE: You can handle errors
		Lang.SetGlobalComponent(cmp.Name, cmp.Content)
	}
}
