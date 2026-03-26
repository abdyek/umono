package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
)

type OptionHandler struct {
	optionService *service.OptionService
}

func NewOptionHandler(os *service.OptionService) *OptionHandler {
	return &OptionHandler{
		optionService: os,
	}
}

func (h *OptionHandler) SaveNotFoundPageOption(c *fiber.Ctx) error {
	opt := models.NotFoundPageOption{
		Title:   c.FormValue("title"),
		Content: c.FormValue("content"),
	}

	err := h.optionService.SaveNotFoundPageOption(opt)
	if err != nil {
		return err
	}

	c.Set("HX-Trigger", "notFoundPageOptionSaved")
	return c.SendStatus(fiber.StatusOK)
}

func (h *OptionHandler) SaveLanguage(c *fiber.Ctx) error {
	if err := h.optionService.SaveLanguage(c.FormValue("language")); err != nil {
		if errors.Is(err, service.ErrInvalidLanguage) {
			return c.Status(fiber.StatusBadRequest).SendString("invalid language")
		}
		return err
	}

	c.Set("HX-Redirect", "/admin/settings/general")
	return c.SendStatus(fiber.StatusOK)
}
