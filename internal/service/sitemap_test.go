package service

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

func TestSitePageServiceSitemap(t *testing.T) {
	db := newListPageTestDB(t)

	zetaAt := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	homeAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("TRT", 3*60*60))

	pages := []models.SitePage{
		{Name: "Zeta", Slug: "zeta", Enabled: true, LastModifiedAt: &zetaAt},
		{Name: "About", Slug: "about", Enabled: true, LastModifiedAt: nil},
		{Name: "Home", Slug: "", Enabled: true, LastModifiedAt: &homeAt},
		{Name: "Draft", Slug: "draft", Enabled: false, LastModifiedAt: nil},
	}
	for i := range pages {
		if err := db.Create(&pages[i]).Error; err != nil {
			t.Fatalf("create page %s: %v", pages[i].Name, err)
		}
	}

	svc := NewSitePageService(repository.NewSitePageRepository(db), nil, nil)

	out, err := svc.Sitemap("https://example.com")
	if err != nil {
		t.Fatalf("Sitemap: %v", err)
	}

	if got, want := strings.HasPrefix(string(out), xml.Header), true; got != want {
		t.Fatalf("Sitemap output prefix = %v, want %v", got, want)
	}

	type sitemapEntry struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod"`
	}
	type sitemapDoc struct {
		XMLName xml.Name       `xml:"urlset"`
		URLs    []sitemapEntry `xml:"url"`
	}

	var doc sitemapDoc
	if err := xml.Unmarshal(out, &doc); err != nil {
		t.Fatalf("Unmarshal sitemap: %v", err)
	}

	if got, want := doc.XMLName.Space, "http://www.sitemaps.org/schemas/sitemap/0.9"; got != want {
		t.Fatalf("doc.XMLName.Space = %q, want %q", got, want)
	}

	if got, want := len(doc.URLs), 3; got != want {
		t.Fatalf("len(doc.URLs) = %d, want %d", got, want)
	}

	type wantEntry struct {
		Loc     string
		LastMod string
	}
	wants := []wantEntry{
		{"https://example.com/", "2026-01-02T00:04:05Z"},
		{"https://example.com/about", ""},
		{"https://example.com/zeta", "2026-03-04T05:06:07Z"},
	}
	for i, w := range wants {
		if got := doc.URLs[i].Loc; got != w.Loc {
			t.Fatalf("doc.URLs[%d].Loc = %q, want %q", i, got, w.Loc)
		}
		if got := doc.URLs[i].LastMod; got != w.LastMod {
			t.Fatalf("doc.URLs[%d].LastMod = %q, want %q", i, got, w.LastMod)
		}
	}

	if got, want := strings.Count(string(out), "<lastmod>"), 2; got != want {
		t.Fatalf("strings.Count(<lastmod>) = %d, want %d", got, want)
	}
}
