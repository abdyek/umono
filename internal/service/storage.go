package service

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
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
)

type StorageService struct {
	repo *repository.StorageRepository
}

type StorageInput struct {
	Name      string
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

type StorageValidationError struct {
	FieldErrors map[string]error
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

func NewStorageService(repo *repository.StorageRepository) *StorageService {
	return &StorageService{repo: repo}
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
	if validationErr := validateStorageInput(input); validationErr != nil {
		return models.Storage{}, validationErr
	}

	id, err := uuid.NewV7()
	if err != nil {
		return models.Storage{}, err
	}

	storage := models.Storage{
		ID:     id.String(),
		Name:   strings.TrimSpace(input.Name),
		Type:   models.StorageTypeS3,
		Config: storageConfig(input),
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

	if validationErr := validateStorageInput(input); validationErr != nil {
		return models.Storage{}, validationErr
	}

	storage.Name = strings.TrimSpace(input.Name)
	storage.Config = storageConfig(input)

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

	return s.repo.Delete(id)
}

func storageConfig(input StorageInput) models.JSONMap {
	return models.JSONMap{
		"endpoint":   strings.TrimSpace(input.Endpoint),
		"region":     strings.TrimSpace(input.Region),
		"bucket":     strings.TrimSpace(input.Bucket),
		"access_key": strings.TrimSpace(input.AccessKey),
		"secret_key": strings.TrimSpace(input.SecretKey),
	}
}

func validateStorageInput(input StorageInput) error {
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
	if strings.TrimSpace(input.SecretKey) == "" {
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
