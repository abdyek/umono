package view

import "github.com/gofiber/fiber/v2"

type ViewData interface {
	Key() string
	Value() any
}

func View(vd ...ViewData) fiber.Map {
	output := fiber.Map{}
	for _, d := range vd {
		output[d.Key()] = d.Value()
	}
	return output
}
