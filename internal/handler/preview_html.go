package handler

import (
	"fmt"
	"html/template"

	"github.com/umono-cms/umono/internal/runtime"
)

func buildPreviewHTML(renderedHTML string) string {
	gridCSS, _ := runtime.GenerateGridCSS(renderedHTML)
	if gridCSS == "" {
		return renderedHTML
	}

	return fmt.Sprintf("<style>%s</style>%s", gridCSS, renderedHTML)
}

func mustPreviewHTML(renderedHTML string) template.HTML {
	return template.HTML(buildPreviewHTML(renderedHTML))
}
