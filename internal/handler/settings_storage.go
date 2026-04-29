package handler

import (
	"context"
	"errors"
	"fmt"
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
	IsDefault   bool
	Endpoint    string
	Region      string
	Bucket      string
	AccessKey   string
	SecretKey   string
	LocalRoot   string
	MediaCount  int64
	CreatedAt   string
	UpdatedAt   string
	IsReadonly  bool
	CanDelete   bool
	DeleteHint  string
	Errors      map[string]string
	GlobalError string
	TestResult  *storageTestResult
	SubmitURL   string
	CancelURL   string
	BackURL     string
	PushURL     string
}

type storageTestResult struct {
	Success bool
	Title   string
	Message string
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
			Type:       models.StorageTypeS3,
			TypeLabel:  translate(c, "settings.storage.type.s3"),
			SubmitURL:  "/admin/settings/storage",
			CancelURL:  "/admin/settings/storage",
			BackURL:    "/admin/settings/storage",
			PushURL:    "/admin/settings/storage/new",
			CanDelete:  false,
			DeleteHint: "",
			Errors:     map[string]string{},
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
	if err := h.persistDefaultStorageSelection(input.IsDefault, storage.ID, ""); err != nil {
		return fiber.ErrInternalServerError
	}

	testResult := h.runStorageTest(c, storage.ID)
	form := h.buildStorageFormData(c, storage, map[string]string{}, "")
	form.TestResult = &testResult

	c.Set("HX-Push-Url", "/admin/settings/storage/"+storage.ID)

	return h.renderStorage(c, storagePageData{
		List:         h.buildStorageList(c, storage.ID, ""),
		HasSelection: true,
		Storage:      form,
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
		form.IsDefault = input.IsDefault
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
	if err := h.persistDefaultStorageSelection(input.IsDefault, storage.ID, id); err != nil {
		return fiber.ErrInternalServerError
	}

	c.Set("HX-Push-Url", "/admin/settings/storage/"+storage.ID)

	return h.renderStorage(c, storagePageData{
		List:         h.buildStorageList(c, storage.ID, ""),
		HasSelection: true,
		Storage:      h.buildStorageFormData(c, storage, map[string]string{}, ""),
	})
}

func (h *settingsHandler) TestStorage(c *fiber.Ctx) error {
	id := c.Params("id")
	storage, err := h.storageService.GetByID(id)
	if err != nil {
		if errors.Is(err, service.ErrStorageNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return fiber.ErrInternalServerError
	}
	if storage.Type != models.StorageTypeS3 {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	form := h.buildStorageFormData(c, storage, map[string]string{}, "")
	testResult := h.runStorageTest(c, id)
	form.TestResult = &testResult

	return h.renderStorage(c, storagePageData{
		List:         h.buildStorageList(c, storage.ID, ""),
		HasSelection: true,
		Storage:      form,
	})
}

func (h *settingsHandler) DeleteStorage(c *fiber.Ctx) error {
	id := c.Params("id")
	storage, err := h.storageService.GetByID(id)
	if err != nil {
		if errors.Is(err, service.ErrStorageNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return fiber.ErrInternalServerError
	}

	deleteState := h.storageService.DeleteState(storage, h.optionService.GetDefaultStorageID())
	if !deleteState.CanDelete {
		return h.renderStorage(c, storagePageData{
			List:         h.buildStorageList(c, storage.ID, ""),
			HasSelection: true,
			Storage:      h.buildStorageFormData(c, storage, map[string]string{}, ""),
		})
	}

	if err := h.storageService.Delete(id); err != nil {
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
					translate(c, "settings.storage.delete_disabled.local"),
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
		IsDefault:   input.IsDefault,
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
	defaultStorageID := h.optionService.GetDefaultStorageID()
	items := make([]storageListItem, 0, len(storages))
	for _, storage := range storages {
		items = append(items, storageListItem{
			ID:        storage.ID,
			Name:      storage.Name,
			Type:      storage.Type,
			TypeLabel: storageTypeLabel(c, storage.Type),
			Summary:   storageSummary(c, storage),
			IsDefault: storage.ID == defaultStorageID,
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
	defaultStorageID := h.optionService.GetDefaultStorageID()
	deleteState := h.storageService.DeleteState(storage, defaultStorageID)

	form := storageFormData{
		ID:          storage.ID,
		Name:        storage.Name,
		Type:        storage.Type,
		TypeLabel:   storageTypeLabel(c, storage.Type),
		IsDefault:   storage.ID == defaultStorageID,
		MediaCount:  deleteState.MediaCount,
		CreatedAt:   storage.CreatedAt.Local().Format(time.DateTime),
		UpdatedAt:   storage.UpdatedAt.Local().Format(time.DateTime),
		IsReadonly:  storage.Type != models.StorageTypeS3,
		CanDelete:   deleteState.CanDelete,
		DeleteHint:  storageDeleteHint(c, deleteState),
		Errors:      errs,
		GlobalError: globalError,
		BackURL:     "/admin/settings/storage",
		PushURL:     "/admin/settings/storage/" + storage.ID,
	}

	if storage.Type == models.StorageTypeS3 {
		form.Endpoint = service.StorageConfigValue(storage, "endpoint")
		form.Region = service.StorageConfigValue(storage, "region")
		form.Bucket = service.StorageConfigValue(storage, "bucket")
		if credentials, err := h.storageService.S3Credentials(storage); err == nil {
			form.AccessKey = credentials.AccessKey
			form.SecretKey = credentials.SecretKey
		}
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
		IsDefault: c.FormValue("is_default") != "",
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

func storageDeleteHint(c *fiber.Ctx, state service.StorageDeleteState) string {
	if state.ReasonKey == "" {
		return ""
	}

	return translate(c, state.ReasonKey)
}

func (h *settingsHandler) persistDefaultStorageSelection(isDefault bool, storageID, previousStorageID string) error {
	currentDefaultID := h.optionService.GetDefaultStorageID()
	if isDefault {
		return h.optionService.SaveDefaultStorageID(storageID)
	}
	if currentDefaultID == storageID || previousStorageID == currentDefaultID {
		return h.optionService.SaveDefaultStorageID(service.DefaultLocalStorageID)
	}
	return nil
}

func (h *settingsHandler) runStorageTest(c *fiber.Ctx, id string) storageTestResult {
	ctx, cancel := context.WithTimeout(c.UserContext(), 20*time.Second)
	defer cancel()

	err := h.storageService.TestS3(ctx, id)
	if err == nil {
		return storageTestResult{
			Success: true,
			Title:   translate(c, "settings.storage.test.success_title"),
			Message: translate(c, "settings.storage.test.success_description"),
		}
	}

	return storageTestResult{
		Success: false,
		Title:   translate(c, "settings.storage.test.failure_title"),
		Message: storageTestErrorMessage(c, err),
	}
}

func storageTestErrorMessage(c *fiber.Ctx, err error) string {
	testErr := &service.StorageTestError{}
	if errors.As(err, &testErr) {
		return fmt.Sprintf(
			"%s: %s",
			storageTestStepLabel(c, testErr.Step),
			translatedValidationError(c, testErr.Err),
		)
	}

	return translatedValidationError(c, err)
}

func storageTestStepLabel(c *fiber.Ctx, step string) string {
	switch step {
	case "put":
		return translate(c, "settings.storage.test.step.put")
	case "get":
		return translate(c, "settings.storage.test.step.get")
	case "delete":
		return translate(c, "settings.storage.test.step.delete")
	default:
		return step
	}
}
