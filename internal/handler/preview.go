package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/service"
)

type previewHandler struct {
	sitePageService  service.SitePageService
	componentService service.ComponentService
}

func NewPreviewHandler(sps service.SitePageService, cs service.ComponentService) *previewHandler {
	return &previewHandler{
		sitePageService:  sps,
		componentService: cs,
	}
}

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
