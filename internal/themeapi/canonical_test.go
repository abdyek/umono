package themeapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCanonical(t *testing.T) {
	files, err := filepath.Glob("testdata/canonical/*.html")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("no test files found")
	}
	for _, file := range files {
		base := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		t.Run(base, func(t *testing.T) {
			input, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}
			expected, err := os.ReadFile(strings.TrimSuffix(file, ".html") + ".canon")
			if err != nil {
				t.Fatalf("read expected: %v", err)
			}
			got, err := Canonical(input)
			if err != nil {
				t.Fatalf("Canonical: %v", err)
			}
			if got != string(expected) {
				t.Fatalf("got %q, want %q", got, string(expected))
			}
		})
	}
}
