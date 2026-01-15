package service

import (
	"errors"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

type PageService interface {
	GetBySlug(string) (models.Page, error)
}

var ErrPageNotFound = errors.New("page not found")

type pageService struct {
	repo *repository.PageRepository
}

func NewPageService(r *repository.PageRepository) PageService {
	return &pageService{repo: r}
}

func (s *pageService) GetBySlug(slug string) (models.Page, error) {

	// TODO: Put here cache

	page := s.repo.GetBySlug(slug)
	if page.ID == 0 {
		return models.Page{}, ErrPageNotFound
	}
	return page, nil
}
