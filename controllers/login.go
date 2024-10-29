package controllers

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/core"
	"github.com/umono-cms/umono/reqbodies"
	"golang.org/x/crypto/bcrypt"
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

	hashedUsername, e := base64.StdEncoding.DecodeString(os.Getenv("HASHED_USERNAME"))
	hashedPassword, e2 := base64.StdEncoding.DecodeString(os.Getenv("HASHED_PASSWORD"))

	if e != nil {
		fmt.Println(e.Error())
	}

	if e2 != nil {
		fmt.Println(e2.Error())
	}

	if bcrypt.CompareHashAndPassword(hashedUsername, []byte(ju.Username)) != nil ||
		bcrypt.CompareHashAndPassword(hashedPassword, []byte(ju.Password)) != nil {
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
