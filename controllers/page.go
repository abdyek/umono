package controllers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/cache"
	"github.com/umono-cms/umono/database"
	"github.com/umono-cms/umono/models"
	"github.com/umono-cms/umono/reqbodies"
	"github.com/umono-cms/umono/validation"
)

func CreatePage(c *fiber.Ctx) error {
	cp := &reqbodies.CreatePage{}

	if err := c.BodyParser(cp); err != nil {
		return err
	}

	if !validation.Validator.Validate(cp) {
		return c.Status(fiber.StatusBadRequest).SendString("")
	}

	db := database.DB

	available := models.Page{
		Slug: cp.Page.Slug,
	}

	(&available).FillBySlug(db)
	if available.ID != 0 {
		return c.Status(fiber.StatusConflict).SendString("")
	}

	now := time.Now()

	saved := models.Page{
		Name:           cp.Page.Name,
		Slug:           cp.Page.Slug,
		Content:        cp.Page.Content,
		Enabled:        cp.Page.Enabled,
		LastModifiedAt: &now,
	}

	db.Create(&saved)

	cache.Page.Load(saved)

	return c.JSON(fiber.Map{
		"page": saved,
	})
}

func ServePage(c *fiber.Ctx) error {
	slug := c.Params("slug")

	page, available := cache.Page.GetPage(slug)
	if !available {
		c.Status(fiber.StatusNotFound)
		page, available = cache.Page.GetPage("_404")
		if !available {
			return c.SendString("")
		}
		return c.Render("page", fiber.Map{
			"Title":   page.Name,
			"Content": page.Content,
		})
	}

	return c.Render("page", fiber.Map{
		"Title":   page.Name,
		"Content": page.Content,
	})
}

func ReadPage(c *fiber.Ctx) error {
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

	return c.JSON(fiber.Map{
		"page": pageFromDB,
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

	if !validation.Validator.Validate(up) {
		return c.Status(fiber.StatusBadRequest).SendString("")
	}

	db := database.DB

	var pageFromDB models.Page

	db.First(&pageFromDB, up.Page.ID)

	if pageFromDB.ID == 0 {
		return c.Status(fiber.StatusNotFound).SendString("")
	}

	available := models.Page{
		Slug: up.Page.Slug,
	}

	(&available).FillBySlug(db)
	if available.ID != 0 && available.ID != pageFromDB.ID {
		return c.Status(fiber.StatusConflict).SendString("")
	}

	now := time.Now()

	updated := models.Page{
		ID:             up.Page.ID,
		Name:           up.Page.Name,
		Slug:           up.Page.Slug,
		Content:        up.Page.Content,
		LastModifiedAt: &now,
		Enabled:        up.Page.Enabled,
	}

	db.Model(&updated).Select("Name", "Slug", "Content", "LastModifiedAt", "Enabled").Updates(updated)

	if up.Page.Slug != pageFromDB.Slug {
		cache.Page.Remove(pageFromDB)
	}

	if up.Page.Enabled {
		cache.Page.Load(updated)
	} else {
		cache.Page.Remove(updated)
	}

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
	cache.Page.Remove(pageFromDB)

	return c.JSON(fiber.Map{
		"status": "OK",
	})
}
