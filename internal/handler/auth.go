package handler

import (
	"encoding/base64"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/umono-cms/umono/internal/handler/middleware"
	"golang.org/x/crypto/bcrypt"
)

type authHandler struct {
	store *session.Store
}

func NewAuthHandler(store *session.Store) *authHandler {
	return &authHandler{
		store: store,
	}
}

func (h *authHandler) RenderLogin(c *fiber.Ctx) error {
	return Render(c, "pages/login", fiber.Map{}, "layouts/admin")
}

func (h *authHandler) Login(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	hashedUsername, _ := base64.StdEncoding.DecodeString(os.Getenv("HASHED_USERNAME"))
	hashedPassword, _ := base64.StdEncoding.DecodeString(os.Getenv("HASHED_PASSWORD"))

	if bcrypt.CompareHashAndPassword(hashedUsername, []byte(username)) != nil ||
		bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)) != nil {
		return Render(c, "partials/invalid-credentials", fiber.Map{})
	}

	session, err := h.store.Get(c)
	if err != nil {
		return err
	}

	if err := session.Regenerate(); err != nil {
		return err
	}

	session.Set("username", username)
	if err := session.Save(); err != nil {
		return err
	}

	c.Set("HX-Redirect", "/admin")
	return nil
}

func (h *authHandler) Logout(c *fiber.Ctx) error {
	session, err := h.store.Get(c)
	if err != nil {
		return c.Redirect("/admin/login")
	}

	if err := session.Destroy(); err != nil {
		return err
	}

	return middleware.SmartRedirect(c, "/admin/login")
}
