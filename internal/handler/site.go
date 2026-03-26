package handler

import (
	"errors"
	"html/template"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/runtime"
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
		if errors.Is(err, service.ErrSitePageNotFound) {
			defaultTitle, defaultContent := localizedNotFoundDefaults(c)
			sitePage, err = h.sitePageService.GetNotFoundPage(defaultTitle, defaultContent)
			if err != nil {
				// TODO: handle other errors
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			c.Status(fiber.StatusNotFound)
		} else {
			// TODO: handle other errors
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	gridCSS, _ := runtime.GenerateGridCSS(sitePage.Content)

	return Render(c, "layouts/page", fiber.Map{
		"Title":   sitePage.Name,
		"Content": template.HTML(sitePage.Content),
		"GridCSS": template.CSS(gridCSS),
	})
}
