package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
)

type adminHandler struct {
	sitePageService  service.SitePageService
	componentService service.ComponentService
}

func NewAdminHandler(sps service.SitePageService, cs service.ComponentService) *adminHandler {
	return &adminHandler{
		sitePageService:  sps,
		componentService: cs,
	}
}

func (h *adminHandler) RenderAdmin(c *fiber.Ctx) error {
	return Render(c, "pages/admin", fiber.Map{
		"SitePageUl":  h.buildSitePageUl(h.sitePageService.GetAll(), 0),
		"ComponentUl": h.buildComponentUl(h.componentService.GetAll(), 0),
	}, "layouts/admin")
}

func (h *adminHandler) RenderAdminSitePage(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.SendStatus(400)
	}

	sitePage, err := h.sitePageService.GetByID(uint(id))
	if err != nil {
		return c.SendStatus(404)
	}

	sitePages := h.sitePageService.GetAll()

	return Render(c, "pages/admin", fiber.Map{
		"SitePageEditor": h.buildSitePageEditor(sitePage),
		"SitePageUl":     h.buildSitePageUl(sitePages, sitePage.ID),
		"ComponentUl":    h.buildComponentUl(h.componentService.GetAll(), 0),
	}, "layouts/admin")
}

func (h *adminHandler) RenderAdminSitePageEditor(c *fiber.Ctx) error {

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.SendStatus(400)
	}

	sitePage, err := h.sitePageService.GetByID(uint(id))
	if err != nil {
		return c.SendStatus(404)
	}

	return Render(c, "partials/htmx/admin-site-page-editor", fiber.Map{
		"SitePageEditor": h.buildSitePageEditor(sitePage),
		"SitePageUl":     h.buildSitePageUl(h.sitePageService.GetAll(), sitePage.ID),
		"ComponentUl":    h.buildComponentUl(h.componentService.GetAll(), 0),
	})
}

func (h *adminHandler) RenderAdminComponent(c *fiber.Ctx) error {

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.SendStatus(400)
	}

	comp, err := h.componentService.GetByID(uint(id))
	if err != nil {
		return c.SendStatus(404)
	}

	return Render(c, "pages/admin", fiber.Map{
		"ComponentMode":   true,
		"ComponentEditor": h.buildComponentEditor(comp),
		"ComponentUl":     h.buildComponentUl(h.componentService.GetAll(), comp.ID),
		"SitePageUl":      h.buildSitePageUl(h.sitePageService.GetAll(), 0),
	}, "layouts/admin")
}

func (h *adminHandler) RenderAdminComponentEditor(c *fiber.Ctx) error {

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.SendStatus(400)
	}

	comp, err := h.componentService.GetByID(uint(id))
	if err != nil {
		return c.SendStatus(404)
	}

	return Render(c, "partials/htmx/admin-component-editor", fiber.Map{
		"ComponentEditor": h.buildComponentEditor(comp),
		"ComponentUl":     h.buildComponentUl(h.componentService.GetAll(), comp.ID),
		"SitePageUl":      h.buildSitePageUl(h.sitePageService.GetAll(), 0),
	})
}

type sitePageLi struct {
	ID    string
	Title string
	/*PageURL   string
	HxGet     string
	HxTarget  string
	HxPushURL string*/
	IsActive  bool
	IsEnabled bool
}

func (h *adminHandler) buildSitePageUl(pages []models.SitePage, activeID uint) []sitePageLi {
	var ul []sitePageLi
	for _, sp := range pages {
		idStr := strconv.FormatUint(uint64(sp.ID), 10)
		ul = append(ul, sitePageLi{
			ID:        idStr,
			Title:     sp.Name,
			IsActive:  sp.ID == activeID,
			IsEnabled: sp.Enabled,
		})
	}
	return ul
}

type sitePageEditor struct {
	Name           string
	Slug           string
	Content        string
	Output         string
	IsEnabled      bool
	LastModifiedAt string
}

func (h *adminHandler) buildSitePageEditor(page models.SitePage) sitePageEditor {
	return sitePageEditor{
		Name:           page.Name,
		Slug:           page.Slug,
		Content:        page.Content,
		Output:         "here is output to preview",
		IsEnabled:      page.Enabled,
		LastModifiedAt: "2 hours ago", // TODO: get relative time string from page.LastModifiedAt
	}
}

type componentLi struct {
	ID       string
	Name     string
	IsActive bool
}

func (h *adminHandler) buildComponentUl(comps []models.Component, activeID uint) []componentLi {
	var ul []componentLi
	for _, c := range comps {
		idStr := strconv.FormatUint(uint64(c.ID), 10)
		ul = append(ul, componentLi{
			ID:       idStr,
			Name:     c.Name,
			IsActive: c.ID == activeID,
		})
	}
	return ul
}

type componentEditor struct {
	Name           string
	Content        string
	Output         string
	LastModifiedAt string
}

func (h *adminHandler) buildComponentEditor(comp models.Component) componentEditor {
	return componentEditor{
		Name:           comp.Name,
		Content:        comp.Content,
		Output:         "here is output to preview",
		LastModifiedAt: "2 hours ago", // TODO: get relative time string from component.LastModifiedAt
	}
}
