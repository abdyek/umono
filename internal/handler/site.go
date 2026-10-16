package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"html/template"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/csp"
	"github.com/umono-cms/umono/internal/service"
)

type siteHandler struct {
	sitePageService *service.SitePageService
	optionService   *service.OptionService
	themeRelease    string
}

func NewSiteHandler(sps *service.SitePageService, os *service.OptionService, themeRelease string) *siteHandler {
	return &siteHandler{
		sitePageService: sps,
		optionService:   os,
		themeRelease:    themeRelease,
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

// RenderSitePage renders a page, or the 404 page, for a visitor, with its
// group's active theme or the fallback. A page's weak ETag is computed from
// the shell without the nonce, so it stays the same while the nonce changes
// with each request. A matching If-None-Match gets a 304 without a
// Content-Security-Policy, so the browser keeps the CSP whose nonce matches
// its copy.
func (h *siteHandler) RenderSitePage(c *fiber.Ctx) error {
	ctx := c.UserContext()

	found := true
	page, err := h.sitePageService.GetRenderedBySlug(ctx, c.Params("slug"))
	if err != nil {
		if errors.Is(err, service.ErrSitePageNotFound) {
			defaultTitle, defaultContent := localizedNotFoundDefaults(c)
			page, err = h.sitePageService.GetNotFoundPage(ctx, defaultTitle, defaultContent)
			if err != nil {
				// TODO: handle other errors
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			c.Status(fiber.StatusNotFound)
			found = false
		} else {
			// TODO: handle other errors
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}

	release := page.Release
	if release == "" {
		release = h.themeRelease
	}

	data := fiber.Map{
		"Title":        page.Name,
		"Content":      template.HTML(page.Content),
		"ThemeRelease": release,
		"Nonce":        "",
	}
	c.Set("Cache-Control", "no-cache")

	if found {
		withoutNonce, err := renderBytes(c, "layouts/page", data)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		sum := sha256.Sum256(withoutNonce)
		etag := "W/" + strconv.Quote(hex.EncodeToString(sum[:]))
		c.Set("ETag", etag)
		if etagMatches(c.Get("If-None-Match"), etag) {
			return c.SendStatus(fiber.StatusNotModified)
		}
	}

	nonce, err := csp.Nonce()
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	c.Set("Content-Security-Policy", csp.Header(nonce, csp.ImageOrigins(page.ImageURLs, c.BaseURL())))

	data["Nonce"] = nonce
	body, err := renderBytes(c, "layouts/page", data)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(body)
}
