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

const mediaCacheControl = "public, max-age=31536000, immutable"

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
	StorageID    string
	OriginalName string
	Alias        string
	MimeType     string
	Reader       io.Reader
}

type PrepareUploadInput struct {
	StorageID    string
	OriginalName string
	Alias        string
	MimeType     string
	Size         int64
	Hash         string
}

type PrepareUploadResult struct {
	Token   string            `json:"token"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
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
	repo        *repository.MediaRepository
	storageRepo *repository.StorageRepository
	optionRepo  *repository.OptionRepository
	secrets     *SecretService
	pendingDir  string
}

func NewMediaService(
	repo *repository.MediaRepository,
	storageRepo *repository.StorageRepository,
	optionRepo *repository.OptionRepository,
	secrets *SecretService,
	pendingDir string,
) *MediaService {
	return &MediaService{
		repo:        repo,
		storageRepo: storageRepo,
		optionRepo:  optionRepo,
		secrets:     secrets,
		pendingDir:  pendingDir,
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

	pathKey := mediaPathKey(id.String(), ext)
	hasher := sha256.New()
	counter := &countingReader{reader: input.Reader}
	reader := io.TeeReader(counter, hasher)
	storageModel, storageBackend, err := s.storageByID(ctx, input.StorageID)
	if err != nil {
		return UploadMediaResult{}, err
	}

	if err := storageBackend.Put(ctx, pathKey, reader, media.ObjectMeta{
		ContentType:  input.MimeType,
		CacheControl: mediaCacheControl,
	}); err != nil {
		return UploadMediaResult{}, err
	}

	record := models.Media{
		ID:           id.String(),
		StorageID:    storageModel.ID,
		OriginalName: strings.TrimSpace(input.OriginalName),
		PathKey:      pathKey,
		MimeType:     input.MimeType,
		Size:         counter.count,
		Hash:         hex.EncodeToString(hasher.Sum(nil)),
		Metadata:     buildMediaMetadata(alias),
	}

	if width, height, err := s.readDimensions(ctx, storageBackend, record.PathKey, record.MimeType); err == nil {
		record.Metadata["width"] = width
		record.Metadata["height"] = height
	}

	existing := s.repo.GetByHash(record.Hash)
	if existing.ID != "" {
		pending, err := s.savePendingUpload(record, existing.ID)
		if err != nil {
			_ = storageBackend.Delete(ctx, record.PathKey)
			return UploadMediaResult{}, err
		}

		existingURL, _ := s.DirectURL(ctx, existing)
		pending.DuplicateURL = existingURL

		return UploadMediaResult{
			Duplicate: &existing,
			Pending:   pending,
		}, nil
	}

	record = s.repo.Create(record)
	return UploadMediaResult{Media: record}, nil
}

func (s *MediaService) PrepareUpload(ctx context.Context, input PrepareUploadInput) (PrepareUploadResult, error) {
	alias := media.NormalizeAlias(input.Alias)
	if alias != "" && !media.IsKebabCase(alias) {
		return PrepareUploadResult{}, ErrInvalidAlias
	}
	if alias != "" && s.aliasExists(alias, "") {
		return PrepareUploadResult{}, ErrAliasAlreadyExists
	}

	ext, ok := media.ExtensionByMimeType(input.MimeType)
	if !ok {
		return PrepareUploadResult{}, ErrUnsupportedMediaType
	}

	storageModel, storageBackend, err := s.storageByID(ctx, input.StorageID)
	if err != nil {
		return PrepareUploadResult{}, err
	}
	if storageModel.Type != models.StorageTypeS3 {
		return PrepareUploadResult{}, media.ErrPresignUnsupported
	}

	id, err := uuid.NewV7()
	if err != nil {
		return PrepareUploadResult{}, err
	}

	record := models.Media{
		ID:           id.String(),
		StorageID:    storageModel.ID,
		OriginalName: strings.TrimSpace(input.OriginalName),
		PathKey:      mediaPathKey(id.String(), ext),
		MimeType:     input.MimeType,
		Size:         input.Size,
		Hash:         strings.TrimSpace(strings.ToLower(input.Hash)),
		Metadata:     buildMediaMetadata(alias),
	}

	pending, err := s.savePendingUpload(record, "")
	if err != nil {
		return PrepareUploadResult{}, err
	}

	url, headers, err := storageBackend.PresignPut(ctx, record.PathKey, media.ObjectMeta{
		ContentType:  record.MimeType,
		CacheControl: mediaCacheControl,
		Size:         record.Size,
	})
	if err != nil {
		_ = s.deletePendingUpload(pending.Token)
		return PrepareUploadResult{}, err
	}

	return PrepareUploadResult{
		Token:   pending.Token,
		URL:     url,
		Headers: headers,
	}, nil
}

func (s *MediaService) CompletePreparedUpload(ctx context.Context, token string) (UploadMediaResult, error) {
	pending, err := s.loadPendingUpload(token)
	if err != nil {
		return UploadMediaResult{}, err
	}

	_, storageBackend, err := s.storageByID(ctx, pending.Media.StorageID)
	if err != nil {
		_ = s.deletePendingUpload(token)
		return UploadMediaResult{}, ErrPendingUploadMissing
	}

	reader, meta, err := storageBackend.Get(ctx, pending.Media.PathKey)
	if err != nil {
		_ = s.deletePendingUpload(token)
		return UploadMediaResult{}, ErrPendingUploadMissing
	}
	reader.Close()

	if meta.Size > 0 {
		pending.Media.Size = meta.Size
	}
	if pending.Media.MimeType == "" && meta.ContentType != "" {
		pending.Media.MimeType = meta.ContentType
	}

	if width, height, err := s.readDimensions(ctx, storageBackend, pending.Media.PathKey, pending.Media.MimeType); err == nil {
		pending.Media.Metadata["width"] = width
		pending.Media.Metadata["height"] = height
	}

	existing := s.repo.GetByHash(pending.Media.Hash)
	if existing.ID != "" {
		pending.DuplicateID = existing.ID
		existingURL, _ := s.DirectURL(ctx, existing)
		pending.DuplicateURL = existingURL
		if err := s.writePendingUpload(pending); err != nil {
			return UploadMediaResult{}, err
		}

		return UploadMediaResult{
			Duplicate: &existing,
			Pending:   &pending,
		}, nil
	}

	record := s.repo.Create(pending.Media)
	_ = s.deletePendingUpload(token)
	return UploadMediaResult{Media: record}, nil
}

func (s *MediaService) ConfirmPendingUpload(ctx context.Context, token string) (models.Media, error) {
	pending, err := s.loadPendingUpload(token)
	if err != nil {
		return models.Media{}, err
	}

	_, storageBackend, err := s.storageByID(ctx, pending.Media.StorageID)
	if err != nil {
		_ = s.deletePendingUpload(token)
		return models.Media{}, ErrPendingUploadMissing
	}

	reader, _, storageErr := storageBackend.Get(ctx, pending.Media.PathKey)
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

	_, storageBackend, err := s.storageByID(ctx, pending.Media.StorageID)
	if err != nil {
		return nil, media.ObjectMeta{}, ErrPendingUploadMissing
	}

	reader, meta, err := storageBackend.Get(ctx, pending.Media.PathKey)
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

	_, storageBackend, err := s.storageByID(ctx, pending.Media.StorageID)
	if err != nil {
		return ErrPendingUploadMissing
	}

	if err := storageBackend.Delete(ctx, pending.Media.PathKey); err != nil {
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

	_, storageBackend, err := s.storageByID(ctx, item.StorageID)
	if err != nil {
		return err
	}

	if err := storageBackend.Delete(ctx, item.PathKey); err != nil {
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

	_, storageBackend, err := s.storageByID(ctx, item.StorageID)
	if err != nil {
		return nil, media.ObjectMeta{}, err
	}

	reader, meta, err := storageBackend.Get(ctx, item.PathKey)
	if err != nil {
		return nil, media.ObjectMeta{}, err
	}

	meta.ContentType = item.MimeType
	meta.Size = item.Size
	return reader, meta, nil
}

func (s *MediaService) PublicURL(item models.Media) (string, error) {
	ext, ok := media.ExtensionByMimeType(item.MimeType)
	if !ok {
		return "", ErrMediaNotFound
	}

	return "/uploads/" + item.ID + "." + ext, nil
}

func (s *MediaService) DirectURL(ctx context.Context, item models.Media) (string, error) {
	storageModel, storageBackend, err := s.storageByID(ctx, item.StorageID)
	if err != nil {
		return "", err
	}
	if storageModel.Type == models.StorageTypeLocal {
		return s.PublicURL(item)
	}

	return storageBackend.PublicURL(ctx, item.PathKey)
}

func mediaPathKey(id, ext string) string {
	return filepath.Join("uploads", id+"."+ext)
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

func (s *MediaService) readDimensions(ctx context.Context, storageBackend media.Storage, pathKey, mimeType string) (int, int, error) {
	reader, _, err := storageBackend.Get(ctx, pathKey)
	if err != nil {
		return 0, 0, err
	}
	defer reader.Close()

	return media.DimensionsFromReader(mimeType, reader)
}

func (s *MediaService) defaultStorage(ctx context.Context) (models.Storage, media.Storage, error) {
	return s.storageByID(ctx, s.defaultStorageID())
}

func (s *MediaService) defaultStorageID() string {
	if s.optionRepo == nil {
		return DefaultLocalStorageID
	}

	option := s.optionRepo.GetOptionByKey(DefaultStorageIDOptionKey)
	if option.Value == "" {
		return DefaultLocalStorageID
	}

	return option.Value
}

func (s *MediaService) storageByID(ctx context.Context, id string) (models.Storage, media.Storage, error) {
	if strings.TrimSpace(id) == "" {
		id = DefaultLocalStorageID
	}

	storageModel := s.storageRepo.GetByID(id)
	if storageModel.ID == "" {
		return models.Storage{}, nil, ErrStorageNotFound
	}

	storageBackend, err := s.storageBackend(ctx, storageModel)
	if err != nil {
		return models.Storage{}, nil, err
	}

	return storageModel, storageBackend, nil
}

func (s *MediaService) storageBackend(ctx context.Context, storageModel models.Storage) (media.Storage, error) {
	switch storageModel.Type {
	case models.StorageTypeLocal:
		root := strings.TrimSpace(fmt.Sprint(storageModel.Config["root"]))
		return media.NewLocalStorage(root), nil
	case models.StorageTypeS3:
		credentials, err := s3CredentialsFromStorage(storageModel, s.secrets)
		if err != nil {
			return nil, err
		}

		return media.NewS3Storage(ctx, media.S3Config{
			Endpoint:  strings.TrimSpace(fmt.Sprint(storageModel.Config["endpoint"])),
			Region:    strings.TrimSpace(fmt.Sprint(storageModel.Config["region"])),
			Bucket:    strings.TrimSpace(fmt.Sprint(storageModel.Config["bucket"])),
			AccessKey: strings.TrimSpace(credentials.AccessKey),
			SecretKey: strings.TrimSpace(credentials.SecretKey),
		})
	default:
		return nil, ErrStorageNotFound
	}
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

	if err := s.writePendingUpload(*pending); err != nil {
		return nil, err
	}

	return pending, nil
}

func (s *MediaService) writePendingUpload(pending PendingUpload) error {
	data, err := json.Marshal(pending)
	if err != nil {
		return err
	}

	if err := os.WriteFile(s.pendingPath(pending.Token), data, 0o644); err != nil {
		return err
	}

	return nil
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
