package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func LogOut(c *fiber.Ctx) error {

	// TODO: No real log outing
	c.Cookie(&fiber.Cookie{
		Name:    "token",
		Value:   "",
		Expires: time.Now().Add(-time.Second),
	})

	return nil
}
