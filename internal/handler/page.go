package handler

import "github.com/gofiber/fiber/v2"

type PageHandler struct{}

func NewPageHandler() *PageHandler {
	return &PageHandler{}
}

func (h *PageHandler) RenderAdmin(c *fiber.Ctx) error {
	return c.Render("pages/admin", fiber.Map{}, "layouts/admin")
}

func (h *PageHandler) GetJoy(c *fiber.Ctx) error {
	return c.Render("partials/joy", fiber.Map{})
}
