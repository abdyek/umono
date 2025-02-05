package controllers

import (
	"bytes"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/reqbodies"
	"github.com/umono-cms/umono/umono"
	"github.com/yuin/goldmark"
)

// Deprecated
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

func UmonoLangToHTML(c *fiber.Ctx) error {

	umonoLangToHTML := &reqbodies.UmonoLangToHTML{}

	if err := c.BodyParser(umonoLangToHTML); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"html": umono.Lang.Convert(umonoLangToHTML.UmonoLang),
	})
}
