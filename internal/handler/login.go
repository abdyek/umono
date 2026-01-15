package handler

import (
	"encoding/base64"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/crypto/bcrypt"
)

type LoginHandler struct {
	store *session.Store
}

func NewLoginHandler(store *session.Store) *LoginHandler {
	return &LoginHandler{
		store: store,
	}
}

func (h *LoginHandler) Login(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	hashedUsername, _ := base64.StdEncoding.DecodeString(os.Getenv("HASHED_USERNAME"))
	hashedPassword, _ := base64.StdEncoding.DecodeString(os.Getenv("HASHED_PASSWORD"))

	if bcrypt.CompareHashAndPassword(hashedUsername, []byte(username)) != nil ||
		bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)) != nil {
		return c.Render("partials/invalid-credentials", fiber.Map{})
	}

	session, _ := h.store.Get(c)
	session.Set("username", username)
	session.Save()

	c.Set("HX-Redirect", "/admin")
	return nil
}
