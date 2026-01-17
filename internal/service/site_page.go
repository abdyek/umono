package service

import (
	"errors"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

type SitePageService interface {
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

func (s *sitePageService) GetBySlug(slug string) (models.SitePage, error) {

	// TODO: Put here cache

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
