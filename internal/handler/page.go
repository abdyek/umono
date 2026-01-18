package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/service"
)

type PageHandler struct {
	sitePageService  service.SitePageService
	componentService service.ComponentService
}

func NewPageHandler(sps service.SitePageService, cs service.ComponentService) *PageHandler {
	return &PageHandler{
		sitePageService:  sps,
		componentService: cs,
	}
}

func (h *PageHandler) RenderAdmin(c *fiber.Ctx) error {
	return Render(c, "pages/admin", fiber.Map{
		"SitePageUl":  BuildSitePageUl(h.sitePageService.GetAll(), 0),
		"ComponentUl": BuildComponentUl(h.componentService.GetAll(), 0),
	}, "layouts/admin")
}

func (h *PageHandler) RenderAdminSitePage(c *fiber.Ctx) error {
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
		"SitePageEditor": BuildSitePageEditor(sitePage),
		"SitePageUl":     BuildSitePageUl(sitePages, sitePage.ID),
		"ComponentUl":    BuildComponentUl(h.componentService.GetAll(), 0),
	}, "layouts/admin")
}

func (h *PageHandler) RenderAdminSitePageEditor(c *fiber.Ctx) error {

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
		"SitePageEditor": BuildSitePageEditor(sitePage),
		"SitePageUl":     BuildSitePageUl(h.sitePageService.GetAll(), sitePage.ID),
		"ComponentUl":    BuildComponentUl(h.componentService.GetAll(), 0),
	})
}

func (h *PageHandler) RenderAdminComponent(c *fiber.Ctx) error {

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
		"ComponentEditor": BuildComponentEditor(comp),
		"ComponentUl":     BuildComponentUl(h.componentService.GetAll(), comp.ID),
		"SitePageUl":      BuildSitePageUl(h.sitePageService.GetAll(), 0),
	}, "layouts/admin")
}

func (h *PageHandler) RenderAdminComponentEditor(c *fiber.Ctx) error {

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
		"ComponentEditor": BuildComponentEditor(comp),
		"ComponentUl":     BuildComponentUl(h.componentService.GetAll(), comp.ID),
		"SitePageUl":      BuildSitePageUl(h.sitePageService.GetAll(), 0),
	})
}

func (h *PageHandler) RenderLogin(c *fiber.Ctx) error {
	return c.Render("pages/login", fiber.Map{}, "layouts/admin")
}

// TODO: Change name RenderPage -> RenderSitePage
func (h *PageHandler) RenderPage(c *fiber.Ctx) error {
	sitePage, err := h.sitePageService.GetBySlug(c.Params("slug"))
	if err != nil {
		// TODO: Here 404 page
		return fiber.ErrNotFound
	}
	return c.Render("layouts/page", fiber.Map{
		"Title":   sitePage.Name,
		"Content": sitePage.Content,
	})
}

func (h *PageHandler) GetJoy(c *fiber.Ctx) error {
	return c.Render("partials/joy", fiber.Map{})
}
