package handler

import (
	"html/template"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/service"
)

type siteHandler struct {
	sitePageService *service.SitePageService
}

func NewSiteHandler(sps *service.SitePageService) *siteHandler {
	return &siteHandler{
		sitePageService: sps,
	}
}

func (h *siteHandler) RenderSitePage(c *fiber.Ctx) error {
	sitePage, err := h.sitePageService.GetRenderedBySlug(c.Params("slug"))
	if err != nil {
		// TODO: Here 404 page
		return fiber.ErrNotFound
	}
	return Render(c, "layouts/page", fiber.Map{
		"Title":   sitePage.Name,
		"Content": template.HTML(sitePage.Content),
	})
}
