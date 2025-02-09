package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/database"
	"github.com/umono-cms/umono/models"
	"github.com/umono-cms/umono/reqbodies"
	"github.com/umono-cms/umono/validation"
)

// NOTE: controllers' functions have same patterns. You can create CRUD services and refactor it.

func CreateComponent(c *fiber.Ctx) error {
	cc := &reqbodies.CreateComponent{}

	if err := c.BodyParser(cc); err != nil {
		return err
	}

	if !validation.Validator.Validate(cc) {
		return c.Status(fiber.StatusBadRequest).SendString("")
	}

	db := database.DB

	available := models.Component{
		Name: cc.Component.Name,
	}

	(&available).FillByName(db)
	if available.ID != 0 {
		return c.Status(fiber.StatusConflict).SendString("")
	}

	now := time.Now()

	saved := models.Component{
		Name:           cc.Component.Name,
		Content:        cc.Component.Content,
		LastModifiedAt: &now,
	}

	db.Create(&saved)

	// TODO: Reload pages that has the component

	return c.JSON(fiber.Map{
		"component": saved,
	})
}
