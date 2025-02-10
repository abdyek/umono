package controllers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/database"
	"github.com/umono-cms/umono/models"
	"github.com/umono-cms/umono/reqbodies"
	"github.com/umono-cms/umono/umono"
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

	umono.Lang.SetGlobalComponent(saved.Name, saved.Content)
	// TODO: Reload pages that has the component

	return c.JSON(fiber.Map{
		"component": saved,
	})
}

func ReadComponent(c *fiber.Ctx) error {
	u64, err := strconv.ParseUint(c.Params("ID"), 10, 64)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("")
	}

	db := database.DB

	var fromDB models.Component
	db.First(&fromDB, uint(u64))

	if fromDB.ID == 0 {
		return c.Status(fiber.StatusNotFound).SendString("")
	}

	return c.JSON(fiber.Map{
		"component": fromDB,
	})
}

func ReadAllComponents(c *fiber.Ctx) error {
	db := database.DB

	var comps []models.Component
	db.Find(&comps)

	return c.JSON(fiber.Map{
		"components": comps,
	})
}

func UpdateComponent(c *fiber.Ctx) error {
	uc := &reqbodies.UpdateComponent{}

	if err := c.BodyParser(uc); err != nil {
		return err
	}

	if !validation.Validator.Validate(uc) {
		return c.Status(fiber.StatusBadRequest).SendString("")
	}

	db := database.DB

	var fromDB models.Component

	db.First(&fromDB, uc.Component.ID)

	if fromDB.ID == 0 {
		return c.Status(fiber.StatusNotFound).SendString("")
	}

	available := models.Component{
		Name: uc.Component.Name,
	}

	(&available).FillByName(db)
	if available.ID != 0 && available.ID != fromDB.ID {
		return c.Status(fiber.StatusConflict).SendString("")
	}

	now := time.Now()

	updated := models.Component{
		ID:             uc.Component.ID,
		Name:           uc.Component.Name,
		Content:        uc.Component.Content,
		LastModifiedAt: &now,
	}

	db.Model(&updated).Select("*").Updates(updated)

	umono.Lang.SetGlobalComponent(updated.Name, updated.Content)
	// TODO: Reload pages that has the component

	return c.JSON(fiber.Map{
		"component": updated,
	})
}

func DeleteComponent(c *fiber.Ctx) error {
	u64, err := strconv.ParseUint(c.Params("ID"), 10, 64)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("")
	}

	db := database.DB

	var fromDB models.Component
	db.First(&fromDB, uint(u64))

	if fromDB.ID == 0 {
		return c.Status(fiber.StatusNotFound).SendString("")
	}

	db.Delete(fromDB)

	umono.Lang.RemoveGlobalComponent(fromDB.Name)
	// TODO: Reload pages that has the component

	return c.JSON(fiber.Map{
		"status": "OK",
	})
}
