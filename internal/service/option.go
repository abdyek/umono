package service

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/umono-cms/umono/internal/i18n"
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
	RobotsTxtOptionKey                      = "robots_txt"
	MaxRobotsTxtBytes                       = 512000
	SiteURLOptionKey                        = "site_url"
)

var (
	ErrInvalidLanguage                     = errors.New("invalid language")
	ErrInvalidLocalStorageImageUploadLimit = errors.New("invalid local storage image upload limit")
	ErrRobotsTxtTooLarge                   = errors.New("robots.txt content too large")
	ErrInvalidSiteURL                      = errors.New("invalid site url")
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

func (s *OptionService) GetRobotsTxt() string {
	return s.repo.GetOptionByKey(RobotsTxtOptionKey).Value
}

func (s *OptionService) SaveRobotsTxt(content string) error {
	content = strings.TrimPrefix(content, "\uFEFF")
	if len(content) > MaxRobotsTxtBytes {
		return ErrRobotsTxtTooLarge
	}

	return s.repo.SaveOption(RobotsTxtOptionKey, content)
}

// GetSiteURL returns the root address of the site, or "" when it is not set.
func (s *OptionService) GetSiteURL() string {
	return s.repo.GetOptionByKey(SiteURLOptionKey).Value
}

// SaveSiteURL saves the root address of the site. An empty value clears it.
func (s *OptionService) SaveSiteURL(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return s.repo.SaveOption(SiteURLOptionKey, "")
	}

	normalized, err := NormalizeSiteURL(value)
	if err != nil {
		return err
	}

	return s.repo.SaveOption(SiteURLOptionKey, normalized)
}

// NormalizeSiteURL returns the root address of the site: the scheme and the
// host of value, without a trailing slash. Only http and https are accepted.
// A path, a query, a fragment or user info is rejected with ErrInvalidSiteURL.
func NormalizeSiteURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if strings.ContainsAny(value, "?#") {
		return "", ErrInvalidSiteURL
	}

	u, err := url.Parse(value)
	if err != nil {
		return "", ErrInvalidSiteURL
	}

	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.Opaque != "" || (u.Path != "" && u.Path != "/") {
		return "", ErrInvalidSiteURL
	}

	return u.Scheme + "://" + u.Host, nil
}

func (s *OptionService) SupportedLanguages() []i18n.LocaleOption {
	return s.bundle.SupportedLocales()
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
