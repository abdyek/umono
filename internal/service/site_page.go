package service

import (
	"bytes"
	"errors"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

// TODO: Remove interface
type SitePageService interface {
	GetRenderedBySlug(string) (models.SitePage, error)
	GetBySlug(string) (models.SitePage, error)
	GetByID(uint) (models.SitePage, error)
	GetAll() []models.SitePage
	Preview(string) (string, error)
	MustPreview(string) string
}

var ErrSitePageNotFound = errors.New("site page not found")

type sitePageService struct {
	repo    *repository.SitePageRepository
	compono compono.Compono
}

func NewSitePageService(r *repository.SitePageRepository, comp compono.Compono) SitePageService {
	return &sitePageService{
		repo:    r,
		compono: comp,
	}
}

func (s *sitePageService) GetRenderedBySlug(slug string) (models.SitePage, error) {
	// TODO: Put here cache
	sitePage, err := s.GetBySlug(slug)
	if err != nil {
		return models.SitePage{}, nil
	}

	output, err := s.convert(sitePage.Content)
	if err != nil {
		return models.SitePage{}, err
	}

	sitePage.Content = output
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

func (s *sitePageService) Preview(source string) (string, error) {
	output, err := s.convert(source)
	if err != nil {
		return "", err
	}
	return output, nil
}

func (s *sitePageService) MustPreview(source string) string {
	output, _ := s.Preview(source)
	return output
}

func (s *sitePageService) convert(source string) (string, error) {
	var buf bytes.Buffer
	if err := s.compono.Convert([]byte(source), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
