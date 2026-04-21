package service

import (
	"errors"

	"github.com/umono-cms/umono/internal/i18n"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

const (
	DefaultLanguage           = "en"
	DefaultStorageIDOptionKey = "default_storage_id"
)

var ErrInvalidLanguage = errors.New("invalid language")

type OptionService struct {
	repo   *repository.OptionRepository
	bundle *i18n.Bundle
}

func NewOptionService(r *repository.OptionRepository, bundle *i18n.Bundle) *OptionService {
	return &OptionService{
		repo:   r,
		bundle: bundle,
	}
}

func (s *OptionService) SaveNotFoundPageOption(option models.NotFoundPageOption) error {
	return s.repo.SaveNotFoundPageOption(option)
}

func (s *OptionService) GetLanguage() string {
	option := s.repo.GetOptionByKey("language")
	if option.Value == "" {
		return DefaultLanguage
	}

	if !s.bundle.HasLocale(option.Value) {
		return DefaultLanguage
	}

	return option.Value
}

func (s *OptionService) SaveLanguage(language string) error {
	if !s.bundle.HasLocale(language) {
		return ErrInvalidLanguage
	}

	return s.repo.SaveOption("language", language)
}

func (s *OptionService) GetDefaultStorageID() string {
	option := s.repo.GetOptionByKey(DefaultStorageIDOptionKey)
	if option.Value == "" {
		return DefaultLocalStorageID
	}

	return option.Value
}

func (s *OptionService) SaveDefaultStorageID(storageID string) error {
	if storageID == "" {
		storageID = DefaultLocalStorageID
	}

	return s.repo.SaveOption(DefaultStorageIDOptionKey, storageID)
}

func (s *OptionService) SupportedLanguages() []i18n.LocaleOption {
	return s.bundle.SupportedLocales()
}

func (s *OptionService) GetNotFoundPageOption() (string, string) {
	var title, content string

	notFoundPageOpt, err := repository.GetOption[models.NotFoundPageOption](*s.repo, "not_found_page_option")
	if err == nil {
		title = notFoundPageOpt.Title
		content = notFoundPageOpt.Content
	}

	return title, content
}
