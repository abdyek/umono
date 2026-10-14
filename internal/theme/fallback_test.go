package theme

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/umono-cms/themecore"
)

func TestNewFallback(t *testing.T) {
	f := NewFallback()

	if !bytes.Equal(f.CSS, themecore.FallbackCSS()) {
		t.Fatalf("got CSS %q, want %q", string(f.CSS), string(themecore.FallbackCSS()))
	}

	sum := sha256.Sum256(themecore.FallbackCSS())
	wantRelease := "fallback-" + hex.EncodeToString(sum[:])[:12]
	if got := f.Release; got != wantRelease {
		t.Fatalf("got Release %q, want %q", got, wantRelease)
	}

	if got, want := len(f.Release), 21; got != want {
		t.Fatalf("got Release length %d, want %d", got, want)
	}
}
