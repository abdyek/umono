package controllers

import (
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/core"
	"github.com/umono-cms/umono/reqbodies"
)

func Login(c *fiber.Ctx) error {

	l := &reqbodies.Login{}

	if err := c.BodyParser(l); err != nil {
		return err
	}

	if err := validator.New().Struct(l); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("")
	}

	ju := &core.JWTUser{}
	ju.Username = l.Username
	ju.Password = l.Password

	if err := ju.GenerateToken(); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("")
	}

	// TODO: We need to change it for security before first release!
	if ju.Username != os.Getenv("USERNAME") || ju.Password != os.Getenv("PASSWORD") {
		return c.Status(fiber.StatusUnauthorized).SendString("")
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "token"
	cookie.Value = ju.Token
	cookie.Expires = time.Now().Add(48 * time.Hour) // NOTE: Look at core/JWTUser.go

	c.Cookie(cookie)

	return c.JSON(fiber.Map{
		"status": "OK",
	})
}
