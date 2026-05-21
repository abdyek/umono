package handler

import (
	"errors"
	"strconv"
	"strings"

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

func (h *OptionHandler) SaveLocalStorageImageUploadLimit(c *fiber.Ctx) error {
	value := strings.TrimSpace(c.FormValue("limit_mb"))
	if value == "" {
		value = strings.TrimSpace(c.FormValue("value"))
	}

	limitMB, err := strconv.Atoi(value)
	if err != nil {
		return h.renderLocalStorageImageUploadLimitStatus(c, false)
	}

	if err := h.optionService.SaveLocalStorageImageUploadLimitMB(limitMB); err != nil {
		if errors.Is(err, service.ErrInvalidLocalStorageImageUploadLimit) {
			return h.renderLocalStorageImageUploadLimitStatus(c, false)
		}
		return err
	}

	c.Set("HX-Trigger", "localStorageImageUploadLimitSaved")
	return h.renderLocalStorageImageUploadLimitStatus(c, true)
}

func (h *OptionHandler) renderLocalStorageImageUploadLimitStatus(c *fiber.Ctx, success bool) error {
	titleKey := "settings.storage.local_upload_limit.saved"
	messageKey := "settings.storage.local_upload_limit.saved_description"
	if !success {
		titleKey = "settings.storage.local_upload_limit.invalid"
		messageKey = "settings.storage.local_upload_limit.invalid_description"
	}

	return Render(c, "partials/local-storage-upload-limit-status", fiber.Map{
		"StatusSuccess": success,
		"StatusTitle":   translate(c, titleKey),
		"StatusMessage": translate(c, messageKey),
	})
}
