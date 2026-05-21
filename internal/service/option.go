package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/umono-cms/umono/internal/i18n"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

const (
	DefaultLanguage                         = "en"
	DefaultStorageIDOptionKey               = "default_storage_id"
	LocalStorageImageUploadLimitMBOptionKey = "local_storage_image_upload_limit_mb"
	DefaultLocalStorageImageUploadLimitMB   = 4
	MinLocalStorageImageUploadLimitMB       = 1
	MaxLocalStorageImageUploadLimitMB       = 100
	LocalStorageImageUploadBodyLimitMB      = MaxLocalStorageImageUploadLimitMB
	bytesInMegabyte                         = 1024 * 1024
)

var (
	ErrInvalidLanguage                     = errors.New("invalid language")
	ErrInvalidLocalStorageImageUploadLimit = errors.New("invalid local storage image upload limit")
)

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

func (s *OptionService) GetLocalStorageImageUploadLimitMB() int {
	if s == nil || s.repo == nil {
		return DefaultLocalStorageImageUploadLimitMB
	}

	option := s.repo.GetOptionByKey(LocalStorageImageUploadLimitMBOptionKey)
	limitMB, err := parseLocalStorageImageUploadLimitMB(option.Value)
	if err != nil {
		return DefaultLocalStorageImageUploadLimitMB
	}

	return limitMB
}

func (s *OptionService) GetLocalStorageImageUploadLimitBytes() int64 {
	return int64(s.GetLocalStorageImageUploadLimitMB()) * bytesInMegabyte
}

func (s *OptionService) SaveLocalStorageImageUploadLimitMB(limitMB int) error {
	if !validLocalStorageImageUploadLimitMB(limitMB) {
		return ErrInvalidLocalStorageImageUploadLimit
	}

	return s.repo.SaveOption(LocalStorageImageUploadLimitMBOptionKey, strconv.Itoa(limitMB))
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

func parseLocalStorageImageUploadLimitMB(value string) (int, error) {
	limitMB, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, err
	}
	if !validLocalStorageImageUploadLimitMB(limitMB) {
		return 0, ErrInvalidLocalStorageImageUploadLimit
	}
	return limitMB, nil
}

func validLocalStorageImageUploadLimitMB(limitMB int) bool {
	return limitMB >= MinLocalStorageImageUploadLimitMB && limitMB <= MaxLocalStorageImageUploadLimitMB
}
