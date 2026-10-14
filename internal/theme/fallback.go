// Package theme holds the themes Umono serves under /_theme.
package theme

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/umono-cms/themecore"
)

// Fallback is the fallback theme, the stylesheet of the pages without an
// active theme.
type Fallback struct {
	// Release is fallback- and the first 12 hex digits of the SHA-256 of CSS.
	// It changes with the content, so the stylesheet can be cached forever.
	Release string
	CSS     []byte
}

// NewFallback returns the fallback theme of themecore.
func NewFallback() Fallback {
	css := themecore.FallbackCSS()
	sum := sha256.Sum256(css)
	return Fallback{
		Release: "fallback-" + hex.EncodeToString(sum[:])[:12],
		CSS:     css,
	}
}
