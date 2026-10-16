package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListInstalled(t *testing.T) {
	dir := writeTheme(t, nil)
	for _, p := range []string{
		filepath.Join(dir, "acme", "minimal", "latest"),
		filepath.Join(dir, "Acme", "x", "1.0.0"),
	} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
	}
	brokenDir := filepath.Join(dir, "acme", "broken", "1.0.0")
	if err := os.MkdirAll(brokenDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", brokenDir, err)
	}
	if err := os.WriteFile(filepath.Join(brokenDir, "built.toml"), []byte("not toml"), 0o644); err != nil {
		t.Fatalf("write broken built.toml: %v", err)
	}
	list, err := ListInstalled(dir)
	if err != nil {
		t.Fatalf("ListInstalled: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(list) = %d, want 2", len(list))
	}
	if list[0].Path != "acme/broken/1.0.0" {
		t.Fatalf("list[0].Path = %q, want %q", list[0].Path, "acme/broken/1.0.0")
	}
	if list[0].Err == nil {
		t.Fatalf("list[0].Err = nil, want non-nil")
	}
	if list[1].Path != "acme/minimal/0.1.0" {
		t.Fatalf("list[1].Path = %q, want %q", list[1].Path, "acme/minimal/0.1.0")
	}
	if list[1].Err != nil {
		t.Fatalf("list[1].Err = %v, want nil", list[1].Err)
	}
	if list[1].Built.Contract != "example/blog@v1.0" {
		t.Fatalf("list[1].Built.Contract = %q, want %q", list[1].Built.Contract, "example/blog@v1.0")
	}
}

func TestListInstalledWithoutThemesDir(t *testing.T) {
	list, err := ListInstalled(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatalf("ListInstalled: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("len(list) = %d, want 0", len(list))
	}
}
