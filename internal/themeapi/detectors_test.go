package themeapi

import (
	"testing"

	"github.com/umono-cms/themecore"
)

// TestThemecoreDetectorCorpus runs themecore's detector corpus with the
// Compono version go.mod pins. A Compono bump that changes the signals the
// detectors read fails it.
func TestThemecoreDetectorCorpus(t *testing.T) {
	if err := themecore.DetectorCorpus(); err != nil {
		t.Fatal(err)
	}
}
