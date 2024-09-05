package controllers

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/database"
	"github.com/umono-cms/umono/models"
	"github.com/umono-cms/umono/reqbodies"
)

func CreatePage(c *fiber.Ctx) error {
	cp := &reqbodies.CreatePage{}

	if err := c.BodyParser(cp); err != nil {
		return err
	}

	if err := validator.New().Struct(cp); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("")
	}

	db := database.DB

	now := time.Now()

	saved := models.Page{
		Name:           cp.Page.Name,
		Slug:           cp.Page.Slug,
		Content:        cp.Page.Content,
		LastModifiedAt: &now,
	}

	db.Create(&saved)

	return c.JSON(fiber.Map{
		"page": saved,
	})
}

func ReadAllPages(c *fiber.Ctx) error {
	db := database.DB

	var pages []models.Page
	db.Find(&pages)

	return c.JSON(fiber.Map{
		"pages": pages,
	})
}
