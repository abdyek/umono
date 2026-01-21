package view

import "github.com/umono-cms/umono/internal/models"

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
