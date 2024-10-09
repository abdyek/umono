package controllers

import (
	"bytes"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/reqbodies"
	"github.com/yuin/goldmark"
)

func MDToHTML(c *fiber.Ctx) error {

	mDToHTML := &reqbodies.MDToHTML{}

	if err := c.BodyParser(mDToHTML); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(mDToHTML.MD), &buf); err != nil {
		panic(err)
	}

	return c.JSON(fiber.Map{
		"html": buf.String(),
	})
}
