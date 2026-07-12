package service

import (
	"strings"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestOptionServiceLocalStorageImageUploadLimitDefaultsToFourMB(t *testing.T) {
	svc := NewOptionService(repository.NewOptionRepository(newOptionTestDB(t)), nil)

	if got := svc.GetLocalStorageImageUploadLimitMB(); got != DefaultLocalStorageImageUploadLimitMB {
		t.Fatalf("GetLocalStorageImageUploadLimitMB() = %d, want %d", got, DefaultLocalStorageImageUploadLimitMB)
	}
	if got := svc.GetLocalStorageImageUploadLimitBytes(); got != 4*1024*1024 {
		t.Fatalf("GetLocalStorageImageUploadLimitBytes() = %d, want 4194304", got)
	}
}

func TestOptionServiceSaveLocalStorageImageUploadLimit(t *testing.T) {
	svc := NewOptionService(repository.NewOptionRepository(newOptionTestDB(t)), nil)

	if err := svc.SaveLocalStorageImageUploadLimitMB(12); err != nil {
		t.Fatalf("SaveLocalStorageImageUploadLimitMB() error = %v", err)
	}
	if got := svc.GetLocalStorageImageUploadLimitMB(); got != 12 {
		t.Fatalf("GetLocalStorageImageUploadLimitMB() = %d, want 12", got)
	}
}

func TestOptionServiceRejectsInvalidLocalStorageImageUploadLimit(t *testing.T) {
	svc := NewOptionService(repository.NewOptionRepository(newOptionTestDB(t)), nil)

	for _, value := range []int{MinLocalStorageImageUploadLimitMB - 1, -1, MaxLocalStorageImageUploadLimitMB + 1} {
		if err := svc.SaveLocalStorageImageUploadLimitMB(value); err == nil {
			t.Fatalf("SaveLocalStorageImageUploadLimitMB(%d) expected error", value)
		}
	}
}

func TestOptionServiceRobotsTxtDefaultsToEmpty(t *testing.T) {
	svc := NewOptionService(repository.NewOptionRepository(newOptionTestDB(t)), nil)

	if got := svc.GetRobotsTxt(); got != "" {
		t.Fatalf("GetRobotsTxt() = %q, want empty", got)
	}
}

func TestOptionServiceSaveRobotsTxtAllowsEmpty(t *testing.T) {
	svc := NewOptionService(repository.NewOptionRepository(newOptionTestDB(t)), nil)

	if err := svc.SaveRobotsTxt(""); err != nil {
		t.Fatalf("SaveRobotsTxt(\"\") error = %v", err)
	}
	if got := svc.GetRobotsTxt(); got != "" {
		t.Fatalf("GetRobotsTxt() = %q, want empty", got)
	}
}

func TestOptionServiceRejectsRobotsTxtOverLimit(t *testing.T) {
	svc := NewOptionService(repository.NewOptionRepository(newOptionTestDB(t)), nil)

	atLimit := strings.Repeat("a", MaxRobotsTxtBytes)
	if err := svc.SaveRobotsTxt(atLimit); err != nil {
		t.Fatalf("SaveRobotsTxt() at limit error = %v", err)
	}

	overLimit := strings.Repeat("a", MaxRobotsTxtBytes+1)
	if err := svc.SaveRobotsTxt(overLimit); err != ErrRobotsTxtTooLarge {
		t.Fatalf("SaveRobotsTxt() over limit error = %v, want ErrRobotsTxtTooLarge", err)
	}
}

func newOptionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Option{}); err != nil {
		t.Fatalf("migrate options: %v", err)
	}
	return db
}
