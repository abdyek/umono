package service

import (
	"encoding/xml"
	"time"
)

// sitemapNamespace is the XML namespace of the sitemap protocol.
const sitemapNamespace = "http://www.sitemaps.org/schemas/sitemap/0.9"

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

// Sitemap returns the sitemap.xml of the enabled pages under siteURL, sorted
// by slug. The address of a page is siteURL, a slash and its slug, so the home
// page is siteURL with a trailing slash. lastmod is the LastModifiedAt of the
// page in UTC and is left out when the page has none.
func (s *SitePageService) Sitemap(siteURL string) ([]byte, error) {
	set := sitemapURLSet{Xmlns: sitemapNamespace}
	for _, page := range s.repo.GetEnabled() {
		u := sitemapURL{Loc: siteURL + "/" + page.Slug}
		if page.LastModifiedAt != nil {
			u.LastMod = page.LastModifiedAt.UTC().Format(time.RFC3339)
		}
		set.URLs = append(set.URLs, u)
	}
	out, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), out...), nil
}
