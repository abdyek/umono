package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/umono-cms/umono/internal/media"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

var (
	ErrStorageNotFound          = errors.New("storage not found")
	ErrStorageDeleteRestricted  = errors.New("storage delete restricted")
	ErrStorageReadonly          = errors.New("storage readonly")
	ErrStorageNameRequired      = errors.New("storage name required")
	ErrStorageEndpointRequired  = errors.New("storage endpoint required")
	ErrStorageRegionRequired    = errors.New("storage region required")
	ErrStorageBucketRequired    = errors.New("storage bucket required")
	ErrStorageAccessKeyRequired = errors.New("storage access key required")
	ErrStorageSecretKeyRequired = errors.New("storage secret key required")
	ErrStorageTestBodyMismatch  = errors.New("storage test body mismatch")
)

const storageTestBody = "umono-storage-test"

type StorageService struct {
	repo          *repository.StorageRepository
	secretService *SecretService
}

type StorageInput struct {
	Name      string
	IsDefault bool
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

type StorageS3Credentials struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

type StorageValidationError struct {
	FieldErrors map[string]error
}

type StorageDeleteState struct {
	CanDelete  bool
	MediaCount int64
	ReasonKey  string
}

type StorageTestError struct {
	Step string
	Err  error
}

func (e *StorageValidationError) Error() string {
	return "storage validation failed"
}

func (e *StorageValidationError) add(field string, err error) {
	if e.FieldErrors == nil {
		e.FieldErrors = map[string]error{}
	}
	e.FieldErrors[field] = err
}

func (e *StorageValidationError) Empty() bool {
	return len(e.FieldErrors) == 0
}

func (e *StorageTestError) Error() string {
	return fmt.Sprintf("%s: %v", e.Step, e.Err)
}

func (e *StorageTestError) Unwrap() error {
	return e.Err
}

func NewStorageService(repo *repository.StorageRepository, secretService *SecretService) *StorageService {
	return &StorageService{
		repo:          repo,
		secretService: secretService,
	}
}

func (s *StorageService) GetAll() []models.Storage {
	storages := s.repo.GetAll()
	slices.SortFunc(storages, func(a, b models.Storage) int {
		if a.ID == DefaultLocalStorageID && b.ID != DefaultLocalStorageID {
			return -1
		}
		if b.ID == DefaultLocalStorageID && a.ID != DefaultLocalStorageID {
			return 1
		}

		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return storages
}

func (s *StorageService) GetByID(id string) (models.Storage, error) {
	storage := s.repo.GetByID(id)
	if storage.ID == "" {
		return models.Storage{}, ErrStorageNotFound
	}

	return storage, nil
}

func (s *StorageService) CreateS3(input StorageInput) (models.Storage, error) {
	if validationErr := validateStorageInput(input, true); validationErr != nil {
		return models.Storage{}, validationErr
	}

	id, err := uuid.NewV7()
	if err != nil {
		return models.Storage{}, err
	}

	secret, err := s.createS3CredentialsSecret(input)
	if err != nil {
		return models.Storage{}, err
	}

	storage := models.Storage{
		ID:     id.String(),
		Name:   strings.TrimSpace(input.Name),
		Type:   models.StorageTypeS3,
		Config: storageConfig(input, secret.ID),
	}

	return s.repo.Create(storage), nil
}

func (s *StorageService) UpdateS3(id string, input StorageInput) (models.Storage, error) {
	storage, err := s.GetByID(id)
	if err != nil {
		return models.Storage{}, err
	}
	if storage.Type != models.StorageTypeS3 {
		return models.Storage{}, ErrStorageReadonly
	}

	if validationErr := validateStorageInput(input, false); validationErr != nil {
		return models.Storage{}, validationErr
	}

	credentialRef, err := s.saveS3Credentials(storage, input)
	if err != nil {
		return models.Storage{}, err
	}

	storage.Name = strings.TrimSpace(input.Name)
	storage.Config = storageConfig(input, credentialRef)

	return s.repo.Update(storage), nil
}

func (s *StorageService) Delete(id string) error {
	storage, err := s.GetByID(id)
	if err != nil {
		return err
	}
	if storage.ID == DefaultLocalStorageID {
		return ErrStorageDeleteRestricted
	}

	credentialRef := StorageConfigValue(storage, "credential_ref")
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	if credentialRef != "" && s.secretService != nil {
		return s.secretService.Delete(credentialRef)
	}

	return nil
}

func (s *StorageService) DeleteState(storage models.Storage, defaultStorageID string) StorageDeleteState {
	mediaCount := s.repo.CountMediaByStorageID(storage.ID)
	isDefault := storage.ID == defaultStorageID

	state := StorageDeleteState{
		CanDelete:  true,
		MediaCount: mediaCount,
	}

	switch {
	case storage.ID == DefaultLocalStorageID:
		state.CanDelete = false
		state.ReasonKey = "settings.storage.delete_disabled.local"
	case isDefault && mediaCount > 0:
		state.CanDelete = false
		state.ReasonKey = "settings.storage.delete_disabled.default_with_media"
	case isDefault:
		state.CanDelete = false
		state.ReasonKey = "settings.storage.delete_disabled.default"
	case mediaCount > 0:
		state.CanDelete = false
		state.ReasonKey = "settings.storage.delete_disabled.has_media"
	}

	return state
}

func (s *StorageService) TestS3(ctx context.Context, id string) error {
	storage, err := s.GetByID(id)
	if err != nil {
		return err
	}
	if storage.Type != models.StorageTypeS3 {
		return ErrStorageReadonly
	}

	cfg, err := s.s3ConfigFromStorage(storage)
	if err != nil {
		return err
	}

	return testS3Config(ctx, cfg)
}

func (s *StorageService) TestS3Input(ctx context.Context, input StorageInput) error {
	if validationErr := validateStorageInput(input, true); validationErr != nil {
		return validationErr
	}

	return testS3Config(ctx, media.S3Config{
		Endpoint:  strings.TrimSpace(input.Endpoint),
		Region:    strings.TrimSpace(input.Region),
		Bucket:    strings.TrimSpace(input.Bucket),
		AccessKey: strings.TrimSpace(input.AccessKey),
		SecretKey: strings.TrimSpace(input.SecretKey),
	})
}

func (s *StorageService) S3Credentials(storage models.Storage) (StorageS3Credentials, error) {
	return s3CredentialsFromStorage(storage, s.secretService)
}

func storageConfig(input StorageInput, credentialRef string) models.JSONMap {
	return models.JSONMap{
		"endpoint":       strings.TrimSpace(input.Endpoint),
		"region":         strings.TrimSpace(input.Region),
		"bucket":         strings.TrimSpace(input.Bucket),
		"credential_ref": strings.TrimSpace(credentialRef),
	}
}

func validateStorageInput(input StorageInput, requireSecretKey bool) error {
	errs := &StorageValidationError{}

	if strings.TrimSpace(input.Name) == "" {
		errs.add("name", ErrStorageNameRequired)
	}
	if strings.TrimSpace(input.Endpoint) == "" {
		errs.add("endpoint", ErrStorageEndpointRequired)
	}
	if strings.TrimSpace(input.Region) == "" {
		errs.add("region", ErrStorageRegionRequired)
	}
	if strings.TrimSpace(input.Bucket) == "" {
		errs.add("bucket", ErrStorageBucketRequired)
	}
	if strings.TrimSpace(input.AccessKey) == "" {
		errs.add("access_key", ErrStorageAccessKeyRequired)
	}
	if requireSecretKey && strings.TrimSpace(input.SecretKey) == "" {
		errs.add("secret_key", ErrStorageSecretKeyRequired)
	}

	if errs.Empty() {
		return nil
	}

	return errs
}

func StorageConfigValue(storage models.Storage, key string) string {
	if storage.Config == nil {
		return ""
	}

	value, ok := storage.Config[key]
	if !ok {
		return ""
	}

	return fmt.Sprint(value)
}

func (s *StorageService) createS3CredentialsSecret(input StorageInput) (models.Secret, error) {
	if s.secretService == nil {
		return models.Secret{}, ErrSecretNotFound
	}

	plaintext, err := s3CredentialsPlaintext(input)
	if err != nil {
		return models.Secret{}, err
	}

	return s.secretService.Create(plaintext)
}

func (s *StorageService) saveS3Credentials(storage models.Storage, input StorageInput) (string, error) {
	if s.secretService == nil {
		return "", ErrSecretNotFound
	}

	if strings.TrimSpace(input.SecretKey) == "" {
		credentials, err := s3CredentialsFromStorage(storage, s.secretService)
		if err != nil {
			return "", err
		}
		input.SecretKey = credentials.SecretKey
	}

	plaintext, err := s3CredentialsPlaintext(input)
	if err != nil {
		return "", err
	}

	credentialRef := StorageConfigValue(storage, "credential_ref")
	if credentialRef != "" {
		if _, err := s.secretService.Update(credentialRef, plaintext); err == nil {
			return credentialRef, nil
		} else if !errors.Is(err, ErrSecretNotFound) {
			return "", err
		}
	}

	secret, err := s.createS3CredentialsSecret(input)
	if err != nil {
		return "", err
	}

	return secret.ID, nil
}

func s3CredentialsPlaintext(input StorageInput) ([]byte, error) {
	return json.Marshal(StorageS3Credentials{
		AccessKey: strings.TrimSpace(input.AccessKey),
		SecretKey: strings.TrimSpace(input.SecretKey),
	})
}

func s3CredentialsFromStorage(storage models.Storage, secretService *SecretService) (StorageS3Credentials, error) {
	credentialRef := strings.TrimSpace(StorageConfigValue(storage, "credential_ref"))
	if credentialRef == "" || secretService == nil {
		return StorageS3Credentials{}, ErrSecretNotFound
	}

	plaintext, err := secretService.DecryptByID(credentialRef)
	if err != nil {
		return StorageS3Credentials{}, err
	}

	var credentials StorageS3Credentials
	if err := json.Unmarshal(plaintext, &credentials); err != nil {
		return StorageS3Credentials{}, err
	}

	return credentials, nil
}

func (s *StorageService) s3ConfigFromStorage(storage models.Storage) (media.S3Config, error) {
	credentials, err := s3CredentialsFromStorage(storage, s.secretService)
	if err != nil {
		return media.S3Config{}, err
	}

	return media.S3Config{
		Endpoint:  strings.TrimSpace(StorageConfigValue(storage, "endpoint")),
		Region:    strings.TrimSpace(StorageConfigValue(storage, "region")),
		Bucket:    strings.TrimSpace(StorageConfigValue(storage, "bucket")),
		AccessKey: strings.TrimSpace(credentials.AccessKey),
		SecretKey: strings.TrimSpace(credentials.SecretKey),
	}, nil
}

func testS3Config(ctx context.Context, cfg media.S3Config) error {
	storage, err := media.NewS3Storage(ctx, cfg)
	if err != nil {
		return &StorageTestError{Step: "put", Err: err}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return &StorageTestError{Step: "put", Err: err}
	}

	key := "umono-test/" + id.String()
	if err := storage.Put(ctx, key, strings.NewReader(storageTestBody), media.ObjectMeta{
		ContentType: "text/plain; charset=utf-8",
	}); err != nil {
		return &StorageTestError{Step: "put", Err: err}
	}

	body, _, err := storage.Get(ctx, key)
	if err != nil {
		return &StorageTestError{Step: "get", Err: err}
	}

	data, readErr := io.ReadAll(body)
	closeErr := body.Close()
	switch {
	case readErr != nil:
		return &StorageTestError{Step: "get", Err: readErr}
	case closeErr != nil:
		return &StorageTestError{Step: "get", Err: closeErr}
	case !bytes.Equal(data, []byte(storageTestBody)):
		return &StorageTestError{Step: "get", Err: ErrStorageTestBodyMismatch}
	}

	if err := storage.Delete(ctx, key); err != nil {
		return &StorageTestError{Step: "delete", Err: err}
	}

	return nil
}
