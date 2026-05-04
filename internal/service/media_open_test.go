package service

import (
	"context"
	"errors"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

func TestOpenByIDAndExtReturnsNotFoundForNonLocalStorage(t *testing.T) {
	db := newMediaVariantTestDB(t)
	mediaRepo := repository.NewMediaRepository(db)
	storageRepo := repository.NewStorageRepository(db)
	mediaSvc := NewMediaService(mediaRepo, storageRepo, nil, nil, t.TempDir())

	storageRepo.Create(models.Storage{
		ID:     "s3-storage",
		Name:   "S3 Storage",
		Type:   models.StorageTypeS3,
		Config: models.JSONMap{},
	})
	mediaRepo.Create(models.Media{
		ID:           "s3-media",
		StorageID:    "s3-storage",
		OriginalName: "hero.png",
		PathKey:      "uploads/s3-media.png",
		MimeType:     "image/png",
		Size:         10,
		Hash:         "s3-media-hash",
		Metadata:     models.JSONMap{},
	})

	reader, _, err := mediaSvc.OpenByIDAndExt(context.Background(), "s3-media", "png")
	if reader != nil {
		reader.Close()
	}
	if !errors.Is(err, ErrMediaNotFound) {
		t.Fatalf("expected ErrMediaNotFound, got %v", err)
	}
}

func TestOpenVariantByPathKeyReturnsNotFoundForNonLocalStorage(t *testing.T) {
	db := newMediaVariantTestDB(t)
	mediaRepo := repository.NewMediaRepository(db)
	storageRepo := repository.NewStorageRepository(db)
	mediaSvc := NewMediaService(mediaRepo, storageRepo, nil, nil, t.TempDir())

	storageRepo.Create(models.Storage{
		ID:     "s3-storage",
		Name:   "S3 Storage",
		Type:   models.StorageTypeS3,
		Config: models.JSONMap{},
	})
	item := mediaRepo.Create(models.Media{
		ID:           "s3-media",
		StorageID:    "s3-storage",
		OriginalName: "hero.png",
		PathKey:      "uploads/s3-media.png",
		MimeType:     "image/png",
		Size:         10,
		Hash:         "s3-media-hash",
		Metadata:     models.JSONMap{},
	})
	if err := db.Create(&models.MediaVariant{
		ID:       "s3-variant",
		MediaID:  item.ID,
		PathKey:  "uploads/variants/s3-media_w320.webp",
		Size:     5,
		MimeType: "image/webp",
		Metadata: models.JSONMap{"width": 320, "config_version": 1},
	}).Error; err != nil {
		t.Fatalf("create variant: %v", err)
	}

	reader, _, err := mediaSvc.OpenVariantByPathKey(context.Background(), "uploads/variants/s3-media_w320.webp")
	if reader != nil {
		reader.Close()
	}
	if !errors.Is(err, ErrMediaNotFound) {
		t.Fatalf("expected ErrMediaNotFound, got %v", err)
	}
}

func TestOpenPendingUploadReturnsMissingForNonLocalStorage(t *testing.T) {
	db := newMediaVariantTestDB(t)
	mediaRepo := repository.NewMediaRepository(db)
	storageRepo := repository.NewStorageRepository(db)
	mediaSvc := NewMediaService(mediaRepo, storageRepo, nil, nil, t.TempDir())

	storageRepo.Create(models.Storage{
		ID:     "s3-storage",
		Name:   "S3 Storage",
		Type:   models.StorageTypeS3,
		Config: models.JSONMap{},
	})
	pending, err := mediaSvc.savePendingUpload(models.Media{
		ID:           "s3-pending",
		StorageID:    "s3-storage",
		OriginalName: "hero.png",
		PathKey:      "uploads/s3-pending.png",
		MimeType:     "image/png",
		Size:         10,
		Hash:         "s3-pending-hash",
		Metadata:     models.JSONMap{},
	}, "")
	if err != nil {
		t.Fatalf("save pending upload: %v", err)
	}

	reader, _, err := mediaSvc.OpenPendingUpload(context.Background(), pending.Token)
	if reader != nil {
		reader.Close()
	}
	if !errors.Is(err, ErrPendingUploadMissing) {
		t.Fatalf("expected ErrPendingUploadMissing, got %v", err)
	}
}
