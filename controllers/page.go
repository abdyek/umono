package controllers

import (
	"strconv"
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

	// TODO: if published true, generate it

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

func UpdatePage(c *fiber.Ctx) error {
	up := &reqbodies.UpdatePage{}

	if err := c.BodyParser(up); err != nil {
		return err
	}

	if err := validator.New().Struct(up); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("")
	}

	db := database.DB

	var pageFromDB models.Page

	db.First(&pageFromDB, up.Page.ID)

	if pageFromDB.ID == 0 {
		return c.Status(fiber.StatusNotFound).SendString("")
	}

	now := time.Now()

	updated := models.Page{
		ID:             up.Page.ID,
		Name:           up.Page.Name,
		Slug:           up.Page.Slug,
		Content:        up.Page.Content,
		LastModifiedAt: &now,
		Published:      up.Page.Published,
	}

	db.Model(&updated).Select("Name", "Slug", "Content", "LastModifiedAt", "Published").Updates(updated)

	// TODO: if published true, regenerate it

	return c.JSON(fiber.Map{
		"page": updated,
	})
}

func DeletePage(c *fiber.Ctx) error {

	u64, err := strconv.ParseUint(c.Params("ID"), 10, 64)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("")
	}

	db := database.DB

	var pageFromDB models.Page
	db.First(&pageFromDB, uint(u64))

	if pageFromDB.ID == 0 {
		return c.Status(fiber.StatusNotFound).SendString("")
	}

	db.Delete(pageFromDB)

	return c.JSON(fiber.Map{
		"status": "OK",
	})
}
