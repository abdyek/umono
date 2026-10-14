// Package csp builds the Content-Security-Policy of the public pages.
package csp

import (
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"strings"
)

// rejected are the characters an image origin must not contain: a wildcard,
// and the separators of a CSP header and of its source lists.
const rejected = ";,' \r\n*"

// Nonce returns a new script nonce: 16 random bytes from crypto/rand in
// standard base64.
func Nonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// ImageOrigins returns the sources img-src needs for the image URLs a render
// produced. Each URL is reduced to scheme://host[:port], or to host[:port] for
// a scheme-relative //host URL. Relative URLs, self, schemes other than http
// and https, and values with a rejected character are left out, never
// cleaned. The result is deduplicated and keeps the first-seen order. The
// URLs come from content, so nothing reaches the header without these rules.
func ImageOrigins(urls []string, self string) []string {
	result := []string{}
	self = strings.ToLower(self)
	seen := map[string]struct{}{}
	for _, raw := range urls {
		origin, ok := imageOrigin(raw)
		if !ok {
			continue
		}
		if origin == self {
			continue
		}
		if _, dup := seen[origin]; dup {
			continue
		}
		seen[origin] = struct{}{}
		result = append(result, origin)
	}
	return result
}

// imageOrigin returns the source of an image URL for img-src.
func imageOrigin(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	if strings.HasPrefix(raw, "//") {
		u, err := url.Parse("https:" + raw)
		if err != nil {
			return "", false
		}
		if u.Host == "" {
			return "", false
		}
		origin := strings.ToLower(u.Host)
		if strings.ContainsAny(origin, rejected) {
			return "", false
		}
		return origin, true
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", false
	}
	if u.Host == "" {
		return "", false
	}
	origin := scheme + "://" + strings.ToLower(u.Host)
	if strings.ContainsAny(origin, rejected) {
		return "", false
	}
	return origin, true
}

// Header returns the Content-Security-Policy of a page with the script nonce
// and the image origins.
func Header(nonce string, imageOrigins []string) string {
	img := "img-src 'self'"
	for _, o := range imageOrigins {
		img += " " + o
	}
	parts := []string{
		"default-src 'none'",
		"script-src 'nonce-" + nonce + "'",
		"style-src 'self'",
		img,
		"font-src 'self'",
		"base-uri 'none'",
		"form-action 'none'",
		"frame-ancestors 'none'",
	}
	return strings.Join(parts, "; ")
}
