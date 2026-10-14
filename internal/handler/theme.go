package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/theme"
)

// themeHandler serves the themes under /_theme/{release}/. The release
// changes with the content, so the files are cached forever.
type themeHandler struct {
	fallback theme.Fallback
}

func NewThemeHandler(fallback theme.Fallback) *themeHandler {
	return &themeHandler{fallback: fallback}
}

// Serve serves the fallback theme's theme.css. Every other path is 404.
func (h *themeHandler) Serve(c *fiber.Ctx) error {
	if c.Params("release") != h.fallback.Release || c.Params("*") != "theme.css" {
		return c.SendStatus(fiber.StatusNotFound)
	}
	c.Set("Content-Type", "text/css; charset=utf-8")
	c.Set("Cache-Control", "public, max-age=31536000, immutable")
	return c.Send(h.fallback.CSS)
}
