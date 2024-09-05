package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/database"
	"github.com/umono-cms/umono/models"
)

func ReadAllPages(c *fiber.Ctx) error {
	db := database.DB

	var pages []models.Page
	db.Find(&pages)

	return c.JSON(fiber.Map{
		"pages": pages,
	})
}
