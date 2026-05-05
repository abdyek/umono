package service

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

var ErrComponentNotFound = errors.New("component not found")

var (
	ErrInvalidComponentName       = errors.New("Invalid name")
	ErrComponentNameAlreadyExists = errors.New("Name already exists for another component")
)

type ComponentService struct {
	repo            *repository.ComponentRepository
	contentCompiler *ContentCompiler
}

func NewComponentService(r *repository.ComponentRepository) *ComponentService {
	return &ComponentService{
		repo: r,
	}
}

func (s *ComponentService) SetContentCompiler(cc *ContentCompiler) {
	s.contentCompiler = cc
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

// TODO: Refactor LoadAsGlobalComponent -> LoadAllGlobalComponents()
func (s *ComponentService) LoadAsGlobalComponent() {
	if s.contentCompiler == nil {
		return
	}

	for _, comp := range s.GetAll() {
		_ = s.LoadGlobalComponent(comp)
	}
}

func (s *ComponentService) LoadGlobalComponent(comp models.Component) error {
	if s.contentCompiler == nil {
		return ErrContentCompilerNotConfigured
	}

	return s.contentCompiler.LoadGlobalComponent(comp)
}

func (s *ComponentService) RemoveGlobalComponent(comp models.Component) error {
	if s.contentCompiler == nil {
		return ErrContentCompilerNotConfigured
	}

	return s.contentCompiler.RemoveGlobalComponent(comp)
}

func (s *ComponentService) ReloadGlobalComponent(comp models.Component) error {
	if s.contentCompiler == nil {
		return ErrContentCompilerNotConfigured
	}

	return s.contentCompiler.ReloadGlobalComponent(comp)
}

func (s *ComponentService) Preview(ctx context.Context, name, source string) (string, error) {
	if s.contentCompiler == nil {
		return "", ErrContentCompilerNotConfigured
	}

	return s.contentCompiler.PreviewComponentWithProviderContext(ctx, name, source)
}

func (s *ComponentService) MustPreview(ctx context.Context, name, source string) string {
	output, _ := s.Preview(ctx, name, source)
	return output
}

func (s *ComponentService) Create(c models.Component) (models.Component, []error) {
	errs := []error{}

	if !s.isNameValid(c.Name) {
		errs = append(errs, ErrInvalidComponentName)
	}

	if s.isNameUsed(c.Name, 0) {
		errs = append(errs, ErrComponentNameAlreadyExists)
	}

	if c.Name == "" {
		errs = append(errs, ErrNameRequired)
	}

	if len(errs) > 0 {
		return models.Component{}, errs
	}

	now := time.Now()
	c.LastModifiedAt = &now

	created := s.repo.Create(c)
	err := s.LoadGlobalComponent(created)
	if err != nil {
		return models.Component{}, []error{err}
	}

	return created, nil
}

func (s *ComponentService) Update(c models.Component) (models.Component, []error) {
	errs := []error{}

	if !s.isNameValid(c.Name) {
		errs = append(errs, ErrInvalidComponentName)
	}

	if s.isNameUsed(c.Name, c.ID) {
		errs = append(errs, ErrComponentNameAlreadyExists)
	}

	if c.Name == "" {
		errs = append(errs, ErrNameRequired)
	}

	if len(errs) > 0 {
		return models.Component{}, errs
	}

	now := time.Now()
	c.LastModifiedAt = &now

	old := s.repo.GetByID(c.ID)
	updated := s.repo.Update(c)

	if old.Name != updated.Name {
		err := s.RemoveGlobalComponent(old)
		if err != nil {
			return models.Component{}, []error{err}
		}
		err = s.LoadGlobalComponent(updated)
		if err != nil {
			return models.Component{}, []error{err}
		}
		return updated, nil
	}

	err := s.ReloadGlobalComponent(updated)
	if err != nil {
		return models.Component{}, []error{err}
	}

	return updated, nil
}

func (s *ComponentService) Delete(id uint) error {
	comp := s.repo.GetByID(id)

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	return s.RemoveGlobalComponent(comp)
}

func (s *ComponentService) isNameValid(name string) bool {
	re := regexp.MustCompile(`^[A-Z0-9]+(?:_[A-Z0-9]+)*$`)
	return re.MatchString(name)
}

func (s *ComponentService) isNameUsed(name string, excluding uint) bool {
	used := s.repo.GetByName(name)
	if used.ID == 0 || used.ID == excluding {
		return false
	}
	return true
}
