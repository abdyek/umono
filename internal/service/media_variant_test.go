package service

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMediaVariantJobsGenerateVariantsAfterUpload(t *testing.T) {
	db := newMediaVariantTestDB(t)
	mediaRepo := repository.NewMediaRepository(db)
	storageRepo := repository.NewStorageRepository(db)
	jobSvc := NewJobService(repository.NewJobRepository(db))
	mediaSvc := NewMediaService(mediaRepo, storageRepo, nil, nil, t.TempDir())
	if err := mediaSvc.EnsureDefaultLocalStorage(t.TempDir()); err != nil {
		t.Fatalf("ensure local storage: %v", err)
	}
	if err := mediaSvc.RegisterVariantJobHandlers(jobSvc); err != nil {
		t.Fatalf("register variant handlers: %v", err)
	}

	result, err := mediaSvc.Upload(context.Background(), UploadMediaInput{
		StorageID:    DefaultLocalStorageID,
		OriginalName: "hero.png",
		MimeType:     "image/png",
		Reader:       bytes.NewReader(testPNG(t, 700, 350)),
	})
	if err != nil {
		t.Fatalf("upload media: %v", err)
	}
	if result.Media.ID == "" {
		t.Fatal("expected media record")
	}

	processAllJobs(t, jobSvc)

	var variants []models.MediaVariant
	if err := db.Order("mime_type ASC").Find(&variants).Error; err != nil {
		t.Fatalf("list variants: %v", err)
	}
	if len(variants) != 7 {
		t.Fatalf("expected 7 variants, got %d: %#v", len(variants), variants)
	}

	seen := map[string]bool{}
	for _, variant := range variants {
		if variant.MediaID != result.Media.ID {
			t.Fatalf("variant media id mismatch: got %q want %q", variant.MediaID, result.Media.ID)
		}
		if variant.PathKey == "" || variant.Size <= 0 {
			t.Fatalf("variant was not stored correctly: %#v", variant)
		}
		if variant.Metadata["width"] == nil || variant.Metadata["config_version"] == nil {
			t.Fatalf("variant target values should be stored in metadata: %#v", variant.Metadata)
		}
		seen[variant.MimeType] = true
	}
	if !seen["image/webp"] || !seen["image/png"] {
		t.Fatalf("expected webp and png variants, got %#v", seen)
	}
}

func TestMediaDeleteRemovesVariantRecords(t *testing.T) {
	db := newMediaVariantTestDB(t)
	mediaRepo := repository.NewMediaRepository(db)
	storageRepo := repository.NewStorageRepository(db)
	mediaSvc := NewMediaService(mediaRepo, storageRepo, nil, nil, t.TempDir())

	storageRoot := t.TempDir()
	if err := mediaSvc.EnsureDefaultLocalStorage(storageRoot); err != nil {
		t.Fatalf("ensure local storage: %v", err)
	}

	writeStorageFile(t, storageRoot, "uploads/original.png")
	writeStorageFile(t, storageRoot, "uploads/variants/original_w320.webp")
	writeStorageFile(t, storageRoot, "uploads/variants/original_w640.webp")

	item := mediaRepo.Create(models.Media{
		ID:           "media-delete-test",
		StorageID:    DefaultLocalStorageID,
		OriginalName: "original.png",
		PathKey:      "uploads/original.png",
		MimeType:     "image/png",
		Size:         10,
		Hash:         "delete-test-hash",
		Metadata:     models.JSONMap{},
	})
	if err := db.Create(&[]models.MediaVariant{
		{
			ID:       "variant-320",
			MediaID:  item.ID,
			PathKey:  "uploads/variants/original_w320.webp",
			Size:     5,
			MimeType: "image/webp",
			Metadata: models.JSONMap{"width": 320, "config_version": 1},
		},
		{
			ID:       "variant-640",
			MediaID:  item.ID,
			PathKey:  "uploads/variants/original_w640.webp",
			Size:     6,
			MimeType: "image/webp",
			Metadata: models.JSONMap{"width": 640, "config_version": 1},
		},
	}).Error; err != nil {
		t.Fatalf("create variants: %v", err)
	}

	if err := mediaSvc.Delete(context.Background(), item.ID); err != nil {
		t.Fatalf("delete media: %v", err)
	}

	var variantCount int64
	if err := db.Model(&models.MediaVariant{}).Where("media_id = ?", item.ID).Count(&variantCount).Error; err != nil {
		t.Fatalf("count variants: %v", err)
	}
	if variantCount != 0 {
		t.Fatalf("expected variant records to be deleted, got %d", variantCount)
	}

	var mediaCount int64
	if err := db.Model(&models.Media{}).Where("id = ?", item.ID).Count(&mediaCount).Error; err != nil {
		t.Fatalf("count media: %v", err)
	}
	if mediaCount != 0 {
		t.Fatalf("expected media record to be deleted, got %d", mediaCount)
	}

	for _, key := range []string{
		"original.png",
		"variants/original_w320.webp",
		"variants/original_w640.webp",
	} {
		if _, err := os.Stat(filepath.Join(storageRoot, key)); !os.IsNotExist(err) {
			t.Fatalf("expected storage file %q to be deleted, stat err: %v", key, err)
		}
	}
}

func processAllJobs(t *testing.T, svc *JobService) {
	t.Helper()

	for i := 0; i < 32; i++ {
		processed, err := svc.processNext(context.Background())
		if err != nil {
			t.Fatalf("process job: %v", err)
		}
		if !processed {
			return
		}
	}
	t.Fatal("job queue did not drain")
}

func testPNG(t *testing.T, width, height int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 180, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func writeStorageFile(t *testing.T, root, key string) {
	t.Helper()

	key = strings.TrimPrefix(filepath.ToSlash(key), "uploads/")
	path := filepath.Join(root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create storage dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("write storage file: %v", err)
	}
}

func newMediaVariantTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Storage{}, &models.Media{}, &models.MediaVariant{}, &models.Job{}); err != nil {
		t.Fatalf("migrate media variant test db: %v", err)
	}
	return db
}
