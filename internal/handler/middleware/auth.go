package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func Logged(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !isAdmin(c, store) {
			return SmartRedirect(c, "/admin/login")
		}
		return c.Next()
	}
}

func Guest(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if isAdmin(c, store) {
			return SmartRedirect(c, "/admin")
		}
		return c.Next()
	}
}

func isAdmin(c *fiber.Ctx, store *session.Store) bool {
	sess, _ := store.Get(c)

	if sess.Get("username") != nil {
		return true
	}
	return false
}
