package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/database"
	"github.com/umono-cms/umono/models"
	"github.com/umono-cms/umono/validation"
)

func SlugCheckUsable(c *fiber.Ctx) error {

	slug := c.Params("slug")

	slugStruct := struct {
		Slug string `validate:"slug"`
	}{Slug: slug}

	var unusable bool
	if !validation.Validator.Validate(&slugStruct) {
		unusable = true
	}

	availablePage := models.Page{
		Slug: slug,
	}

	if !unusable {
		(&availablePage).FillBySlug(database.DB)
	}

	var alreadyUsed bool
	if availablePage.ID != 0 {
		alreadyUsed = true
	}

	return c.JSON(fiber.Map{
		"already_used": alreadyUsed,
		"unusable":     unusable,
	})
}
