package handler

import (
	"github.com/gofiber/fiber/v2"
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

	output, diagnostics := sitePagePreview(c.UserContext(), h.sitePageService, req.Content)

	return Render(c, "partials/editor-preview", fiber.Map{
		"Output":      output,
		"Diagnostics": diagnostics,
	})
}

func (h *previewHandler) RenderComponentPreview(c *fiber.Ctx) error {
	var req struct {
		Content string `form:"content"`
		Name    string `form:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return err
	}

	output, diagnostics := componentPreview(c.UserContext(), h.componentService, req.Name, req.Content)

	return Render(c, "partials/editor-preview", fiber.Map{
		"Output":      output,
		"Diagnostics": diagnostics,
	})
}

func (h *previewHandler) NotFoundPagePreview(c *fiber.Ctx) error {
	content := c.FormValue("content")
	if content == "" {
		_, content = localizedNotFoundDefaults(c)
	}

	output, diagnostics := sitePagePreview(c.UserContext(), h.sitePageService, content)

	return Render(c, "partials/editor-preview", fiber.Map{
		"Output":      output,
		"Diagnostics": diagnostics,
	})
}
