package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/umono-cms/umono/internal/media"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

const DefaultLocalStorageID = "local"

var (
	ErrMediaNotFound        = errors.New("media not found")
	ErrUnsupportedMediaType = errors.New("unsupported media type")
	ErrAliasAlreadyExists   = errors.New("alias already exists")
	ErrInvalidAlias         = errors.New("invalid alias")
	ErrPendingUploadMissing = errors.New("pending upload not found")
)

type countingReader struct {
	reader io.Reader
	count  int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.count += int64(n)
	return n, err
}

type UploadMediaInput struct {
	OriginalName string
	Alias        string
	MimeType     string
	Reader       io.Reader
}

type UploadMediaResult struct {
	Media     models.Media
	Duplicate *models.Media
	Pending   *PendingUpload
}

type PendingUpload struct {
	Token        string       `json:"token"`
	Media        models.Media `json:"media"`
	DuplicateID  string       `json:"duplicate_id"`
	DuplicateURL string       `json:"duplicate_url"`
}

type MediaService struct {
	repo           *repository.MediaRepository
	storageRepo    *repository.StorageRepository
	storageBackend media.Storage
	pendingDir     string
}

func NewMediaService(
	repo *repository.MediaRepository,
	storageRepo *repository.StorageRepository,
	storageBackend media.Storage,
	pendingDir string,
) *MediaService {
	return &MediaService{
		repo:           repo,
		storageRepo:    storageRepo,
		storageBackend: storageBackend,
		pendingDir:     pendingDir,
	}
}

func (s *MediaService) EnsureDefaultLocalStorage(root string) error {
	cfg := models.JSONMap{"root": root}

	storage := s.storageRepo.GetByID(DefaultLocalStorageID)
	if storage.ID == "" {
		s.storageRepo.Create(models.Storage{
			ID:     DefaultLocalStorageID,
			Name:   "Local Storage",
			Type:   models.StorageTypeLocal,
			Config: cfg,
		})
		return nil
	}

	if storage.Type != models.StorageTypeLocal || fmt.Sprint(storage.Config["root"]) != root {
		storage.Name = "Local Storage"
		storage.Type = models.StorageTypeLocal
		storage.Config = cfg
		s.storageRepo.Update(storage)
	}

	return nil
}

func (s *MediaService) GetAll() []models.Media {
	return s.repo.GetAll()
}

func (s *MediaService) GetByID(id string) (models.Media, error) {
	media := s.repo.GetByID(id)
	if media.ID == "" {
		return models.Media{}, ErrMediaNotFound
	}
	return media, nil
}

func (s *MediaService) Upload(ctx context.Context, input UploadMediaInput) (UploadMediaResult, error) {
	alias := media.NormalizeAlias(input.Alias)
	if alias != "" && !media.IsKebabCase(alias) {
		return UploadMediaResult{}, ErrInvalidAlias
	}
	if alias != "" && s.aliasExists(alias, "") {
		return UploadMediaResult{}, ErrAliasAlreadyExists
	}

	ext, ok := media.ExtensionByMimeType(input.MimeType)
	if !ok {
		return UploadMediaResult{}, ErrUnsupportedMediaType
	}

	id, err := uuid.NewV7()
	if err != nil {
		return UploadMediaResult{}, err
	}

	pathKey := id.String() + "." + ext
	hasher := sha256.New()
	counter := &countingReader{reader: input.Reader}
	reader := io.TeeReader(counter, hasher)

	if err := s.storageBackend.Put(ctx, pathKey, reader, media.ObjectMeta{
		ContentType: input.MimeType,
	}); err != nil {
		return UploadMediaResult{}, err
	}

	record := models.Media{
		ID:           id.String(),
		StorageID:    DefaultLocalStorageID,
		OriginalName: strings.TrimSpace(input.OriginalName),
		PathKey:      pathKey,
		MimeType:     input.MimeType,
		Size:         counter.count,
		Hash:         hex.EncodeToString(hasher.Sum(nil)),
		Metadata:     buildMediaMetadata(alias),
	}

	if width, height, err := s.readDimensions(ctx, record.PathKey, record.MimeType); err == nil {
		record.Metadata["width"] = width
		record.Metadata["height"] = height
	}

	existing := s.repo.GetByHash(record.Hash)
	if existing.ID != "" {
		pending, err := s.savePendingUpload(record, existing.ID)
		if err != nil {
			_ = s.storageBackend.Delete(ctx, record.PathKey)
			return UploadMediaResult{}, err
		}

		existingURL, _ := s.PublicURL(existing)
		pending.DuplicateURL = existingURL

		return UploadMediaResult{
			Duplicate: &existing,
			Pending:   pending,
		}, nil
	}

	record = s.repo.Create(record)
	return UploadMediaResult{Media: record}, nil
}

func (s *MediaService) ConfirmPendingUpload(ctx context.Context, token string) (models.Media, error) {
	pending, err := s.loadPendingUpload(token)
	if err != nil {
		return models.Media{}, err
	}

	reader, _, storageErr := s.storageBackend.Get(ctx, pending.Media.PathKey)
	if storageErr == nil {
		reader.Close()
	}
	if storageErr != nil {
		_ = s.deletePendingUpload(token)
		return models.Media{}, ErrPendingUploadMissing
	}

	alias := MediaAlias(pending.Media)
	if alias != "" && !media.IsKebabCase(alias) {
		return models.Media{}, ErrInvalidAlias
	}
	if alias != "" && s.aliasExists(alias, "") {
		return models.Media{}, ErrAliasAlreadyExists
	}

	created := s.repo.Create(pending.Media)
	_ = s.deletePendingUpload(token)
	return created, nil
}

func (s *MediaService) OpenPendingUpload(ctx context.Context, token string) (io.ReadCloser, media.ObjectMeta, error) {
	pending, err := s.loadPendingUpload(token)
	if err != nil {
		return nil, media.ObjectMeta{}, err
	}

	reader, meta, err := s.storageBackend.Get(ctx, pending.Media.PathKey)
	if err != nil {
		return nil, media.ObjectMeta{}, ErrPendingUploadMissing
	}

	meta.ContentType = pending.Media.MimeType
	meta.Size = pending.Media.Size
	return reader, meta, nil
}

func (s *MediaService) CancelPendingUpload(ctx context.Context, token string) error {
	pending, err := s.loadPendingUpload(token)
	if err != nil {
		return err
	}

	if err := s.storageBackend.Delete(ctx, pending.Media.PathKey); err != nil {
		return err
	}

	return s.deletePendingUpload(token)
}

func (s *MediaService) UpdateAlias(id, alias string) (models.Media, error) {
	item, err := s.GetByID(id)
	if err != nil {
		return models.Media{}, err
	}

	alias = media.NormalizeAlias(alias)
	if alias != "" && !media.IsKebabCase(alias) {
		return models.Media{}, ErrInvalidAlias
	}
	if alias != "" && s.aliasExists(alias, id) {
		return models.Media{}, ErrAliasAlreadyExists
	}

	metadata := buildMediaMetadata(alias)
	copyDimensionMetadata(metadata, item.Metadata)
	item.Metadata = metadata
	item = s.repo.Update(item)
	return item, nil
}

func (s *MediaService) Delete(ctx context.Context, id string) error {
	item, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.storageBackend.Delete(ctx, item.PathKey); err != nil {
		return err
	}

	return s.repo.Delete(id)
}

func (s *MediaService) OpenByIDAndExt(ctx context.Context, id, ext string) (io.ReadCloser, media.ObjectMeta, error) {
	item, err := s.GetByID(id)
	if err != nil {
		return nil, media.ObjectMeta{}, err
	}

	expectedExt, ok := media.ExtensionByMimeType(item.MimeType)
	if !ok || !strings.EqualFold(strings.TrimSpace(ext), expectedExt) {
		return nil, media.ObjectMeta{}, ErrMediaNotFound
	}

	reader, meta, err := s.storageBackend.Get(ctx, item.PathKey)
	if err != nil {
		return nil, media.ObjectMeta{}, err
	}

	meta.ContentType = item.MimeType
	meta.Size = item.Size
	return reader, meta, nil
}

func (s *MediaService) PublicURL(item models.Media) (string, error) {
	return s.storageBackend.PublicURL(context.Background(), item.PathKey)
}

func MediaAlias(item models.Media) string {
	if item.Metadata == nil {
		return ""
	}

	raw, ok := item.Metadata["alias"]
	if !ok {
		return ""
	}

	alias, _ := raw.(string)
	return media.NormalizeAlias(alias)
}

func buildMediaMetadata(alias string) models.JSONMap {
	metadata := models.JSONMap{}
	if alias != "" {
		metadata["alias"] = alias
	}
	return metadata
}

func copyDimensionMetadata(dst, src models.JSONMap) {
	if src == nil {
		return
	}

	if width, ok := src["width"]; ok {
		dst["width"] = width
	}
	if height, ok := src["height"]; ok {
		dst["height"] = height
	}
}

func (s *MediaService) aliasExists(alias, excludeID string) bool {
	found := s.repo.GetByAlias(alias)
	return found.ID != "" && found.ID != excludeID
}

func (s *MediaService) readDimensions(ctx context.Context, pathKey, mimeType string) (int, int, error) {
	reader, _, err := s.storageBackend.Get(ctx, pathKey)
	if err != nil {
		return 0, 0, err
	}
	defer reader.Close()

	return media.DimensionsFromReader(mimeType, reader)
}

func (s *MediaService) savePendingUpload(item models.Media, duplicateID string) (*PendingUpload, error) {
	token, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	pending := &PendingUpload{
		Token:       token.String(),
		Media:       item,
		DuplicateID: duplicateID,
	}

	if err := os.MkdirAll(s.pendingDir, 0o755); err != nil {
		return nil, err
	}

	data, err := json.Marshal(pending)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(s.pendingPath(token.String()), data, 0o644); err != nil {
		return nil, err
	}

	return pending, nil
}

func (s *MediaService) loadPendingUpload(token string) (PendingUpload, error) {
	data, err := os.ReadFile(s.pendingPath(token))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return PendingUpload{}, ErrPendingUploadMissing
		}
		return PendingUpload{}, err
	}

	var pending PendingUpload
	if err := json.Unmarshal(data, &pending); err != nil {
		return PendingUpload{}, err
	}

	return pending, nil
}

func (s *MediaService) deletePendingUpload(token string) error {
	err := os.Remove(s.pendingPath(token))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *MediaService) pendingPath(token string) string {
	return filepath.Join(s.pendingDir, token+".json")
}
