package themeapi

import (
	"testing"

	"github.com/umono-cms/themecore"
)

// TestThemecoreTOMLConformance checks that the TOML decoder this binary
// links rejects every input outside TOML v1.0.0. Theme files are TOML
// v1.0.0, and the decoder version Go selects for Umono can differ from
// themecore's.
func TestThemecoreTOMLConformance(t *testing.T) {
	if err := themecore.TOMLConformance(); err != nil {
		t.Fatal(err)
	}
}
