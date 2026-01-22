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
}

func ComponentEditor(comp models.Component, output, nameErr string) ComponentEditorData {
	return ComponentEditorData{
		ID:             comp.ID,
		Name:           comp.Name,
		Content:        comp.Content,
		Output:         template.HTML(output),
		LastModifiedAt: RelativeTime(comp.LastModifiedAt),
		NameErr:        nameErr,
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
