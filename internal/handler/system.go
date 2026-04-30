package handler

import (
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
	"github.com/umono-cms/umono/internal/view"
)

type systemHandler struct {
	systemService *service.SystemService
	jobService    *service.JobService
}

func NewSystemHandler(ss *service.SystemService, js *service.JobService) *systemHandler {
	return &systemHandler{
		systemService: ss,
		jobService:    js,
	}
}

func (h *systemHandler) Index(c *fiber.Ctx) error {
	menuItems := h.systemService.MenuItems()
	if len(menuItems) == 0 {
		return c.SendStatus(fiber.StatusOK)
	}

	return c.Redirect("/admin/system/" + menuItems[0].Slug)
}

func (h *systemHandler) RenderJobs(c *fiber.Ctx) error {
	partial := "partials/system-jobs"
	layouts := []string{"layouts/system", "layouts/admin"}

	if isSystemContentSwap(c) {
		partial = "partials/htmx/system-jobs"
		layouts = []string{}
	}

	jobs, err := h.jobService.ListAll(c.UserContext())
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return Render(c, partial, fiber.Map{
		"SystemUl":   h.systemUl(c, path.Base(c.Path())),
		"Jobs":       buildSystemJobs(c, jobs),
		"RetryLabel": translate(c, "system.jobs.retry"),
	}, layouts...)
}

func (h *systemHandler) RetryJob(c *fiber.Ctx) error {
	err := h.jobService.RetryFailed(c.UserContext(), c.Params("id"))
	if err != nil {
		if errors.Is(err, service.ErrJobNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		if errors.Is(err, service.ErrJobRetryNotAllowed) {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		return fiber.ErrInternalServerError
	}

	return h.RenderJobs(c)
}

func (h *systemHandler) systemUl(c *fiber.Ctx, activeSlug string) []view.SettingsLi {
	ul := []view.SettingsLi{}
	translator, _ := c.Locals("I18n").(interface{ T(string) string })
	for _, mi := range h.systemService.MenuItems() {
		title := mi.TitleKey
		if translator != nil {
			title = translator.T(mi.TitleKey)
		}

		ul = append(ul, view.SettingsLi{
			Title:    title,
			Slug:     mi.Slug,
			IsActive: mi.Slug == activeSlug,
		})
	}
	return ul
}

type systemJobRow struct {
	ID             string
	Type           string
	Status         string
	StatusLabel    string
	StatusClasses  string
	Attempts       string
	RunAt          string
	LastError      string
	UpdatedAt      string
	CreatedAt      string
	CanRetry       bool
	RetryURL       string
	RetryAriaLabel string
}

func buildSystemJobs(c *fiber.Ctx, jobs []models.Job) []systemJobRow {
	rows := make([]systemJobRow, 0, len(jobs))
	for _, job := range jobs {
		rows = append(rows, systemJobRow{
			ID:             job.ID,
			Type:           job.Type,
			Status:         job.Status,
			StatusLabel:    jobStatusLabel(c, job.Status),
			StatusClasses:  jobStatusClasses(job.Status),
			Attempts:       jobAttempts(job),
			RunAt:          formatJobTime(job.RunAt),
			LastError:      jobLastError(c, job.LastError),
			UpdatedAt:      formatJobTime(job.UpdatedAt),
			CreatedAt:      formatJobTime(job.CreatedAt),
			CanRetry:       job.Status == models.JobStatusFailed,
			RetryURL:       "/admin/system/jobs/" + job.ID + "/retry",
			RetryAriaLabel: translate(c, "system.jobs.retry"),
		})
	}
	return rows
}

func jobStatusLabel(c *fiber.Ctx, status string) string {
	switch status {
	case models.JobStatusFailed:
		return translate(c, "system.jobs.status.failed")
	case models.JobStatusProcessing:
		return translate(c, "system.jobs.status.processing")
	case models.JobStatusPending:
		return translate(c, "system.jobs.status.pending")
	case models.JobStatusDone:
		return translate(c, "system.jobs.status.done")
	default:
		return status
	}
}

func jobStatusClasses(status string) string {
	switch status {
	case models.JobStatusFailed:
		return "border-red-500/25 bg-red-500/10 text-red-200"
	case models.JobStatusProcessing:
		return "border-sky-500/25 bg-sky-500/10 text-sky-200"
	case models.JobStatusPending:
		return "border-amber-500/25 bg-amber-500/10 text-amber-100"
	case models.JobStatusDone:
		return "border-emerald-500/25 bg-emerald-500/10 text-emerald-200"
	default:
		return "border-neutral-700 bg-neutral-800 text-neutral-300"
	}
}

func jobAttempts(job models.Job) string {
	return fmt.Sprintf("%d / %d", job.Attempts, job.MaxRetry)
}

func jobLastError(c *fiber.Ctx, lastError *string) string {
	if lastError == nil || *lastError == "" {
		return translate(c, "system.jobs.empty_value")
	}
	return *lastError
}

func formatJobTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format(time.DateTime)
}

func isSystemContentSwap(c *fiber.Ctx) bool {
	if c.Get("HX-Request") != "true" {
		return false
	}

	switch c.Get("HX-Target") {
	case "system-content", "system-jobs-content":
		return true
	default:
		return false
	}
}
