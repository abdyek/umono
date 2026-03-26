package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/i18n"
	"github.com/umono-cms/umono/internal/service"
)

func I18nContext(optionService *service.OptionService, bundle *i18n.Bundle) fiber.Handler {
	return func(c *fiber.Ctx) error {
		language := optionService.GetLanguage()
		translator := bundle.Translator(language)
		locale := translator.Locale()

		c.Locals("I18n", translator)
		c.Locals("Lang", translator.Lang())
		c.Locals("Dir", translator.Dir())
		c.Locals("Locale", locale)
		c.Set("Content-Language", translator.Lang())

		return c.Next()
	}
}
