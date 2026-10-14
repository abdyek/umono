package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
	"github.com/umono-cms/umono/internal/view"
)

type ComponentHandler struct {
	sitePageService  *service.SitePageService
	componentService *service.ComponentService
}

func NewComponentHandler(cs *service.ComponentService, sps *service.SitePageService) *ComponentHandler {
	return &ComponentHandler{
		componentService: cs,
		sitePageService:  sps,
	}
}

func (h *ComponentHandler) Create(c *fiber.Ctx) error {
	comp := models.Component{
		Name:    strings.TrimSpace(c.FormValue("name")),
		Content: c.FormValue("content"),
	}

	created, errs := h.componentService.Create(comp)

	nameErr := ""
	// TODO: Refactor
	if len(errs) > 0 {
		if err := service.ErrInvalidComponentName; service.ErrorsIs(errs, err) {
			nameErr = translatedValidationError(c, err)
		} else if err := service.ErrComponentNameAlreadyExists; service.ErrorsIs(errs, err) {
			nameErr = translatedValidationError(c, err)
		}
		if err := service.ErrNameRequired; service.ErrorsIs(errs, err) {
			nameErr = translatedValidationError(c, err)
		}
	} else {
		c.Set("HX-Push-Url", "/admin/components/"+strconv.FormatUint(uint64(created.ID), 10))
		comp = created
	}

	output, diagnostics := componentPreview(c.UserContext(), h.componentService, comp.Name, comp.Content)
	return Render(c, "partials/htmx/component-editor", fiber.Map{
		"ComponentEditor": view.ComponentEditor(comp, output, diagnostics, nameErr, c.Locals("I18n")),
		"ComponentUl":     componentList(h.componentService, 0, comp.ID),
		"SitePageUl":      sitePageList(h.sitePageService, 0, 0),
	})
}

func (h *ComponentHandler) Update(c *fiber.Ctx) error {
	u64, _ := strconv.ParseUint(c.FormValue("id"), 10, 0)

	comp := models.Component{
		ID:      uint(u64),
		Name:    strings.TrimSpace(c.FormValue("name")),
		Content: c.FormValue("content"),
	}

	updated, errs := h.componentService.Update(comp)

	nameErr := ""
	// TODO: Refactor
	if len(errs) > 0 {
		if err := service.ErrInvalidComponentName; service.ErrorsIs(errs, err) {
			nameErr = translatedValidationError(c, err)
		} else if err := service.ErrComponentNameAlreadyExists; service.ErrorsIs(errs, err) {
			nameErr = translatedValidationError(c, err)
		}
		if err := service.ErrNameRequired; service.ErrorsIs(errs, err) {
			nameErr = translatedValidationError(c, err)
		}
	} else {
		comp = updated
	}

	output, diagnostics := componentPreview(c.UserContext(), h.componentService, comp.Name, comp.Content)
	return Render(c, "partials/htmx/component-editor", fiber.Map{
		"ComponentEditor": view.ComponentEditor(comp, output, diagnostics, nameErr, c.Locals("I18n")),
		"ComponentUl":     componentList(h.componentService, 0, comp.ID),
		"SitePageUl":      sitePageList(h.sitePageService, 0, 0),
	})
}

func (h *ComponentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.ErrBadRequest
	}

	if err := h.componentService.Delete(uint(id)); err != nil {
		return err
	}

	c.Set("HX-Redirect", "/admin/components/new")
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ComponentHandler) RenderComponentEditor(c *fiber.Ctx) error {
	return Render(c, "pages/admin", fiber.Map{
		"ComponentMode":   true,
		"ComponentEditor": view.ComponentEditor(models.Component{}, mustPreviewHTML(""), nil, "", c.Locals("I18n")),
		"SitePageUl":      sitePageList(h.sitePageService, 0, 0),
		"ComponentUl":     componentList(h.componentService, 0, 0),
	}, "layouts/admin")
}
