package runtime

import (
	"strings"
	"testing"
)

func TestOptimizeHTMLAddsFullViewportSizes(t *testing.T) {
	output, err := OptimizeHTML(`<compono-image><picture><source type="image/webp" srcset="/hero-640.webp 640w"><img src="/hero.jpg" alt="Hero" width="1280" height="720"></picture></compono-image>`)
	if err != nil {
		t.Fatalf("OptimizeHTML() error = %v", err)
	}

	if !strings.Contains(output, `sizes="100vw"`) {
		t.Fatalf("output = %q", output)
	}
}
