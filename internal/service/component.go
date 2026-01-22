package service

import (
	"bytes"
	"errors"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

var ErrComponentNotFound = errors.New("component not found")

type ComponentService struct {
	repo    *repository.ComponentRepository
	compono compono.Compono
}

func NewComponentService(r *repository.ComponentRepository, comp compono.Compono) *ComponentService {
	return &ComponentService{
		repo:    r,
		compono: comp,
	}
}

func (s *ComponentService) GetByID(ID uint) (models.Component, error) {
	c := s.repo.GetByID(ID)
	if c.ID == 0 {
		return models.Component{}, ErrComponentNotFound
	}
	return c, nil
}

func (s *ComponentService) GetAll() []models.Component {
	return s.repo.GetAll()
}

func (s *ComponentService) LoadAsGlobalComponent() {
	for _, comp := range s.GetAll() {
		s.compono.RegisterGlobalComponent(comp.Name, []byte(comp.Content))
	}
}

func (s *ComponentService) Preview(name, source string) (string, error) {
	var buf bytes.Buffer
	if err := s.compono.ConvertGlobalComponent(name, []byte(source), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *ComponentService) MustPreview(name, source string) string {
	output, _ := s.Preview(name, source)
	return output
}
