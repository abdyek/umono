package service

import (
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

type OptionService struct {
	repo *repository.OptionRepository
}

func NewOptionService(r *repository.OptionRepository) *OptionService {
	return &OptionService{
		repo: r,
	}
}

func (s *OptionService) SaveNotFoundPageOption(option models.NotFoundPageOption) error {
	return s.repo.SaveNotFoundPageOption(option)
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
