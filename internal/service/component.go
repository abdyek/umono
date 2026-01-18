package service

import (
	"errors"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

// TODO: Remove interface
type ComponentService interface {
	GetByID(uint) (models.Component, error)
	GetAll() []models.Component
}

var ErrComponentNotFound = errors.New("component not found")

type componentService struct {
	repo *repository.ComponentRepository
}

func NewComponentService(r *repository.ComponentRepository) ComponentService {
	return &componentService{repo: r}
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
