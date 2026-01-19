package service

import (
	"bytes"
	"errors"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

// TODO: Remove interface
type ComponentService interface {
	GetByID(uint) (models.Component, error)
	GetAll() []models.Component
	LoadAsGlobalComponent()
	Preview(string, string) (string, error)
	MustPreview(string, string) string
}

var ErrComponentNotFound = errors.New("component not found")

type componentService struct {
	repo    *repository.ComponentRepository
	compono compono.Compono
}

func NewComponentService(r *repository.ComponentRepository, comp compono.Compono) ComponentService {
	return &componentService{
		repo:    r,
		compono: comp,
	}
}

func (s *componentService) GetByID(ID uint) (models.Component, error) {
	c := s.repo.GetByID(ID)
	if c.ID == 0 {
		return models.Component{}, ErrComponentNotFound
	}
	return c, nil
}

func (s *componentService) GetAll() []models.Component {
	return s.repo.GetAll()
}

func (s *componentService) LoadAsGlobalComponent() {
	for _, comp := range s.GetAll() {
		s.compono.RegisterGlobalComponent(comp.Name, []byte(comp.Content))
	}
}

func (s *componentService) Preview(name, source string) (string, error) {
	var buf bytes.Buffer
	if err := s.compono.ConvertGlobalComponent(name, []byte(source), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *componentService) MustPreview(name, source string) string {
	output, _ := s.Preview(name, source)
	return output
}
