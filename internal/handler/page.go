package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/service"
)

type PageHandler struct {
	pageService service.PageService
}

func NewPageHandler(ps service.PageService) *PageHandler {
	return &PageHandler{
		pageService: ps,
	}
}

func (h *PageHandler) RenderAdmin(c *fiber.Ctx) error {
	return c.Render("pages/admin", fiber.Map{}, "layouts/admin")
}

func (h *PageHandler) RenderLogin(c *fiber.Ctx) error {
	return c.Render("pages/login", fiber.Map{}, "layouts/admin")
}

func (h *PageHandler) RenderPage(c *fiber.Ctx) error {
	page, err := h.pageService.GetBySlug(c.Params("slug"))
	if err != nil {
		// TODO: Here 404 page
		return fiber.ErrNotFound
	}
	return c.Render("layouts/page", fiber.Map{
		"Title":   page.Name,
		"Content": page.Content,
	})
}

func (h *PageHandler) GetJoy(c *fiber.Ctx) error {
	return c.Render("partials/joy", fiber.Map{})
}
