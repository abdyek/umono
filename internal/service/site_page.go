package service

import (
	"bytes"
	"errors"
	"regexp"
	"time"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

var (
	ErrSlugAlreadyExists = errors.New("Slug already exists for another page")
	ErrInvalidSlug       = errors.New("Invalid slug")
	ErrNameRequired      = errors.New("Name required")
)

// TODO: Remove interface
type SitePageService interface {
	GetRenderedBySlug(string) (models.SitePage, error)
	GetBySlug(string) (models.SitePage, error)
	GetByID(uint) (models.SitePage, error)
	GetAll() []models.SitePage
	Preview(string) (string, error)
	MustPreview(string) string
	Create(models.SitePage) (models.SitePage, []error)
	Update(models.SitePage) (models.SitePage, []error)
	CheckSlug(slug string, exclude uint) error
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

func (s *sitePageService) Create(sp models.SitePage) (models.SitePage, []error) {
	errs := []error{}

	if !s.isSlugValid(sp.Slug) {
		errs = append(errs, ErrInvalidSlug)
	}

	if s.isSlugUsed(sp.Slug, 0) {
		errs = append(errs, ErrSlugAlreadyExists)
	}

	if sp.Name == "" {
		errs = append(errs, ErrNameRequired)
	}

	if len(errs) > 0 {
		return models.SitePage{}, errs
	}

	now := time.Now()
	sp.LastModifiedAt = &now

	return s.repo.Create(sp), nil
}

func (s *sitePageService) Update(sp models.SitePage) (models.SitePage, []error) {
	errs := []error{}

	if !s.isSlugValid(sp.Slug) {
		errs = append(errs, ErrInvalidSlug)
	}

	if s.isSlugUsed(sp.Slug, sp.ID) {
		errs = append(errs, ErrSlugAlreadyExists)
	}

	if sp.Name == "" {
		errs = append(errs, ErrNameRequired)
	}

	if len(errs) > 0 {
		return models.SitePage{}, errs
	}

	now := time.Now()
	sp.LastModifiedAt = &now

	return s.repo.Update(sp), nil
}

func (s *sitePageService) CheckSlug(slug string, exclude uint) error {
	if !s.isSlugValid(slug) {
		return ErrInvalidSlug
	}
	if s.isSlugUsed(slug, exclude) {
		return ErrSlugAlreadyExists
	}
	return nil
}

func (s *sitePageService) convert(source string) (string, error) {
	var buf bytes.Buffer
	if err := s.compono.Convert([]byte(source), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *sitePageService) isSlugValid(slug string) bool {
	if slug == "" {
		return true
	}
	return regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(slug)
}

func (s *sitePageService) isSlugUsed(slug string, excluding uint) bool {
	used := s.repo.GetBySlug(slug)
	if used.ID == 0 || used.ID == excluding {
		return false
	}
	return true
}
