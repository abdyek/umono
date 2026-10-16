package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
	"github.com/umono-cms/umono/internal/view"
)

type SitePageHandler struct {
	sitePageService  *service.SitePageService
	componentService *service.ComponentService
}

func NewSitePageHandler(sps *service.SitePageService, cs *service.ComponentService) *SitePageHandler {
	return &SitePageHandler{
		sitePageService:  sps,
		componentService: cs,
	}
}

func (h *SitePageHandler) Create(c *fiber.Ctx) error {
	// TODO: Add validator

	sitePage := models.SitePage{
		Name:    c.FormValue("name"),
		Slug:    c.FormValue("slug"),
		Content: c.FormValue("content"),
		Enabled: c.FormValue("enabled") == "1",
	}

	created, errs := h.sitePageService.Create(sitePage)

	slugErr := ""
	nameErr := ""
	if len(errs) > 0 {
		if err := service.ErrInvalidSlug; service.ErrorsIs(errs, err) {
			slugErr = translatedValidationError(c, err)
		} else if err := service.ErrSlugAlreadyExists; service.ErrorsIs(errs, err) {
			slugErr = translatedValidationError(c, err)
		}
		if err := service.ErrNameRequired; service.ErrorsIs(errs, err) {
			nameErr = translatedValidationError(c, err)
		}
	} else {
		c.Set("HX-Push-Url", "/admin/site-pages/"+strconv.FormatUint(uint64(created.ID), 10))
		sitePage = created
	}

	output, diagnostics := sitePagePreview(c.UserContext(), h.sitePageService, sitePage.Content)
	return Render(c, "partials/htmx/site-page-editor", fiber.Map{
		"SitePageEditor": view.SitePageEditor(sitePage, output, diagnostics, slugErr, nameErr, c.Locals("I18n")),
		"SitePageUl":     sitePageList(h.sitePageService, 0, sitePage.ID),
		"ComponentUl":    componentList(h.componentService, 0, 0),
	})
}

func (h *SitePageHandler) Update(c *fiber.Ctx) error {
	// TODO: Add validator
	u64, _ := strconv.ParseUint(c.FormValue("id"), 10, 0)

	sitePage := models.SitePage{
		ID:      uint(u64),
		Name:    c.FormValue("name"),
		Slug:    c.FormValue("slug"),
		Content: c.FormValue("content"),
		Enabled: c.FormValue("enabled") == "1",
	}

	updated, errs := h.sitePageService.Update(sitePage)

	slugErr := ""
	nameErr := ""
	if len(errs) > 0 {
		if err := service.ErrInvalidSlug; service.ErrorsIs(errs, err) {
			slugErr = translatedValidationError(c, err)
		} else if err := service.ErrSlugAlreadyExists; service.ErrorsIs(errs, err) {
			slugErr = translatedValidationError(c, err)
		}
		if err := service.ErrNameRequired; service.ErrorsIs(errs, err) {
			nameErr = translatedValidationError(c, err)
		}
	} else {
		sitePage = updated
	}

	output, diagnostics := sitePagePreview(c.UserContext(), h.sitePageService, sitePage.Content)
	return Render(c, "partials/htmx/site-page-editor", fiber.Map{
		"SitePageEditor": view.SitePageEditor(sitePage, output, diagnostics, slugErr, nameErr, c.Locals("I18n")),
		"SitePageUl":     sitePageList(h.sitePageService, 0, sitePage.ID),
		"ComponentUl":    componentList(h.componentService, 0, 0),
	})
}

func (h *SitePageHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.ErrBadRequest
	}

	if err := h.sitePageService.Delete(uint(id)); err != nil {
		return err
	}

	c.Set("HX-Redirect", "/admin/site-pages/new")
	return c.SendStatus(fiber.StatusNoContent)
}

// SaveNotFoundPage saves the 404 page from its settings screen.
func (h *SitePageHandler) SaveNotFoundPage(c *fiber.Ctx) error {
	contractID, layoutID, err := bindingValue(c.FormValue("layout"))
	if err != nil {
		return err
	}
	if err := h.sitePageService.SaveNotFoundPage(c.FormValue("title"), c.FormValue("content"), contractID, layoutID); err != nil {
		return err
	}

	c.Set("HX-Trigger", "notFoundPageOptionSaved")
	return c.SendStatus(fiber.StatusOK)
}

func (h *SitePageHandler) CheckSlug(c *fiber.Ctx) error {
	slug := c.Query("slug")
	id := c.Query("id")

	u64ID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.SendStatus(400)
	}

	err = h.sitePageService.CheckSlug(slug, uint(u64ID))
	if err != nil {
		return Render(c, "partials/slug-error", fiber.Map{
			"SlugErr": translatedValidationError(c, err),
		})
	}

	return c.SendString("")
}

func (h *SitePageHandler) RenderNewPageSiteEditor(c *fiber.Ctx) error {
	return Render(c, "pages/admin", fiber.Map{
		"SitePageEditor": view.SitePageEditor(models.SitePage{}, mustPreviewHTML(""), nil, "", "", c.Locals("I18n")),
		"SitePageUl":     sitePageList(h.sitePageService, 0, 0),
		"ComponentUl":    componentList(h.componentService, 0, 0),
	}, "layouts/admin")
}

// bindingValue parses the Layout field's value, "<record ID>:<layout id>",
// into a binding. An empty value is no binding.
func bindingValue(v string) (*uint, *string, error) {
	if v == "" {
		return nil, nil, nil
	}
	id, layout, ok := strings.Cut(v, ":")
	if !ok || layout == "" {
		return nil, nil, fiber.ErrBadRequest
	}
	n, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, nil, fiber.ErrBadRequest
	}
	contractID := uint(n)
	return &contractID, &layout, nil
}
