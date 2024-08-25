package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/core"
)

func Me(c *fiber.Ctx) error {

	ju := &core.JWTUser{}
	ju.Token = c.Cookies("token")

	err := ju.Resolve()

	var loggedIn bool
	if err == nil {
		loggedIn = true
	}

	return c.JSON(fiber.Map{
		"logged_in": loggedIn,
	})
}
