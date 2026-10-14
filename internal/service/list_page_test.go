package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func newListPageTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.SitePage{}, &models.Component{}, &models.Storage{}, &models.Media{}, &models.MediaVariant{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestSitePageServiceGetPage(t *testing.T) {
	db := newListPageTestDB(t)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i <= 50; i++ {
		last := base.Add(time.Duration(i) * time.Minute)
		if err := db.Create(&models.SitePage{
			Name:           fmt.Sprintf("Page %d", i),
			Slug:           fmt.Sprintf("page-%d", i),
			Enabled:        true,
			LastModifiedAt: &last,
		}).Error; err != nil {
			t.Fatalf("create page %d: %v", i, err)
		}
	}

	svc := NewSitePageService(repository.NewSitePageRepository(db), nil, nil)

	items, next := svc.GetPage(0)
	if got, want := len(items), 50; got != want {
		t.Fatalf("GetPage(0) len(items) = %d, want %d", got, want)
	}
	if got, want := next, 50; got != want {
		t.Fatalf("GetPage(0) next = %d, want %d", got, want)
	}
	if got, want := items[0].Name, "Page 50"; got != want {
		t.Fatalf("GetPage(0) items[0].Name = %q, want %q", got, want)
	}
	if got, want := items[49].Name, "Page 1"; got != want {
		t.Fatalf("GetPage(0) items[49].Name = %q, want %q", got, want)
	}

	items, next = svc.GetPage(50)
	if got, want := len(items), 1; got != want {
		t.Fatalf("GetPage(50) len(items) = %d, want %d", got, want)
	}
	if got, want := next, 0; got != want {
		t.Fatalf("GetPage(50) next = %d, want %d", got, want)
	}
	if got, want := items[0].Name, "Page 0"; got != want {
		t.Fatalf("GetPage(50) items[0].Name = %q, want %q", got, want)
	}

	items, next = svc.GetPage(-1)
	if got, want := len(items), 50; got != want {
		t.Fatalf("GetPage(-1) len(items) = %d, want %d", got, want)
	}
	if got, want := next, 50; got != want {
		t.Fatalf("GetPage(-1) next = %d, want %d", got, want)
	}
	if got, want := items[0].Name, "Page 50"; got != want {
		t.Fatalf("GetPage(-1) items[0].Name = %q, want %q", got, want)
	}
}

func TestSitePageServiceGetPageBreaksTiesByID(t *testing.T) {
	db := newListPageTestDB(t)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, p := range []struct {
		name string
		slug string
	}{
		{"A", "a"},
		{"B", "b"},
		{"C", "c"},
	} {
		last := base
		if err := db.Create(&models.SitePage{
			Name:           p.name,
			Slug:           p.slug,
			Enabled:        true,
			LastModifiedAt: &last,
		}).Error; err != nil {
			t.Fatalf("create page %s: %v", p.name, err)
		}
	}

	svc := NewSitePageService(repository.NewSitePageRepository(db), nil, nil)

	items, next := svc.GetPage(0)
	if got, want := next, 0; got != want {
		t.Fatalf("GetPage(0) next = %d, want %d", got, want)
	}
	wantNames := []string{"C", "B", "A"}
	for i, want := range wantNames {
		if got := items[i].Name; got != want {
			t.Fatalf("GetPage(0) items[%d].Name = %q, want %q", i, got, want)
		}
	}
}

func TestComponentServiceGetPage(t *testing.T) {
	db := newListPageTestDB(t)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i <= 50; i++ {
		last := base.Add(time.Duration(i) * time.Minute)
		if err := db.Create(&models.Component{
			Name:           fmt.Sprintf("COMP_%d", i),
			LastModifiedAt: &last,
		}).Error; err != nil {
			t.Fatalf("create component %d: %v", i, err)
		}
	}

	svc := NewComponentService(repository.NewComponentRepository(db))

	items, next := svc.GetPage(0)
	if got, want := len(items), 50; got != want {
		t.Fatalf("GetPage(0) len(items) = %d, want %d", got, want)
	}
	if got, want := next, 50; got != want {
		t.Fatalf("GetPage(0) next = %d, want %d", got, want)
	}
	if got, want := items[0].Name, "COMP_50"; got != want {
		t.Fatalf("GetPage(0) items[0].Name = %q, want %q", got, want)
	}
	if got, want := items[49].Name, "COMP_1"; got != want {
		t.Fatalf("GetPage(0) items[49].Name = %q, want %q", got, want)
	}

	items, next = svc.GetPage(50)
	if got, want := len(items), 1; got != want {
		t.Fatalf("GetPage(50) len(items) = %d, want %d", got, want)
	}
	if got, want := next, 0; got != want {
		t.Fatalf("GetPage(50) next = %d, want %d", got, want)
	}
	if got, want := items[0].Name, "COMP_0"; got != want {
		t.Fatalf("GetPage(50) items[0].Name = %q, want %q", got, want)
	}
}

func TestMediaServiceGetPage(t *testing.T) {
	db := newListPageTestDB(t)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i <= 50; i++ {
		if err := db.Omit(clause.Associations).Create(&models.Media{
			ID:           fmt.Sprintf("media-%02d", i),
			StorageID:    "storage-1",
			OriginalName: "a.png",
			PathKey:      fmt.Sprintf("a-%d.png", i),
			MimeType:     "image/png",
			Size:         1,
			Hash:         fmt.Sprintf("hash-%d", i),
			CreatedAt:    base.Add(time.Duration(i) * time.Minute),
		}).Error; err != nil {
			t.Fatalf("create media %d: %v", i, err)
		}
	}

	svc := NewMediaService(repository.NewMediaRepository(db), repository.NewStorageRepository(db), nil, nil, t.TempDir())

	items, next := svc.GetPage(0)
	if got, want := len(items), 50; got != want {
		t.Fatalf("GetPage(0) len(items) = %d, want %d", got, want)
	}
	if got, want := next, 50; got != want {
		t.Fatalf("GetPage(0) next = %d, want %d", got, want)
	}
	if got, want := items[0].ID, "media-50"; got != want {
		t.Fatalf("GetPage(0) items[0].ID = %q, want %q", got, want)
	}
	if got, want := items[49].ID, "media-01"; got != want {
		t.Fatalf("GetPage(0) items[49].ID = %q, want %q", got, want)
	}

	items, next = svc.GetPage(50)
	if got, want := len(items), 1; got != want {
		t.Fatalf("GetPage(50) len(items) = %d, want %d", got, want)
	}
	if got, want := next, 0; got != want {
		t.Fatalf("GetPage(50) next = %d, want %d", got, want)
	}
	if got, want := items[0].ID, "media-00"; got != want {
		t.Fatalf("GetPage(50) items[0].ID = %q, want %q", got, want)
	}
}
