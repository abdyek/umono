package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/theme"
)

// themeAssets are the assets of the active themes by release.
type themeAssets interface {
	Asset(release, name string) ([]byte, bool)
}

// themeHandler serves the themes under /_theme/{release}/. The release
// changes with the content, so the files are cached forever.
type themeHandler struct {
	fallback theme.Fallback
	assets   themeAssets
}

func NewThemeHandler(fallback theme.Fallback, assets themeAssets) *themeHandler {
	return &themeHandler{fallback: fallback, assets: assets}
}

// Serve serves the fallback theme's theme.css and the files under assets/ of
// the active themes. Every other path is 404.
func (h *themeHandler) Serve(c *fiber.Ctx) error {
	release, name := c.Params("release"), c.Params("*")
	if release == h.fallback.Release {
		if name != "theme.css" {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return sendThemeFile(c, theme.ContentType(name), h.fallback.CSS)
	}
	contentType := theme.ContentType(name)
	if contentType == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}
	b, ok := h.assets.Asset(release, name)
	if !ok {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return sendThemeFile(c, contentType, b)
}

func sendThemeFile(c *fiber.Ctx, contentType string, b []byte) error {
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=31536000, immutable")
	return c.Send(b)
}
