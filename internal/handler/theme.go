package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/theme"
)

// themeAssets are the assets of the active themes by release, and those of
// the theme in development.
type themeAssets interface {
	Asset(release, name string) ([]byte, bool)
	DevAsset(name string) ([]byte, bool, error)
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
// the active themes and of the theme in development. Every other path is 404.
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
	if release == theme.DevRelease {
		return h.serveDev(c, contentType, name)
	}
	b, ok := h.assets.Asset(release, name)
	if !ok {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return sendThemeFile(c, contentType, b)
}

// serveDev serves assets/name of the theme in development. Every response is
// no-store, so a reload gets the current file (privilege #6). A failed check
// of theme.css is an error status; the dev proxy shows its text.
func (h *themeHandler) serveDev(c *fiber.Ctx, contentType, name string) error {
	c.Set("Cache-Control", "no-store")
	b, ok, err := h.assets.DevAsset(name)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !ok {
		return c.SendStatus(fiber.StatusNotFound)
	}
	c.Set("Content-Type", contentType)
	return c.Send(b)
}

func sendThemeFile(c *fiber.Ctx, contentType string, b []byte) error {
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=31536000, immutable")
	return c.Send(b)
}
