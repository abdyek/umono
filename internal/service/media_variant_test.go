package service

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
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
