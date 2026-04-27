package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/media"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
	"github.com/umono-cms/umono/internal/view"
)

type mediaHandler struct {
	mediaService   *service.MediaService
	storageService *service.StorageService
	optionService  *service.OptionService
}

const mediaCacheControl = "public, max-age=31536000, immutable"

func NewMediaHandler(ms *service.MediaService, ss *service.StorageService, os *service.OptionService) *mediaHandler {
	return &mediaHandler{
		mediaService:   ms,
		storageService: ss,
		optionService:  os,
	}
}

func (h *mediaHandler) Index(c *fiber.Ctx) error {
	items := h.mediaService.GetAll()

	return h.render(c, "partials/media-index", fiber.Map{
		"MediaList": h.buildMediaList(items, ""),
	})
}

func (h *mediaHandler) RenderNew(c *fiber.Ctx) error {
	items := h.mediaService.GetAll()

	return h.render(c, "partials/media-new", fiber.Map{
		"MediaList":   h.buildMediaList(items, ""),
		"MediaUpload": h.buildUploadForm(c, mediaUploadFormData{}),
	})
}

func (h *mediaHandler) RenderShow(c *fiber.Ctx) error {
	item, err := h.mediaService.GetByID(c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	items := h.mediaService.GetAll()

	return h.render(c, "partials/media-show", fiber.Map{
		"MediaList": h.buildMediaList(items, item.ID),
		"Media":     h.buildMediaDetail(item, c.Locals("I18n")),
	})
}

func (h *mediaHandler) Upload(c *fiber.Ctx) error {
	storageID := strings.TrimSpace(c.FormValue("storage_id"))
	selectedStorage, err := h.storageService.GetByID(firstNonEmpty(storageID, service.DefaultLocalStorageID))
	if err != nil {
		if wantsMediaUploadJSON(c) {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": translate(c, "media.errors.invalid_storage"),
			})
		}
		return Render(c, "partials/media-new-content", fiber.Map{
			"MediaUpload": h.buildUploadForm(c, mediaUploadFormData{
				Alias:           strings.TrimSpace(c.FormValue("alias")),
				StorageID:       storageID,
				SelectedStorage: storageID,
				ErrorMsg:        translate(c, "media.errors.invalid_storage"),
			}),
		})
	}
	if selectedStorage.Type != models.StorageTypeLocal {
		if wantsMediaUploadJSON(c) {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": translate(c, "media.errors.direct_upload_requires_js"),
			})
		}
		return Render(c, "partials/media-new-content", fiber.Map{
			"MediaUpload": h.buildUploadForm(c, mediaUploadFormData{
				Alias:           strings.TrimSpace(c.FormValue("alias")),
				StorageID:       storageID,
				SelectedStorage: storageID,
				ErrorMsg:        translate(c, "media.errors.direct_upload_requires_js"),
			}),
		})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		if wantsMediaUploadJSON(c) {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": translate(c, "media.errors.file_required"),
			})
		}
		return Render(c, "partials/media-new-content", fiber.Map{
			"MediaUpload": h.buildUploadForm(c, mediaUploadFormData{
				Alias:           strings.TrimSpace(c.FormValue("alias")),
				StorageID:       storageID,
				SelectedStorage: storageID,
				ErrorMsg:        translate(c, "media.errors.file_required"),
			}),
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return fiber.ErrInternalServerError
	}
	defer file.Close()

	reader, mimeType, err := sniffMedia(file)
	if err != nil {
		return fiber.ErrBadRequest
	}

	if !media.AllowedMimeType(mimeType) {
		if wantsMediaUploadJSON(c) {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": translate(c, "media.errors.unsupported_type"),
			})
		}
		return Render(c, "partials/media-new-content", fiber.Map{
			"MediaUpload": h.buildUploadForm(c, mediaUploadFormData{
				Alias:           strings.TrimSpace(c.FormValue("alias")),
				StorageID:       storageID,
				SelectedStorage: storageID,
				ErrorMsg:        translate(c, "media.errors.unsupported_type"),
			}),
		})
	}

	result, err := h.mediaService.Upload(c.UserContext(), service.UploadMediaInput{
		StorageID:    storageID,
		OriginalName: fileHeader.Filename,
		Alias:        c.FormValue("alias"),
		MimeType:     mimeType,
		Reader:       reader,
	})
	if err != nil {
		errMsg := translate(c, "media.errors.upload_failed")
		aliasError := false
		if errors.Is(err, service.ErrAliasAlreadyExists) {
			errMsg = translate(c, "media.errors.alias_exists")
			aliasError = true
		} else if errors.Is(err, service.ErrInvalidAlias) {
			errMsg = translate(c, "media.errors.invalid_alias")
			aliasError = true
		} else if errors.Is(err, service.ErrUnsupportedMediaType) {
			errMsg = translate(c, "media.errors.unsupported_type")
		}

		if wantsMediaUploadJSON(c) {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error":       errMsg,
				"alias_error": aliasError,
			})
		}

		return Render(c, "partials/media-new-content", fiber.Map{
			"MediaUpload": h.buildUploadForm(c, mediaUploadFormData{
				Alias:           strings.TrimSpace(c.FormValue("alias")),
				StorageID:       storageID,
				ErrorMsg:        errMsg,
				AliasError:      aliasError,
				SelectedStorage: storageID,
			}),
		})
	}

	if result.Pending != nil && result.Duplicate != nil {
		return Render(c, "partials/media-confirm-content", fiber.Map{
			"Pending": buildPendingDuplicate(result.Pending, *result.Duplicate),
		})
	}

	items := h.mediaService.GetAll()
	c.Set("HX-Push-Url", "/admin/media")
	return h.render(c, "partials/media-index", fiber.Map{
		"MediaList": h.buildMediaList(items, ""),
	})
}

func (h *mediaHandler) PresignUpload(c *fiber.Ctx) error {
	input := service.PrepareUploadInput{
		StorageID:    c.FormValue("storage_id"),
		OriginalName: strings.TrimSpace(c.FormValue("original_name")),
		Alias:        c.FormValue("alias"),
		MimeType:     strings.TrimSpace(c.FormValue("mime_type")),
		Hash:         strings.TrimSpace(c.FormValue("hash")),
	}

	size, _ := strconv.ParseInt(strings.TrimSpace(c.FormValue("size")), 10, 64)
	input.Size = size

	result, err := h.mediaService.PrepareUpload(c.UserContext(), input)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error":       mediaUploadErrorMessage(c, err),
			"alias_error": errors.Is(err, service.ErrAliasAlreadyExists) || errors.Is(err, service.ErrInvalidAlias),
		})
	}

	return c.JSON(result)
}

func (h *mediaHandler) CompleteUpload(c *fiber.Ctx) error {
	result, err := h.mediaService.CompletePreparedUpload(c.UserContext(), c.FormValue("token"))
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error":       mediaUploadErrorMessage(c, err),
			"alias_error": errors.Is(err, service.ErrAliasAlreadyExists) || errors.Is(err, service.ErrInvalidAlias),
		})
	}

	if result.Pending != nil && result.Duplicate != nil {
		return Render(c, "partials/media-confirm-content", fiber.Map{
			"Pending": buildPendingDuplicate(result.Pending, *result.Duplicate),
		})
	}

	items := h.mediaService.GetAll()
	c.Set("HX-Push-Url", "/admin/media")
	return h.render(c, "partials/media-index", fiber.Map{
		"MediaList": h.buildMediaList(items, ""),
	})
}

func (h *mediaHandler) ConfirmUpload(c *fiber.Ctx) error {
	_, err := h.mediaService.ConfirmPendingUpload(c.UserContext(), c.FormValue("token"))
	if err != nil {
		errMsg := translate(c, "media.errors.pending_complete_failed")
		aliasError := false
		if errors.Is(err, service.ErrAliasAlreadyExists) {
			errMsg = translate(c, "media.errors.alias_exists")
			aliasError = true
		} else if errors.Is(err, service.ErrInvalidAlias) {
			errMsg = translate(c, "media.errors.invalid_alias")
			aliasError = true
		}

		return Render(c, "partials/media-new-content", fiber.Map{
			"MediaUpload": h.buildUploadForm(c, mediaUploadFormData{
				ErrorMsg:   errMsg,
				AliasError: aliasError,
			}),
		})
	}

	items := h.mediaService.GetAll()
	c.Set("HX-Push-Url", "/admin/media")
	return h.render(c, "partials/media-index", fiber.Map{
		"MediaList": h.buildMediaList(items, ""),
	})
}

func (h *mediaHandler) CancelUpload(c *fiber.Ctx) error {
	if err := h.mediaService.CancelPendingUpload(c.UserContext(), c.FormValue("token")); err != nil {
		return Render(c, "partials/media-new-content", fiber.Map{
			"MediaUpload": h.buildUploadForm(c, mediaUploadFormData{
				ErrorMsg: translate(c, "media.errors.pending_cancel_failed"),
			}),
		})
	}

	c.Set("HX-Push-Url", "/admin/media/new")
	return Render(c, "partials/media-new-content", fiber.Map{
		"MediaUpload": h.buildUploadForm(c, mediaUploadFormData{}),
	})
}

func (h *mediaHandler) UpdateAlias(c *fiber.Ctx) error {
	submittedAlias := strings.TrimSpace(c.FormValue("alias"))
	item, err := h.mediaService.UpdateAlias(c.Params("id"), submittedAlias)
	if err != nil {
		if errors.Is(err, service.ErrMediaNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}

		current, getErr := h.mediaService.GetByID(c.Params("id"))
		if getErr != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}

		detail := h.buildMediaDetail(current, c.Locals("I18n"))
		detail.Alias = submittedAlias
		if errors.Is(err, service.ErrAliasAlreadyExists) {
			detail.ErrorMsg = translate(c, "media.errors.alias_exists")
		} else if errors.Is(err, service.ErrInvalidAlias) {
			detail.ErrorMsg = translate(c, "media.errors.invalid_alias")
		} else {
			detail.ErrorMsg = translate(c, "media.errors.alias_update_failed")
		}

		return Render(c, "partials/media-show-content", fiber.Map{
			"Media": detail,
		})
	}

	items := h.mediaService.GetAll()
	return h.render(c, "partials/media-show", fiber.Map{
		"MediaList": h.buildMediaList(items, item.ID),
		"Media":     h.buildMediaDetail(item, c.Locals("I18n")),
	})
}

func (h *mediaHandler) Delete(c *fiber.Ctx) error {
	if err := h.mediaService.Delete(c.UserContext(), c.Params("id")); err != nil {
		if errors.Is(err, service.ErrMediaNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return fiber.ErrInternalServerError
	}

	items := h.mediaService.GetAll()
	c.Set("HX-Push-Url", "/admin/media")
	return h.render(c, "partials/media-index", fiber.Map{
		"MediaList": h.buildMediaList(items, ""),
	})
}

func (h *mediaHandler) ServePending(c *fiber.Ctx) error {
	reader, meta, err := h.mediaService.OpenPendingUpload(c.UserContext(), c.Params("token"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	c.Set(fiber.HeaderContentType, meta.ContentType)
	if meta.Size > 0 {
		c.Set(fiber.HeaderContentLength, strconv.FormatInt(meta.Size, 10))
	}

	return c.SendStream(reader, int(meta.Size))
}

func (h *mediaHandler) Serve(c *fiber.Ctx) error {
	filename := c.Params("filename")
	ext := strings.TrimPrefix(filepath.Ext(filename), ".")
	id := strings.TrimSuffix(filename, "."+ext)
	if id == "" || ext == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	item, err := h.mediaService.GetByID(id)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	reader, meta, err := h.mediaService.OpenByIDAndExt(c.UserContext(), id, ext)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	storage, err := h.storageService.GetByID(item.StorageID)
	if err == nil && storage.Type == models.StorageTypeLocal {
		c.Set(fiber.HeaderCacheControl, mediaCacheControl)
	}

	c.Type(ext)
	c.Set(fiber.HeaderContentType, meta.ContentType)
	if meta.Size > 0 {
		c.Set(fiber.HeaderContentLength, strconv.FormatInt(meta.Size, 10))
	}

	return c.SendStream(reader, int(meta.Size))
}

type mediaDetailData struct {
	ID          string
	Name        string
	Alias       string
	URL         string
	StorageName string
	ContentType string
	Size        int64
	Width       int
	Height      int
	CreatedAt   string
	ErrorMsg    string
}

type mediaStorageOption struct {
	ID         string
	Name       string
	Type       string
	TypeLabel  string
	IsDefault  bool
	IsSelected bool
}

type mediaUploadFormData struct {
	Alias           string
	StorageID       string
	SelectedStorage string
	ErrorMsg        string
	AliasError      bool
	Storages        []mediaStorageOption
}

func (h *mediaHandler) buildMediaDetail(item models.Media, translator any) mediaDetailData {
	storageName := item.StorageID
	if storage, err := h.storageService.GetByID(item.StorageID); err == nil && strings.TrimSpace(storage.Name) != "" {
		storageName = storage.Name
	}

	return mediaDetailData{
		ID:          item.ID,
		Name:        item.OriginalName,
		Alias:       service.MediaAlias(item),
		URL:         h.directURL(item),
		StorageName: storageName,
		ContentType: item.MimeType,
		Size:        item.Size,
		Width:       metadataInt(item.Metadata, "width"),
		Height:      metadataInt(item.Metadata, "height"),
		CreatedAt:   view.RelativeTimeWithTranslator(&item.CreatedAt, translator),
	}
}

func metadataInt(metadata models.JSONMap, key string) int {
	if metadata == nil {
		return 0
	}

	value, ok := metadata[key]
	if !ok {
		return 0
	}

	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

type pendingDuplicateData struct {
	Token        string
	NewURL       string
	DuplicateURL string
	Name         string
	Alias        string
}

func buildPendingDuplicate(pending *service.PendingUpload, duplicate models.Media) pendingDuplicateData {
	return pendingDuplicateData{
		Token:        pending.Token,
		NewURL:       "/admin/media/pending/" + pending.Token + "/preview",
		DuplicateURL: pending.DuplicateURL,
		Name:         pending.Media.OriginalName,
		Alias:        service.MediaAlias(pending.Media),
	}
}

func (h *mediaHandler) buildUploadForm(c *fiber.Ctx, form mediaUploadFormData) mediaUploadFormData {
	defaultStorageID := h.optionService.GetDefaultStorageID()
	form.Storages = h.buildStorageOptions(c, defaultStorageID, firstNonEmpty(form.SelectedStorage, form.StorageID, defaultStorageID))
	if len(form.Storages) > 0 && firstNonEmpty(form.SelectedStorage, form.StorageID) == "" {
		form.StorageID = form.Storages[0].ID
		form.SelectedStorage = form.Storages[0].ID
	}
	return form
}

func (h *mediaHandler) buildStorageOptions(c *fiber.Ctx, defaultStorageID, selectedStorageID string) []mediaStorageOption {
	storages := h.storageService.GetAll()
	options := make([]mediaStorageOption, 0, len(storages))

	appendStorage := func(storage models.Storage) {
		options = append(options, mediaStorageOption{
			ID:         storage.ID,
			Name:       storage.Name,
			Type:       storage.Type,
			TypeLabel:  storageTypeLabel(c, storage.Type),
			IsDefault:  storage.ID == defaultStorageID,
			IsSelected: storage.ID == selectedStorageID,
		})
	}

	for _, storage := range storages {
		if storage.ID == defaultStorageID {
			appendStorage(storage)
			break
		}
	}

	for _, storage := range storages {
		if storage.ID == defaultStorageID {
			continue
		}
		appendStorage(storage)
	}

	if len(options) > 0 {
		hasSelected := false
		for _, option := range options {
			if option.IsSelected {
				hasSelected = true
				break
			}
		}
		if !hasSelected {
			options[0].IsSelected = true
		}
	}

	return options
}

func (h *mediaHandler) buildMediaList(items []models.Media, activeID string) []view.MediaListItem {
	return view.MediaList(items, activeID, h.directURL)
}

func (h *mediaHandler) directURL(item models.Media) string {
	url, err := h.mediaService.DirectURL(context.Background(), item)
	if err != nil {
		url, err = h.mediaService.PublicURL(item)
		if err != nil {
			return ""
		}
	}

	return url
}

func mediaUploadErrorMessage(c *fiber.Ctx, err error) string {
	switch {
	case errors.Is(err, service.ErrAliasAlreadyExists):
		return translate(c, "media.errors.alias_exists")
	case errors.Is(err, service.ErrInvalidAlias):
		return translate(c, "media.errors.invalid_alias")
	case errors.Is(err, service.ErrUnsupportedMediaType):
		return translate(c, "media.errors.unsupported_type")
	case errors.Is(err, service.ErrStorageNotFound):
		return translate(c, "media.errors.invalid_storage")
	case errors.Is(err, service.ErrPendingUploadMissing):
		return translate(c, "media.errors.pending_missing")
	default:
		return translate(c, "media.errors.upload_failed")
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func sniffMedia(file io.Reader) (io.Reader, string, error) {
	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, "", err
	}

	detected := http.DetectContentType(head[:n])
	return io.MultiReader(bytes.NewReader(head[:n]), file), detected, nil
}

func (h *mediaHandler) render(c *fiber.Ctx, partial string, data fiber.Map) error {
	layouts := []string{"layouts/media", "layouts/admin"}
	if isMediaContentSwap(c) {
		partial = "partials/htmx/" + strings.TrimPrefix(partial, "partials/")
		layouts = []string{}
	}

	return Render(c, partial, data, layouts...)
}

func isMediaContentSwap(c *fiber.Ctx) bool {
	return c.Get("HX-Request") == "true" && c.Get("HX-Target") == "media-content"
}

func wantsMediaUploadJSON(c *fiber.Ctx) bool {
	return c.Get("X-Umono-Media-Upload") == "true"
}
