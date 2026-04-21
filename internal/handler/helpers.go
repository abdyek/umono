package handler

import (
	"errors"
	"fmt"
	"html/template"
	"strings"

	umono "github.com/umono-cms/umono"
	"github.com/umono-cms/umono/internal/runtime"
	"github.com/umono-cms/umono/internal/service"

	"github.com/gofiber/fiber/v2"
)

func Render(c *fiber.Ctx, view string, data fiber.Map, layouts ...string) error {
	if data == nil {
		data = fiber.Map{}
	}
	data["I18n"] = c.Locals("I18n")
	data["Lang"] = c.Locals("Lang")
	data["Dir"] = c.Locals("Dir")
	data["IsHTMX"] = c.Locals("IsHTMX")
	data["Version"] = umono.Version
	data["IsAlreadyOnSettings"] = strings.Contains(c.Get("HX-Current-URL"), "/admin/settings")
	data["IsAlreadyOnSitePages"] = strings.Contains(c.Get("HX-Current-URL"), "/admin/site-pages")
	data["IsAlreadyOnComponents"] = strings.Contains(c.Get("HX-Current-URL"), "/admin/components")
	data["IsAlreadyOnMedia"] = strings.Contains(c.Get("HX-Current-URL"), "/admin/media")
	return c.Render(view, data, layouts...)
}

func localizedNotFoundDefaults(c *fiber.Ctx) (string, string) {
	translator, _ := c.Locals("I18n").(interface{ T(string) string })
	if translator == nil {
		return "Page Not Found", "# Page Not Found\nThe page you're looking for may have been moved, deleted, or never existed."
	}

	return translator.T("settings.404_page.default_title"), translator.T("settings.404_page.default_content")
}

func translate(c *fiber.Ctx, key string) string {
	translator, _ := c.Locals("I18n").(interface{ T(string) string })
	if translator == nil {
		return key
	}

	return translator.T(key)
}

func translatedValidationError(c *fiber.Ctx, err error) string {
	switch {
	case errors.Is(err, service.ErrInvalidSlug):
		return translate(c, "admin.pages.errors.invalid_slug")
	case errors.Is(err, service.ErrSlugAlreadyExists):
		return translate(c, "admin.pages.errors.slug_exists")
	case errors.Is(err, service.ErrNameRequired):
		return translate(c, "common.errors.name_required")
	case errors.Is(err, service.ErrInvalidComponentName):
		return translate(c, "admin.components.errors.invalid_name")
	case errors.Is(err, service.ErrComponentNameAlreadyExists):
		return translate(c, "admin.components.errors.name_exists")
	case errors.Is(err, service.ErrStorageNameRequired):
		return translate(c, "settings.storage.errors.name_required")
	case errors.Is(err, service.ErrStorageEndpointRequired):
		return translate(c, "settings.storage.errors.endpoint_required")
	case errors.Is(err, service.ErrStorageRegionRequired):
		return translate(c, "settings.storage.errors.region_required")
	case errors.Is(err, service.ErrStorageBucketRequired):
		return translate(c, "settings.storage.errors.bucket_required")
	case errors.Is(err, service.ErrStorageAccessKeyRequired):
		return translate(c, "settings.storage.errors.access_key_required")
	case errors.Is(err, service.ErrStorageSecretKeyRequired):
		return translate(c, "settings.storage.errors.secret_key_required")
	default:
		return err.Error()
	}
}

func buildPreviewHTML(renderedHTML string) string {
	gridCSS, _ := runtime.GenerateGridCSS(renderedHTML)
	if gridCSS == "" {
		return renderedHTML
	}

	return fmt.Sprintf("<style>%s</style>%s", gridCSS, renderedHTML)
}

func mustPreviewHTML(renderedHTML string) template.HTML {
	return template.HTML(buildPreviewHTML(renderedHTML))
}
