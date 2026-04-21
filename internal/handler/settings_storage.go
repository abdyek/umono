package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
)

type storagePageData struct {
	List          []storageListItem
	Storage       storageFormData
	HasSelection  bool
	IsCreateMode  bool
	DeleteConfirm string
}

type storageListItem struct {
	ID        string
	Name      string
	Type      string
	TypeLabel string
	Summary   string
	IsDefault bool
	IsActive  bool
}

type storageFormData struct {
	ID          string
	Name        string
	Type        string
	TypeLabel   string
	Endpoint    string
	Region      string
	Bucket      string
	AccessKey   string
	SecretKey   string
	LocalRoot   string
	CreatedAt   string
	UpdatedAt   string
	IsReadonly  bool
	CanDelete   bool
	Errors      map[string]string
	GlobalError string
	SubmitURL   string
	CancelURL   string
	BackURL     string
	PushURL     string
}

func (h *settingsHandler) RenderStorageIndex(c *fiber.Ctx) error {
	return h.renderStorage(c, storagePageData{
		List:         h.buildStorageList(c, "", ""),
		HasSelection: false,
	})
}

func (h *settingsHandler) RenderStorageNew(c *fiber.Ctx) error {
	page := storagePageData{
		List:         h.buildStorageList(c, "", "new"),
		HasSelection: true,
		IsCreateMode: true,
		Storage: storageFormData{
			Type:      models.StorageTypeS3,
			TypeLabel: translate(c, "settings.storage.type.s3"),
			SubmitURL: "/admin/settings/storage",
			CancelURL: "/admin/settings/storage",
			BackURL:   "/admin/settings/storage",
			PushURL:   "/admin/settings/storage/new",
			Errors:    map[string]string{},
		},
	}

	return h.renderStorage(c, page)
}

func (h *settingsHandler) RenderStorageShow(c *fiber.Ctx) error {
	storage, err := h.storageService.GetByID(c.Params("id"))
	if err != nil {
		if errors.Is(err, service.ErrStorageNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return fiber.ErrInternalServerError
	}

	return h.renderStorage(c, storagePageData{
		List:         h.buildStorageList(c, storage.ID, ""),
		HasSelection: true,
		Storage:      h.buildStorageFormData(c, storage, map[string]string{}, ""),
	})
}

func (h *settingsHandler) CreateStorage(c *fiber.Ctx) error {
	input := storageInputFromRequest(c)
	storage, err := h.storageService.CreateS3(input)
	if err != nil {
		return h.renderStorageCreateError(c, input, err)
	}

	c.Set("HX-Push-Url", "/admin/settings/storage/"+storage.ID)

	return h.renderStorage(c, storagePageData{
		List:         h.buildStorageList(c, storage.ID, ""),
		HasSelection: true,
		Storage:      h.buildStorageFormData(c, storage, map[string]string{}, ""),
	})
}

func (h *settingsHandler) UpdateStorage(c *fiber.Ctx) error {
	id := c.Params("id")
	input := storageInputFromRequest(c)
	storage, err := h.storageService.UpdateS3(id, input)
	if err != nil {
		if errors.Is(err, service.ErrStorageNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		if errors.Is(err, service.ErrStorageReadonly) {
			current, getErr := h.storageService.GetByID(id)
			if getErr != nil {
				return c.SendStatus(fiber.StatusNotFound)
			}
			return h.renderStorage(c, storagePageData{
				List:         h.buildStorageList(c, current.ID, ""),
				HasSelection: true,
				Storage: h.buildStorageFormData(
					c,
					current,
					map[string]string{},
					translate(c, "settings.storage.local.readonly_notice"),
				),
			})
		}

		current, getErr := h.storageService.GetByID(id)
		if getErr != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}

		form := h.buildStorageFormData(c, current, storageErrors(c, err), "")
		form.Name = input.Name
		form.Endpoint = input.Endpoint
		form.Region = input.Region
		form.Bucket = input.Bucket
		form.AccessKey = input.AccessKey
		form.SecretKey = input.SecretKey

		return h.renderStorage(c, storagePageData{
			List:         h.buildStorageList(c, current.ID, ""),
			HasSelection: true,
			Storage:      form,
		})
	}

	c.Set("HX-Push-Url", "/admin/settings/storage/"+storage.ID)

	return h.renderStorage(c, storagePageData{
		List:         h.buildStorageList(c, storage.ID, ""),
		HasSelection: true,
		Storage:      h.buildStorageFormData(c, storage, map[string]string{}, ""),
	})
}

func (h *settingsHandler) DeleteStorage(c *fiber.Ctx) error {
	if err := h.storageService.Delete(c.Params("id")); err != nil {
		if errors.Is(err, service.ErrStorageNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		if errors.Is(err, service.ErrStorageDeleteRestricted) {
			storage, getErr := h.storageService.GetByID(c.Params("id"))
			if getErr != nil {
				return c.SendStatus(fiber.StatusNotFound)
			}

			return h.renderStorage(c, storagePageData{
				List:         h.buildStorageList(c, storage.ID, ""),
				HasSelection: true,
				Storage: h.buildStorageFormData(
					c,
					storage,
					map[string]string{},
					translate(c, "settings.storage.local.delete_notice"),
				),
			})
		}
		return fiber.ErrInternalServerError
	}

	c.Set("HX-Push-Url", "/admin/settings/storage")

	return h.renderStorage(c, storagePageData{
		List:         h.buildStorageList(c, "", ""),
		HasSelection: false,
	})
}

func (h *settingsHandler) renderStorageCreateError(c *fiber.Ctx, input service.StorageInput, err error) error {
	form := storageFormData{
		Type:        models.StorageTypeS3,
		TypeLabel:   translate(c, "settings.storage.type.s3"),
		Name:        input.Name,
		Endpoint:    input.Endpoint,
		Region:      input.Region,
		Bucket:      input.Bucket,
		AccessKey:   input.AccessKey,
		SecretKey:   input.SecretKey,
		SubmitURL:   "/admin/settings/storage",
		CancelURL:   "/admin/settings/storage",
		BackURL:     "/admin/settings/storage",
		PushURL:     "/admin/settings/storage/new",
		Errors:      storageErrors(c, err),
		GlobalError: storageGlobalError(c, err),
	}

	return h.renderStorage(c, storagePageData{
		List:         h.buildStorageList(c, "", "new"),
		HasSelection: true,
		IsCreateMode: true,
		Storage:      form,
	})
}

func (h *settingsHandler) renderStorage(c *fiber.Ctx, page storagePageData) error {
	partial := "partials/settings-storage"
	layouts := []string{"layouts/settings", "layouts/admin"}

	if isSettingsContentSwap(c) {
		partial = "partials/htmx/settings-storage"
		layouts = []string{}
	}

	return Render(c, partial, fiber.Map{
		"SettingsUl":     h.settingsUl(c, "storage"),
		"StoragePage":    page,
		"DeleteConfirm":  translate(c, "settings.storage.delete_confirm"),
		"ReadonlyNotice": translate(c, "settings.storage.local.readonly_notice"),
	}, layouts...)
}

func (h *settingsHandler) buildStorageList(c *fiber.Ctx, activeID, createRoute string) []storageListItem {
	storages := h.storageService.GetAll()
	items := make([]storageListItem, 0, len(storages))
	for _, storage := range storages {
		items = append(items, storageListItem{
			ID:        storage.ID,
			Name:      storage.Name,
			Type:      storage.Type,
			TypeLabel: storageTypeLabel(c, storage.Type),
			Summary:   storageSummary(c, storage),
			IsDefault: storage.ID == service.DefaultLocalStorageID,
			IsActive:  storage.ID == activeID,
		})
	}

	if createRoute == "new" {
		items = append(items, storageListItem{
			ID:        "new",
			Name:      translate(c, "settings.storage.new.title"),
			TypeLabel: translate(c, "common.create"),
			Summary:   translate(c, "settings.storage.new.description"),
			IsActive:  true,
		})
	}

	return items
}

func (h *settingsHandler) buildStorageFormData(c *fiber.Ctx, storage models.Storage, errs map[string]string, globalError string) storageFormData {
	form := storageFormData{
		ID:          storage.ID,
		Name:        storage.Name,
		Type:        storage.Type,
		TypeLabel:   storageTypeLabel(c, storage.Type),
		CreatedAt:   storage.CreatedAt.Local().Format(time.DateTime),
		UpdatedAt:   storage.UpdatedAt.Local().Format(time.DateTime),
		IsReadonly:  storage.Type != models.StorageTypeS3,
		CanDelete:   storage.ID != service.DefaultLocalStorageID,
		Errors:      errs,
		GlobalError: globalError,
		BackURL:     "/admin/settings/storage",
		PushURL:     "/admin/settings/storage/" + storage.ID,
	}

	if storage.Type == models.StorageTypeS3 {
		form.Endpoint = service.StorageConfigValue(storage, "endpoint")
		form.Region = service.StorageConfigValue(storage, "region")
		form.Bucket = service.StorageConfigValue(storage, "bucket")
		form.AccessKey = service.StorageConfigValue(storage, "access_key")
		form.SecretKey = service.StorageConfigValue(storage, "secret_key")
		form.SubmitURL = "/admin/settings/storage/" + storage.ID
		form.CancelURL = "/admin/settings/storage"
		return form
	}

	form.LocalRoot = service.StorageConfigValue(storage, "root")
	return form
}

func storageInputFromRequest(c *fiber.Ctx) service.StorageInput {
	return service.StorageInput{
		Name:      c.FormValue("name"),
		Endpoint:  c.FormValue("endpoint"),
		Region:    c.FormValue("region"),
		Bucket:    c.FormValue("bucket"),
		AccessKey: c.FormValue("access_key"),
		SecretKey: c.FormValue("secret_key"),
	}
}

func storageErrors(c *fiber.Ctx, err error) map[string]string {
	validationErr := &service.StorageValidationError{}
	if !errors.As(err, &validationErr) {
		return map[string]string{}
	}

	out := make(map[string]string, len(validationErr.FieldErrors))
	for field, fieldErr := range validationErr.FieldErrors {
		out[field] = translatedValidationError(c, fieldErr)
	}
	return out
}

func storageGlobalError(c *fiber.Ctx, err error) string {
	if _, ok := err.(*service.StorageValidationError); ok {
		return ""
	}
	return translatedValidationError(c, err)
}

func storageTypeLabel(c *fiber.Ctx, storageType string) string {
	switch storageType {
	case models.StorageTypeLocal:
		return translate(c, "settings.storage.type.local")
	case models.StorageTypeS3:
		return translate(c, "settings.storage.type.s3")
	default:
		return storageType
	}
}

func storageSummary(c *fiber.Ctx, storage models.Storage) string {
	switch storage.Type {
	case models.StorageTypeLocal:
		return service.StorageConfigValue(storage, "root")
	case models.StorageTypeS3:
		bucket := service.StorageConfigValue(storage, "bucket")
		endpoint := service.StorageConfigValue(storage, "endpoint")
		if bucket == "" {
			return endpoint
		}
		if endpoint == "" {
			return bucket
		}
		return bucket + " • " + endpoint
	default:
		return translate(c, "settings.storage.type.unknown")
	}
}
