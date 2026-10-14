package migration

import (
	"fmt"
	"testing"
	"time"

	"github.com/umono-cms/umono/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Component{}, &models.Option{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestGlobalParamLineMarksOldParameterLines(t *testing.T) {
	db := newTestDB(t)

	pageTitleContent := fmt.Sprintf("title = %q\n# {{ title }}", "")

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "PAGE_TITLE",
			content: pageTitleContent,
			want:    "~ " + pageTitleContent,
		},
		{
			name:    "WRAPPER",
			content: "content = NO_MATTER\n{{ content }}",
			want:    "~ content = NO_MATTER\n{{ content }}",
		},
		{
			name:    "COUNTER",
			content: "count = 0\n{{ count }}",
			want:    "~ count = 0\n{{ count }}",
		},
		{
			name:    "GREETING",
			content: "name = context(user/name)\nHello {{ name }}",
			want:    "~ name = context(user/name)\nHello {{ name }}",
		},
		{
			name:    "PLAIN",
			content: "# Hello",
			want:    "# Hello",
		},
		{
			name:    "BLANK_FIRST",
			content: "\ncount = 0\n{{ count }}",
			want:    "\ncount = 0\n{{ count }}",
		},
		{
			name:    "MARKED",
			content: "~ count = 0\n{{ count }}",
			want:    "~ count = 0\n{{ count }}",
		},
		{
			name:    "LOWER_VALUE",
			content: "a = b",
			want:    "a = b",
		},
	}

	for _, tt := range tests {
		if err := db.Create(&models.Component{Name: tt.name, Content: tt.content}).Error; err != nil {
			t.Fatalf("create %s: %v", tt.name, err)
		}
	}

	if err := GlobalParamLine(db); err != nil {
		t.Fatalf("GlobalParamLine: %v", err)
	}

	for _, tt := range tests {
		var got models.Component
		if err := db.Where("name = ?", tt.name).First(&got).Error; err != nil {
			t.Fatalf("read %s: %v", tt.name, err)
		}
		if got.Content != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got.Content, tt.want)
		}
	}
}

func TestGlobalParamLineRunsOnce(t *testing.T) {
	db := newTestDB(t)

	if err := GlobalParamLine(db); err != nil {
		t.Fatalf("first GlobalParamLine: %v", err)
	}

	if err := db.Create(&models.Component{Name: "COUNTER", Content: "count = 0\n{{ count }}"}).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := GlobalParamLine(db); err != nil {
		t.Fatalf("second GlobalParamLine: %v", err)
	}

	var got models.Component
	if err := db.Where("name = ?", "COUNTER").First(&got).Error; err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Content != "count = 0\n{{ count }}" {
		t.Errorf("got %q, want %q", got.Content, "count = 0\n{{ count }}")
	}
}

func TestGlobalParamLineKeepsLastModifiedAt(t *testing.T) {
	db := newTestDB(t)

	mod := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := db.Create(&models.Component{
		Name:           "COUNTER",
		Content:        "count = 0\n{{ count }}",
		LastModifiedAt: &mod,
	}).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := GlobalParamLine(db); err != nil {
		t.Fatalf("GlobalParamLine: %v", err)
	}

	var got models.Component
	if err := db.Where("name = ?", "COUNTER").First(&got).Error; err != nil {
		t.Fatalf("read: %v", err)
	}

	wantContent := "~ count = 0\n{{ count }}"
	if got.Content != wantContent {
		t.Errorf("content: got %q, want %q", got.Content, wantContent)
	}

	if got.LastModifiedAt == nil {
		t.Fatalf("LastModifiedAt is nil")
	}
	if got.LastModifiedAt.Unix() != mod.Unix() {
		t.Errorf("LastModifiedAt.Unix(): got %d, want %d", got.LastModifiedAt.Unix(), mod.Unix())
	}
}
