package theme

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func sameList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func readOnly(t *testing.T, root string) {
	t.Helper()
	var files []string
	var dirs []string
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	for _, f := range files {
		if err := os.Chmod(f, 0o444); err != nil {
			t.Fatalf("chmod %s: %v", f, err)
		}
	}
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})
	for _, d := range dirs {
		if err := os.Chmod(d, 0o555); err != nil {
			t.Fatalf("chmod %s: %v", d, err)
		}
	}
}

func TestListPending(t *testing.T) {
	themes := writeTheme(t, nil)
	pending := writeTheme(t, func(files map[string][]byte) {
		files["LICENSE"] = append(append([]byte{}, files["LICENSE"]...), "x\n"...)
		files["assets/logo.svg"] = []byte("<svg/>")
		delete(files, "attributes.toml")
		files["built.toml"] = bytes.Replace(files["built.toml"], []byte("2026-10-14T00:00:00Z"), []byte("2026-10-15T00:00:00Z"), 1)
	})
	list, err := ListPending(themes, pending)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(list))
	}
	p := list[0]
	if p.Path != "acme/minimal/0.1.0" {
		t.Fatalf("Path = %q, want %q", p.Path, "acme/minimal/0.1.0")
	}
	if !p.Installed {
		t.Fatalf("Installed = false, want true")
	}
	if p.Err != nil {
		t.Fatalf("Err = %v, want nil", p.Err)
	}
	wantOld := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	if !p.OldBuilt.BuiltAt.Equal(wantOld) {
		t.Fatalf("OldBuilt.BuiltAt = %v, want %v", p.OldBuilt.BuiltAt, wantOld)
	}
	wantNew := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	if !p.NewBuilt.BuiltAt.Equal(wantNew) {
		t.Fatalf("NewBuilt.BuiltAt = %v, want %v", p.NewBuilt.BuiltAt, wantNew)
	}
	if !sameList(p.Changed, []string{"LICENSE", "built.toml"}) {
		t.Fatalf("Changed = %v, want %v", p.Changed, []string{"LICENSE", "built.toml"})
	}
	if !sameList(p.Added, []string{"assets/logo.svg"}) {
		t.Fatalf("Added = %v, want %v", p.Added, []string{"assets/logo.svg"})
	}
	if !sameList(p.Removed, []string{"attributes.toml"}) {
		t.Fatalf("Removed = %v, want %v", p.Removed, []string{"attributes.toml"})
	}
}

func TestListPendingWithoutInstalledTheme(t *testing.T) {
	themes := t.TempDir()
	pending := writeTheme(t, nil)
	list, err := ListPending(themes, pending)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(list))
	}
	p := list[0]
	if p.Installed {
		t.Fatalf("Installed = true, want false")
	}
	if p.Err != nil {
		t.Fatalf("Err = %v, want nil", p.Err)
	}
	if !p.OldBuilt.BuiltAt.IsZero() {
		t.Fatalf("OldBuilt.BuiltAt = %v, want zero", p.OldBuilt.BuiltAt)
	}
	if p.NewBuilt.ID != "acme/minimal" {
		t.Fatalf("NewBuilt.ID = %q, want %q", p.NewBuilt.ID, "acme/minimal")
	}
	if len(p.Changed)+len(p.Added)+len(p.Removed) != 0 {
		t.Fatalf("lists not empty: Changed=%v Added=%v Removed=%v", p.Changed, p.Added, p.Removed)
	}
}

func TestListPendingWithoutPendingDir(t *testing.T) {
	list, err := ListPending(t.TempDir(), filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("len(list) = %d, want 0", len(list))
	}
}

func TestListPendingUnreadable(t *testing.T) {
	themes := writeTheme(t, nil)
	pending := writeTheme(t, nil)
	builtPath := filepath.Join(pending, "acme", "minimal", "0.1.0", "built.toml")
	if err := os.WriteFile(builtPath, []byte("not toml"), 0o644); err != nil {
		t.Fatalf("rewrite built.toml: %v", err)
	}
	list, err := ListPending(themes, pending)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(list))
	}
	p := list[0]
	if p.Path != "acme/minimal/0.1.0" {
		t.Fatalf("Path = %q, want %q", p.Path, "acme/minimal/0.1.0")
	}
	if p.Err == nil {
		t.Fatalf("Err = nil, want non-nil")
	}
}

func TestListPendingSkipsOtherDirectories(t *testing.T) {
	themes := writeTheme(t, nil)
	pending := writeTheme(t, nil)
	for _, d := range []string{
		filepath.Join(pending, "Acme", "x", "1.0.0"),
		filepath.Join(pending, "acme", "minimal", "latest"),
		filepath.Join(pending, "acme", "minimal", "0.2"),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	list, err := ListPending(themes, pending)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(list))
	}
	if list[0].Path != "acme/minimal/0.1.0" {
		t.Fatalf("Path = %q, want %q", list[0].Path, "acme/minimal/0.1.0")
	}
}

func TestRejectPending(t *testing.T) {
	pending := writeTheme(t, nil)
	readOnly(t, filepath.Join(pending, "acme", "minimal", "0.1.0"))
	if err := RejectPending(pending, "acme/minimal/0.1.0"); err != nil {
		t.Fatalf("RejectPending: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pending, "acme")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("os.Stat(acme): err = %v, want fs.ErrNotExist", err)
	}
	if _, err := os.Stat(pending); err != nil {
		t.Fatalf("os.Stat(pending): %v", err)
	}
}

func TestRejectPendingKeepsOtherVersions(t *testing.T) {
	pending := writeTheme(t, nil)
	if err := os.MkdirAll(filepath.Join(pending, "acme", "minimal", "0.2.0"), 0o755); err != nil {
		t.Fatalf("mkdir 0.2.0: %v", err)
	}
	if err := RejectPending(pending, "acme/minimal/0.1.0"); err != nil {
		t.Fatalf("RejectPending: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pending, "acme", "minimal", "0.2.0")); err != nil {
		t.Fatalf("os.Stat(0.2.0): %v", err)
	}
	if _, err := os.Stat(filepath.Join(pending, "acme", "minimal", "0.1.0")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("os.Stat(0.1.0): err = %v, want fs.ErrNotExist", err)
	}
}

func TestRejectPendingRejectsInvalidPath(t *testing.T) {
	pending := writeTheme(t, nil)
	for _, p := range []string{"../acme/minimal/0.1.0", "acme/minimal", ""} {
		if err := RejectPending(pending, p); !errors.Is(err, ErrInvalidThemePath) {
			t.Fatalf("RejectPending(%q): err = %v, want ErrInvalidThemePath", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(pending, "acme", "minimal", "0.1.0")); err != nil {
		t.Fatalf("os.Stat(0.1.0): %v", err)
	}
}
