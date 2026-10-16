package service

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/umono-cms/umono/internal/models"
)

// imageCall returns an IMAGE call with a media literal whose url is url.
func imageCall(url string) string {
	return "{{ IMAGE media = { url: " + strconv.Quote(url) + ", width: 1, height: 1, mime-type: " + strconv.Quote("image/png") + " } alt = " + strconv.Quote("") + " }}"
}

func TestSitePageServiceRenderedImageURLs(t *testing.T) {
	svc, _ := newPublicRenderTestService(t)

	if _, errs := svc.Create(models.SitePage{
		Name:    "Gallery",
		Slug:    "gallery",
		Enabled: true,
		Content: imageCall("https://cdn.example.net/a.png") + "\n\n" + imageCall("/uploads/b.png"),
	}); len(errs) > 0 {
		t.Fatalf("Create errors: %v", errs)
	}

	r, err := svc.GetRenderedBySlug(context.Background(), "gallery")
	if err != nil {
		t.Fatalf("GetRenderedBySlug: %v", err)
	}
	urls := r.ImageURLs
	want := []string{"https://cdn.example.net/a.png", "/uploads/b.png"}
	if !reflect.DeepEqual(urls, want) {
		t.Fatalf("GetRenderedBySlug urls = %v, want %v", urls, want)
	}

	r, err = svc.GetNotFoundPage(context.Background(), "Not found", imageCall("https://cdn.example.net/c.png"))
	if err != nil {
		t.Fatalf("GetNotFoundPage: %v", err)
	}
	urls = r.ImageURLs
	want = []string{"https://cdn.example.net/c.png"}
	if !reflect.DeepEqual(urls, want) {
		t.Fatalf("GetNotFoundPage urls = %v, want %v", urls, want)
	}
}
