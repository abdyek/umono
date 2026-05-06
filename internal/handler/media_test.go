package handler

import (
	"testing"

	"github.com/umono-cms/umono/internal/models"
)

func TestMediaClipboardSnippetUsesAliasForImages(t *testing.T) {
	item := models.Media{
		ID:       "media-1",
		MimeType: "image/png",
		Metadata: models.JSONMap{
			"alias": "hero-banner",
		},
	}

	got := mediaClipboard(item)
	want := `{{ IMAGE media = context(media-by-alias/hero-banner) alt = "" }}`
	if got.Snippet != want {
		t.Fatalf("mediaClipboard().Snippet = %q, want %q", got.Snippet, want)
	}
	if got.Label != "media.detail.copy_image_builtin" {
		t.Fatalf("mediaClipboard().Label = %q", got.Label)
	}
}

func TestMediaClipboardSnippetFallsBackToIDForImages(t *testing.T) {
	item := models.Media{
		ID:       "media-1",
		MimeType: "image/webp",
	}

	got := mediaClipboard(item)
	want := `{{ IMAGE media = context(media-by-id/media-1) alt = "" }}`
	if got.Snippet != want {
		t.Fatalf("mediaClipboard().Snippet = %q, want %q", got.Snippet, want)
	}
}

func TestMediaClipboardSnippetIgnoresUnsupportedMediaTypes(t *testing.T) {
	item := models.Media{
		ID:       "media-1",
		MimeType: "application/pdf",
	}

	if got := mediaClipboard(item); got != (mediaClipboardData{}) {
		t.Fatalf("mediaClipboard() = %#v, want empty", got)
	}
}
