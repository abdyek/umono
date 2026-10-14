package service

import (
	"errors"
	"testing"

	"github.com/umono-cms/umono/internal/repository"
)

func TestNormalizeSiteURL(t *testing.T) {
	accepted := []struct {
		input string
		want  string
	}{
		{"https://example.com", "https://example.com"},
		{"https://example.com/", "https://example.com"},
		{"http://localhost:8999", "http://localhost:8999"},
		{"  https://example.com  ", "https://example.com"},
		{"HTTPS://example.com", "https://example.com"},
	}

	for _, tc := range accepted {
		got, err := NormalizeSiteURL(tc.input)
		if err != nil {
			t.Fatalf("NormalizeSiteURL(%q) error = %v, want nil", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizeSiteURL(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}

	rejected := []string{
		"",
		"example.com",
		"//example.com",
		"ftp://example.com",
		"javascript:alert(1)",
		"https://",
		"https://example.com/blog",
		"https://example.com?page=2",
		"https://example.com?",
		"https://example.com#top",
		"https://example.com#",
		"https://user@example.com",
		"https://user:pass@example.com",
	}

	for _, input := range rejected {
		got, err := NormalizeSiteURL(input)
		if !errors.Is(err, ErrInvalidSiteURL) {
			t.Fatalf("NormalizeSiteURL(%q) error = %v, want ErrInvalidSiteURL", input, err)
		}
		if got != "" {
			t.Fatalf("NormalizeSiteURL(%q) = %q, want empty", input, got)
		}
	}
}

func TestOptionServiceSavesSiteURL(t *testing.T) {
	svc := NewOptionService(repository.NewOptionRepository(newOptionTestDB(t)), nil)

	if got := svc.GetSiteURL(); got != "" {
		t.Fatalf("GetSiteURL() = %q, want empty", got)
	}

	if err := svc.SaveSiteURL("https://example.com/"); err != nil {
		t.Fatalf("SaveSiteURL(%q) error = %v, want nil", "https://example.com/", err)
	}
	if got := svc.GetSiteURL(); got != "https://example.com" {
		t.Fatalf("GetSiteURL() = %q, want %q", got, "https://example.com")
	}

	if err := svc.SaveSiteURL("ftp://example.com"); !errors.Is(err, ErrInvalidSiteURL) {
		t.Fatalf("SaveSiteURL(%q) error = %v, want ErrInvalidSiteURL", "ftp://example.com", err)
	}
	if got := svc.GetSiteURL(); got != "https://example.com" {
		t.Fatalf("GetSiteURL() = %q, want %q", got, "https://example.com")
	}

	if err := svc.SaveSiteURL("   "); err != nil {
		t.Fatalf("SaveSiteURL(%q) error = %v, want nil", "   ", err)
	}
	if got := svc.GetSiteURL(); got != "" {
		t.Fatalf("GetSiteURL() = %q, want empty", got)
	}
}
