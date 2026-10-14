package service

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"
	"sync"
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

var ErrSitePageNotFound = errors.New("site page not found")

type SitePageService struct {
	repo            *repository.SitePageRepository
	optRepo         *repository.OptionRepository
	contentCompiler *ContentCompiler
	diagnosticRepo  *repository.PageDiagnosticRepository
	// diagnosticMu serializes the recomputations of the stored diagnostics so
	// that the last one writes what the latest inputs render.
	diagnosticMu sync.Mutex
}

func NewSitePageService(r *repository.SitePageRepository, or *repository.OptionRepository, cc *ContentCompiler) *SitePageService {
	return &SitePageService{
		repo:            r,
		optRepo:         or,
		contentCompiler: cc,
	}
}

func (s *SitePageService) GetRenderedBySlug(ctx context.Context, slug string) (models.SitePage, error) {
	// TODO: Put here cache
	sitePage, err := s.GetBySlug(slug)
	if err != nil {
		return models.SitePage{}, err
	}

	if !sitePage.Enabled {
		return models.SitePage{}, ErrSitePageNotFound
	}

	output, err := s.convert(ctx, sitePage.Content)
	if err != nil {
		return models.SitePage{}, err
	}

	sitePage.Content = output
	return sitePage, nil
}

func (s *SitePageService) GetNotFoundPage(ctx context.Context, defaultTitle, defaultContent string) (models.SitePage, error) {
	title := defaultTitle
	content := defaultContent

	notFoundPageOpt, err := repository.GetOption[models.NotFoundPageOption](*s.optRepo, "not_found_page_option")
	if err == nil {
		if notFoundPageOpt.Title != "" {
			title = notFoundPageOpt.Title
		}
		if notFoundPageOpt.Content != "" {
			content = notFoundPageOpt.Content
		}
	}

	rendered, err := s.convert(ctx, content)
	if err != nil {
		return models.SitePage{}, err
	}

	return models.SitePage{
		Name:    title,
		Content: rendered,
	}, nil
}

func (s *SitePageService) GetBySlug(slug string) (models.SitePage, error) {
	sp := s.repo.GetBySlug(slug)
	if sp.ID == 0 {
		return models.SitePage{}, ErrSitePageNotFound
	}
	return sp, nil
}

func (s *SitePageService) GetByID(ID uint) (models.SitePage, error) {
	sp := s.repo.GetByID(ID)
	if sp.ID == 0 {
		return models.SitePage{}, ErrSitePageNotFound
	}
	return sp, nil
}

func (s *SitePageService) GetAll() []models.SitePage {
	return s.repo.GetAll()
}

// GetPage returns a page of the pages as the admin sidebar lists them and the offset of the next page, or 0 when there is none.
func (s *SitePageService) GetPage(offset int) ([]models.SitePage, int) {
	if offset < 0 {
		offset = 0
	}
	items := s.repo.GetPage(offset, listPageSize+1)
	return pageOf(items, offset)
}

func (s *SitePageService) Preview(ctx context.Context, source string) (string, []compono.Diagnostic, error) {
	return s.contentCompiler.CompileWithProviderContext(ctx, source)
}

func (s *SitePageService) MustPreview(ctx context.Context, source string) (string, []compono.Diagnostic) {
	output, diagnostics, _ := s.Preview(ctx, source)
	return output, diagnostics
}

func (s *SitePageService) Create(sp models.SitePage) (models.SitePage, []error) {
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

	created := s.repo.Create(sp)
	if err := s.RecomputeDiagnostics(context.Background(), created.ID); err != nil {
		log.Printf("recompute diagnostics of page %d: %v", created.ID, err)
	}
	return created, nil
}

func (s *SitePageService) Update(sp models.SitePage) (models.SitePage, []error) {
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

	updated := s.repo.Update(sp)
	if err := s.RecomputeDiagnostics(context.Background(), updated.ID); err != nil {
		log.Printf("recompute diagnostics of page %d: %v", updated.ID, err)
	}
	return updated, nil
}

func (s *SitePageService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	if s.diagnosticRepo != nil {
		return s.diagnosticRepo.DeleteByPage(id)
	}
	return nil
}

func (s *SitePageService) CheckSlug(slug string, exclude uint) error {
	if !s.isSlugValid(slug) {
		return ErrInvalidSlug
	}
	if s.isSlugUsed(slug, exclude) {
		return ErrSlugAlreadyExists
	}
	return nil
}

// SetPageDiagnosticRepository sets where the diagnostics of the pages' last renders are stored.
func (s *SitePageService) SetPageDiagnosticRepository(r *repository.PageDiagnosticRepository) {
	s.diagnosticRepo = r
}

// RecomputeDiagnostics renders the page with the id and replaces its stored diagnostics.
func (s *SitePageService) RecomputeDiagnostics(ctx context.Context, id uint) error {
	if s.diagnosticRepo == nil {
		return nil
	}
	s.diagnosticMu.Lock()
	defer s.diagnosticMu.Unlock()
	page := s.repo.GetByID(id)
	if page.ID == 0 {
		return nil
	}
	return s.recomputeDiagnostics(ctx, page)
}

// RecomputeAllDiagnostics renders every page and replaces its stored diagnostics.
func (s *SitePageService) RecomputeAllDiagnostics(ctx context.Context) error {
	if s.diagnosticRepo == nil {
		return nil
	}
	s.diagnosticMu.Lock()
	defer s.diagnosticMu.Unlock()
	for _, page := range s.repo.GetAll() {
		if err := s.recomputeDiagnostics(ctx, page); err != nil {
			return err
		}
	}
	return nil
}

// diagnosticsFilledKey is the option that records that FillDiagnosticsOnce has
// filled the stored diagnostics.
const diagnosticsFilledKey = "page_diagnostics_filled"

// FillDiagnosticsOnce recomputes the stored diagnostics of every page on the
// first launch that stores them, so that the errors of content saved before
// show up in the list. Later calls change nothing.
func (s *SitePageService) FillDiagnosticsOnce(ctx context.Context) error {
	if s.diagnosticRepo == nil || s.optRepo == nil {
		return nil
	}
	if s.optRepo.GetOptionByKey(diagnosticsFilledKey).ID != 0 {
		return nil
	}
	if err := s.RecomputeAllDiagnostics(ctx); err != nil {
		return err
	}
	return s.optRepo.SaveOption(diagnosticsFilledKey, "done")
}

// PublishedDiagnostics returns the stored diagnostics of the enabled pages.
func (s *SitePageService) PublishedDiagnostics() []models.PublishedDiagnostic {
	if s.diagnosticRepo == nil {
		return []models.PublishedDiagnostic{}
	}
	return s.diagnosticRepo.GetPublished()
}

// PublishedPageCountOfGlobal returns the number of enabled pages with a stored
// diagnostic whose source is the global name or below it.
func (s *SitePageService) PublishedPageCountOfGlobal(name string) int {
	if s.diagnosticRepo == nil {
		return 0
	}
	return int(s.diagnosticRepo.CountPublishedPagesBySource(name))
}

func (s *SitePageService) recomputeDiagnostics(ctx context.Context, page models.SitePage) error {
	_, diagnostics, err := s.contentCompiler.CompileWithProviderContext(ctx, page.Content)
	if err != nil {
		return err
	}
	items := make([]models.PageDiagnostic, 0, len(diagnostics))
	for _, d := range diagnostics {
		items = append(items, models.PageDiagnostic{
			Code:    string(d.Code),
			Message: d.Message,
			Source:  strings.Join(d.Source, " > "),
		})
	}
	return s.diagnosticRepo.Replace(page.ID, items)
}

func (s *SitePageService) convert(ctx context.Context, source string) (string, error) {
	output, _, err := s.contentCompiler.CompileWithProviderContext(ctx, source)
	return output, err
}

func (s *SitePageService) isSlugValid(slug string) bool {
	if slug == "" {
		return true
	}
	return regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(slug)
}

func (s *SitePageService) isSlugUsed(slug string, excluding uint) bool {
	used := s.repo.GetBySlug(slug)
	if used.ID == 0 || used.ID == excluding {
		return false
	}
	return true
}
