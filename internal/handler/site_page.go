package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
	"github.com/umono-cms/umono/internal/view"
)

type SitePageHandler struct {
	sitePageService  service.SitePageService
	componentService service.ComponentService
}

func NewSitePageHandler(sps service.SitePageService, cs service.ComponentService) *SitePageHandler {
	return &SitePageHandler{
		sitePageService:  sps,
		componentService: cs,
	}
}

func (h *SitePageHandler) Create(c *fiber.Ctx) error {

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
			slugErr = err.Error()
		} else if err := service.ErrSlugAlreadyExists; service.ErrorsIs(errs, err) {
			slugErr = err.Error()
		}
		if err := service.ErrNameRequired; service.ErrorsIs(errs, err) {
			nameErr = err.Error()
		}
	} else {
		c.Set("HX-Push-Url", "/admin/site-pages/"+strconv.FormatUint(uint64(created.ID), 10))
		sitePage = created
	}

	return Render(c, "partials/htmx/admin-site-page-editor", fiber.Map{
		"SitePageEditor": view.SitePageEditor(sitePage, h.sitePageService.MustPreview(sitePage.Content), slugErr, nameErr),
		"SitePageUl":     view.SitePageUl(h.sitePageService.GetAll(), sitePage.ID),
	})
}

func (h *SitePageHandler) Update(c *fiber.Ctx) error {
	// TODO: Complete it
	return c.SendString("Here site page update")
}

func (h *SitePageHandler) Delete(c *fiber.Ctx) error {
	// TODO: Complete it
	return c.SendString("Here site page delete")
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
			"SlugErr": err.Error(),
		})
	}

	return c.SendString("")
}

func (h *SitePageHandler) RenderNewPageSiteEditor(c *fiber.Ctx) error {
	return Render(c, "pages/admin", fiber.Map{
		"SitePageEditor": view.SitePageEditor(models.SitePage{}, "", "", ""),
		"SitePageUl":     view.SitePageUl(h.sitePageService.GetAll(), 0),
		"ComponentUl":    view.ComponentUl(h.componentService.GetAll(), 0),
	}, "layouts/admin")
}
