package handler

import (
	"bytes"
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
	mediaService *service.MediaService
}

func NewMediaHandler(ms *service.MediaService) *mediaHandler {
	return &mediaHandler{mediaService: ms}
}

func (h *mediaHandler) Index(c *fiber.Ctx) error {
	items := h.mediaService.GetAll()

	return Render(c, "partials/media-index", fiber.Map{
		"MediaList": view.MediaList(items, ""),
	}, "layouts/media", "layouts/admin")
}

func (h *mediaHandler) RenderNew(c *fiber.Ctx) error {
	items := h.mediaService.GetAll()

	return Render(c, "partials/media-new", fiber.Map{
		"MediaList": view.MediaList(items, ""),
	}, "layouts/media", "layouts/admin")
}

func (h *mediaHandler) RenderShow(c *fiber.Ctx) error {
	item, err := h.mediaService.GetByID(c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	items := h.mediaService.GetAll()

	return Render(c, "partials/media-show", fiber.Map{
		"MediaList": view.MediaList(items, item.ID),
		"Media":     buildMediaDetail(item),
	}, "layouts/media", "layouts/admin")
}

func (h *mediaHandler) Upload(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return Render(c, "partials/media-new-content", fiber.Map{
			"Alias":    strings.TrimSpace(c.FormValue("alias")),
			"ErrorMsg": "Select a PNG, JPEG, or WEBP image to continue.",
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
		return Render(c, "partials/media-new-content", fiber.Map{
			"Alias":    strings.TrimSpace(c.FormValue("alias")),
			"ErrorMsg": "Only PNG, JPEG, and WEBP files are allowed.",
		})
	}

	result, err := h.mediaService.Upload(c.UserContext(), service.UploadMediaInput{
		OriginalName: fileHeader.Filename,
		Alias:        c.FormValue("alias"),
		MimeType:     mimeType,
		Reader:       reader,
	})
	if err != nil {
		errMsg := "Upload failed."
		if errors.Is(err, service.ErrAliasAlreadyExists) {
			errMsg = "This alias is already in use."
		} else if errors.Is(err, service.ErrUnsupportedMediaType) {
			errMsg = "Only PNG, JPEG, and WEBP files are allowed."
		}

		return Render(c, "partials/media-new-content", fiber.Map{
			"Alias":    strings.TrimSpace(c.FormValue("alias")),
			"ErrorMsg": errMsg,
		})
	}

	if result.Pending != nil && result.Duplicate != nil {
		return Render(c, "partials/media-confirm-content", fiber.Map{
			"Pending": buildPendingDuplicate(result.Pending, *result.Duplicate),
		})
	}

	c.Set("HX-Redirect", "/admin/media")
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *mediaHandler) ConfirmUpload(c *fiber.Ctx) error {
	_, err := h.mediaService.ConfirmPendingUpload(c.UserContext(), c.FormValue("token"))
	if err != nil {
		errMsg := "The pending upload could not be completed."
		if errors.Is(err, service.ErrAliasAlreadyExists) {
			errMsg = "This alias is already in use."
		}

		return Render(c, "partials/media-new-content", fiber.Map{
			"ErrorMsg": errMsg,
		})
	}

	c.Set("HX-Redirect", "/admin/media")
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *mediaHandler) CancelUpload(c *fiber.Ctx) error {
	if err := h.mediaService.CancelPendingUpload(c.UserContext(), c.FormValue("token")); err != nil {
		return Render(c, "partials/media-new-content", fiber.Map{
			"ErrorMsg": "The pending upload could not be canceled cleanly.",
		})
	}

	return Render(c, "partials/media-new-content", fiber.Map{})
}

func (h *mediaHandler) UpdateAlias(c *fiber.Ctx) error {
	item, err := h.mediaService.UpdateAlias(c.Params("id"), c.FormValue("alias"))
	if err != nil {
		if errors.Is(err, service.ErrMediaNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}

		current, getErr := h.mediaService.GetByID(c.Params("id"))
		if getErr != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}

		detail := buildMediaDetail(current)
		if errors.Is(err, service.ErrAliasAlreadyExists) {
			detail.ErrorMsg = "This alias is already in use."
		} else {
			detail.ErrorMsg = "Alias update failed."
		}

		return Render(c, "partials/media-show-content", fiber.Map{
			"Media": detail,
		})
	}

	return Render(c, "partials/media-show-content", fiber.Map{
		"Media": buildMediaDetail(item),
	})
}

func (h *mediaHandler) Delete(c *fiber.Ctx) error {
	if err := h.mediaService.Delete(c.UserContext(), c.Params("id")); err != nil {
		if errors.Is(err, service.ErrMediaNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return fiber.ErrInternalServerError
	}

	c.Set("HX-Redirect", "/admin/media")
	return c.SendStatus(fiber.StatusNoContent)
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

	reader, meta, err := h.mediaService.OpenByIDAndExt(c.UserContext(), id, ext)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
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
	ContentType string
	Size        int64
	Width       int
	Height      int
	CreatedAt   string
	ErrorMsg    string
}

func buildMediaDetail(item models.Media) mediaDetailData {
	ext, _ := media.ExtensionByMimeType(item.MimeType)
	return mediaDetailData{
		ID:          item.ID,
		Name:        item.OriginalName,
		Alias:       service.MediaAlias(item),
		URL:         "/uploads/" + item.ID + "." + ext,
		ContentType: item.MimeType,
		Size:        item.Size,
		Width:       metadataInt(item.Metadata, "width"),
		Height:      metadataInt(item.Metadata, "height"),
		CreatedAt:   item.CreatedAt.Format("2006-01-02 15:04"),
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

func sniffMedia(file io.Reader) (io.Reader, string, error) {
	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, "", err
	}

	detected := http.DetectContentType(head[:n])
	return io.MultiReader(bytes.NewReader(head[:n]), file), detected, nil
}
