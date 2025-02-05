package umono

import (
	ul "github.com/umono-cms/umono-lang"
	"github.com/umono-cms/umono-lang/converters"
)

var Lang *ul.UmonoLang

func InitLang() {
	Lang = ul.New(converters.NewHTML())
}
