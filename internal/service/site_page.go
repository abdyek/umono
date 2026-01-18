package service

import (
	"errors"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

// TODO: Remove interface
type SitePageService interface {
	GetRenderedBySlug(string) (models.SitePage, error)
	GetBySlug(string) (models.SitePage, error)
	GetByID(uint) (models.SitePage, error)
	GetAll() []models.SitePage
}

var ErrSitePageNotFound = errors.New("site page not found")

type sitePageService struct {
	repo *repository.SitePageRepository
}

func NewSitePageService(r *repository.SitePageRepository) SitePageService {
	return &sitePageService{repo: r}
}

func (s *sitePageService) GetRenderedBySlug(slug string) (models.SitePage, error) {
	// TODO: Put here cache
	sitePage, err := s.GetBySlug(slug)
	if err != nil {
		return models.SitePage{}, nil
	}
	// TODO: Add Compono
	sitePage.Content = "Here will be Compono rendered output"
	return sitePage, nil
}

func (s *sitePageService) GetBySlug(slug string) (models.SitePage, error) {
	sp := s.repo.GetBySlug(slug)
	if sp.ID == 0 {
		return models.SitePage{}, ErrSitePageNotFound
	}
	return sp, nil
}

func (s *sitePageService) GetByID(ID uint) (models.SitePage, error) {
	sp := s.repo.GetByID(ID)
	if sp.ID == 0 {
		return models.SitePage{}, ErrSitePageNotFound
	}
	return sp, nil
}

func (s *sitePageService) GetAll() []models.SitePage {
	return s.repo.GetAll()
}
