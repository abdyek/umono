package handler

import (
	"errors"
	"html/template"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/service"
)

type siteHandler struct {
	sitePageService *service.SitePageService
	optionService   *service.OptionService
}

func NewSiteHandler(sps *service.SitePageService, os *service.OptionService) *siteHandler {
	return &siteHandler{
		sitePageService: sps,
		optionService:   os,
	}
}

func (h *siteHandler) RenderRobotsTxt(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/plain; charset=utf-8")
	c.Set("X-Content-Type-Options", "nosniff")
	return c.SendString(h.optionService.GetRobotsTxt())
}

// RenderSitemap serves /sitemap.xml. It is 404 while the site address is not
// set.
func (h *siteHandler) RenderSitemap(c *fiber.Ctx) error {
	siteURL := h.optionService.GetSiteURL()
	if siteURL == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}
	body, err := h.sitePageService.Sitemap(siteURL)
	if err != nil {
		return err
	}
	c.Set("Content-Type", "application/xml; charset=utf-8")
	c.Set("X-Content-Type-Options", "nosniff")
	return c.Send(body)
}

func (h *siteHandler) RenderSitePage(c *fiber.Ctx) error {
	ctx := c.UserContext()

	sitePage, err := h.sitePageService.GetRenderedBySlug(ctx, c.Params("slug"))
	if err != nil {
		if errors.Is(err, service.ErrSitePageNotFound) {
			defaultTitle, defaultContent := localizedNotFoundDefaults(c)
			sitePage, err = h.sitePageService.GetNotFoundPage(ctx, defaultTitle, defaultContent)
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

	return Render(c, "layouts/page", fiber.Map{
		"Title":   sitePage.Name,
		"Content": template.HTML(sitePage.Content),
	})
}
