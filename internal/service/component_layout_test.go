package service

import (
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newLayoutComponentService(t *testing.T) (*ComponentService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Component{}); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	cc, err := NewContentCompiler(nil)
	if err != nil {
		t.Fatalf("new content compiler: %v", err)
	}

	svc := NewComponentService(repository.NewComponentRepository(db))
	svc.SetContentCompiler(cc)
	return svc, db
}

func TestGlobalCannotCallTheLayout(t *testing.T) {
	svc, db := newLayoutComponentService(t)

	contents := []string{
		"{{ UMONO_LAYOUT }}\n",
		"Hello {{ UMONO_LAYOUT title = \"x\" }}\n",
		"{{ CARD body = UMONO_LAYOUT }}\n",
	}

	for _, content := range contents {
		_, errs := svc.Create(models.Component{Name: "HEADER_BAR", Content: content})
		if !ErrorsIs(errs, ErrGlobalCallsLayout) {
			t.Fatalf("expected ErrGlobalCallsLayout for content %q, got %v", content, errs)
		}
	}

	var n int64
	if err := db.Model(&models.Component{}).Count(&n).Error; err != nil {
		t.Fatalf("count components: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 components, got %d", n)
	}
}

func TestGlobalUpdateCannotCallTheLayout(t *testing.T) {
	svc, db := newLayoutComponentService(t)

	created, errs := svc.Create(models.Component{Name: "HEADER_BAR", Content: "# Hi\n"})
	if len(errs) != 0 {
		t.Fatalf("create component: %v", errs)
	}

	created.Content = "{{ UMONO_LAYOUT }}\n"
	_, errs = svc.Update(created)
	if !ErrorsIs(errs, ErrGlobalCallsLayout) {
		t.Fatalf("expected ErrGlobalCallsLayout, got %v", errs)
	}

	var got models.Component
	if err := db.First(&got, created.ID).Error; err != nil {
		t.Fatalf("reload component: %v", err)
	}
	if got.Content != "# Hi\n" {
		t.Fatalf("expected content %q, got %q", "# Hi\n", got.Content)
	}
}

func TestGlobalWithoutLayoutCallIsSaved(t *testing.T) {
	svc, _ := newLayoutComponentService(t)

	_, errs := svc.Create(models.Component{Name: "HEADER_BAR", Content: "~ title = \"\"\n# {{ title }}\n"})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}
