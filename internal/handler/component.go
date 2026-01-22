package handler

import (
	"strconv"

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
		Name:    c.FormValue("name"),
		Content: c.FormValue("content"),
	}

	created, errs := h.componentService.Create(comp)

	nameErr := ""
	// TODO: Refactor
	if len(errs) > 0 {
		if err := service.ErrInvalidComponentName; service.ErrorsIs(errs, err) {
			nameErr = err.Error()
		} else if err := service.ErrComponentNameAlreadyExists; service.ErrorsIs(errs, err) {
			nameErr = err.Error()
		}
	} else {
		c.Set("HX-Push-Url", "/admin/components/"+strconv.FormatUint(uint64(created.ID), 10))
		comp = created
	}

	return Render(c, "partials/htmx/component-editor", fiber.Map{
		"ComponentEditor": view.ComponentEditor(comp, h.componentService.MustPreview(comp.Name, comp.Content), nameErr),
		"ComponentUl":     view.ComponentUl(h.componentService.GetAll(), comp.ID),
		"SitePageUl":      view.SitePageUl(h.sitePageService.GetAll(), 0),
	})
}

func (h *ComponentHandler) Update(c *fiber.Ctx) error {

	u64, _ := strconv.ParseUint(c.FormValue("id"), 10, 0)

	comp := models.Component{
		ID:      uint(u64),
		Name:    c.FormValue("name"),
		Content: c.FormValue("content"),
	}

	updated, errs := h.componentService.Update(comp)

	nameErr := ""
	// TODO: Refactor
	if len(errs) > 0 {
		if err := service.ErrInvalidComponentName; service.ErrorsIs(errs, err) {
			nameErr = err.Error()
		} else if err := service.ErrComponentNameAlreadyExists; service.ErrorsIs(errs, err) {
			nameErr = err.Error()
		}
	} else {
		comp = updated
	}

	return Render(c, "partials/htmx/component-editor", fiber.Map{
		"ComponentEditor": view.ComponentEditor(comp, h.componentService.MustPreview(comp.Name, comp.Content), nameErr),
		"ComponentUl":     view.ComponentUl(h.componentService.GetAll(), comp.ID),
		"SitePageUl":      view.SitePageUl(h.sitePageService.GetAll(), 0),
	})
}

func (h *ComponentHandler) Delete(c *fiber.Ctx) error {
	// TODO: Delete
	return nil
}

func (h *ComponentHandler) RenderComponentEditor(c *fiber.Ctx) error {
	return Render(c, "pages/admin", fiber.Map{
		"ComponentMode":   true,
		"ComponentEditor": view.ComponentEditor(models.Component{}, "", ""),
		"SitePageUl":      view.SitePageUl(h.sitePageService.GetAll(), 0),
		"ComponentUl":     view.ComponentUl(h.componentService.GetAll(), 0),
	}, "layouts/admin")
}
