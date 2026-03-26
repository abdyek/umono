package view

import (
	"html/template"

	"github.com/umono-cms/umono/internal/models"
)

type ComponentEditorData struct {
	ID             uint
	Name           string
	Content        string
	Output         template.HTML
	LastModifiedAt string
	NameErr        string
	I18n           any
}

func ComponentEditor(comp models.Component, output template.HTML, nameErr string, translator any) ComponentEditorData {
	return ComponentEditorData{
		ID:             comp.ID,
		Name:           comp.Name,
		Content:        comp.Content,
		Output:         output,
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
