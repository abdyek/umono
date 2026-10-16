package view

import (
	"html/template"

	"github.com/umono-cms/umono/internal/models"
)

type ComponentEditorData struct {
	ID          uint
	Name        string
	Content     string
	Output      template.HTML
	Diagnostics []DiagnosticItem
	// PageErrorCount is the number of published pages with errors from this
	// global.
	PageErrorCount int
	LastModifiedAt string
	NameErr        string
	ContentErr     string
	I18n           any
}

func ComponentEditor(comp models.Component, output template.HTML, diagnostics []DiagnosticItem, nameErr string, translator any) ComponentEditorData {
	return ComponentEditorData{
		ID:             comp.ID,
		Name:           comp.Name,
		Content:        comp.Content,
		Output:         output,
		Diagnostics:    diagnostics,
		LastModifiedAt: RelativeTimeWithTranslator(comp.LastModifiedAt, translator),
		NameErr:        nameErr,
		I18n:           translator,
	}
}

type ComponentLi struct {
	ID       uint
	Name     string
	IsActive bool
}

func ComponentUl(comps []models.Component, activeID uint) []ComponentLi {
	ul := []ComponentLi{}
	for _, comp := range comps {
		ul = append(ul, ComponentLi{
			ID:       comp.ID,
			Name:     comp.Name,
			IsActive: comp.ID == activeID,
		})
	}
	return ul
}

// ComponentList is one page of the component sidebar. Next is the offset of
// the next page, or 0 when there is none.
type ComponentList struct {
	Items    []ComponentLi
	Next     int
	ActiveID uint
}
