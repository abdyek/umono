package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/service"
)

type PageHandler struct {
	sitePageService service.SitePageService
}

func NewPageHandler(sps service.SitePageService) *PageHandler {
	return &PageHandler{
		sitePageService: sps,
	}
}

func (h *PageHandler) RenderAdmin(c *fiber.Ctx) error {
	sitePages := h.sitePageService.GetAll()
	return c.Render("pages/admin", fiber.Map{
		"SitePageUl": BuildSitePageUl(sitePages, 0),
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

	return c.Render("pages/admin", fiber.Map{
		"SitePageUl": BuildSitePageUl(sitePages, sitePage.ID),
	}, "layouts/admin")
}

func (h *PageHandler) RenderLogin(c *fiber.Ctx) error {
	return c.Render("pages/login", fiber.Map{}, "layouts/admin")
}

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
