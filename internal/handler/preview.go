package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
)

type previewHandler struct {
	sitePageService  *service.SitePageService
	componentService *service.ComponentService
}

func NewPreviewHandler(sps *service.SitePageService, cs *service.ComponentService) *previewHandler {
	return &previewHandler{
		sitePageService:  sps,
		componentService: cs,
	}
}

// TODO: Change names
// RenderSitePagePreview -> SitePageReview
// RenderComponentPreview -> ComponentPreview

func (h *previewHandler) RenderSitePagePreview(c *fiber.Ctx) error {
	var req struct {
		Content string `form:"content"`
	}

	if err := c.BodyParser(&req); err != nil {
		return err
	}

	output, err := h.sitePageService.Preview(req.Content)
	if err != nil {
		return err
	}

	return c.SendString(output)
}

func (h *previewHandler) RenderComponentPreview(c *fiber.Ctx) error {
	var req struct {
		Content string `form:"content"`
		Name    string `form:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return err
	}

	output, err := h.componentService.Preview(req.Name, req.Content)
	if err != nil {
		return err
	}

	return c.SendString(output)
}

func (h *previewHandler) NotFoundPagePreview(c *fiber.Ctx) error {
	content := c.FormValue("content")
	if content == "" {
		content = models.DefaultNotFoundContent
	}

	output, err := h.sitePageService.Preview(content)
	if err != nil {
		return err
	}

	return c.SendString(output)
}
