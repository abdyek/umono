package theme

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/umono-cms/themecore"
)

func writeTheme(t *testing.T, mutate func(files map[string][]byte)) string {
	t.Helper()
	files := map[string][]byte{}
	for _, name := range []string{"built.toml", "contract.toml", "attributes.toml", "LICENSE"} {
		b, err := os.ReadFile(filepath.Join("testdata", "package", name))
		if err != nil {
			t.Fatalf("read fixture %s: %v", name, err)
		}
		files[name] = b
	}
	files["assets/theme.css"] = themecore.FallbackCSS()
	if mutate != nil {
		mutate(files)
	}
	sum, err := themecore.Checksums(themecore.Package{Files: files})
	if err != nil {
		t.Fatalf("checksums: %v", err)
	}
	files["checksums.json"] = sum
	themesDir := t.TempDir()
	for name, b := range files {
		full := filepath.Join(themesDir, "acme", "minimal", "0.1.0", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, b, 0o644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}
	return themesDir
}

func TestCheckInstalled(t *testing.T) {
	dir := writeTheme(t, nil)
	got, err := CheckInstalled(dir, "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("CheckInstalled: %v", err)
	}
	if got.Path != "acme/minimal/0.1.0" {
		t.Fatalf("Path = %q, want %q", got.Path, "acme/minimal/0.1.0")
	}
	checksumsPath := filepath.Join(dir, "acme", "minimal", "0.1.0", "checksums.json")
	checksumsOnDisk, err := os.ReadFile(checksumsPath)
	if err != nil {
		t.Fatalf("read checksums.json: %v", err)
	}
	sum := sha256.Sum256(checksumsOnDisk)
	wantRelease := hex.EncodeToString(sum[:])[:12]
	if got.Release != wantRelease {
		t.Fatalf("Release = %q, want %q", got.Release, wantRelease)
	}
	if got.Built.ID != "acme/minimal" {
		t.Fatalf("Built.ID = %q, want %q", got.Built.ID, "acme/minimal")
	}
	if got.Built.Contract != "example/blog@v1.0" {
		t.Fatalf("Built.Contract = %q, want %q", got.Built.Contract, "example/blog@v1.0")
	}
	contractOnDisk, err := os.ReadFile(filepath.Join("testdata", "package", "contract.toml"))
	if err != nil {
		t.Fatalf("read contract fixture: %v", err)
	}
	if !bytes.Equal(got.ContractSource, contractOnDisk) {
		t.Fatalf("ContractSource = %q, want %q", got.ContractSource, contractOnDisk)
	}
	if got.Result.Contract.ID != "example/blog" {
		t.Fatalf("Result.Contract.ID = %q, want %q", got.Result.Contract.ID, "example/blog")
	}
	wantAttrs := map[string]string{"class": "layout-post"}
	if !reflect.DeepEqual(got.Result.Attributes.Root("post"), wantAttrs) {
		t.Fatalf("Attributes.Root(post) = %v, want %v", got.Result.Attributes.Root("post"), wantAttrs)
	}
}

func TestCheckInstalledRejectsChangedFile(t *testing.T) {
	dir := writeTheme(t, nil)
	licensePath := filepath.Join(dir, "acme", "minimal", "0.1.0", "LICENSE")
	original, err := os.ReadFile(licensePath)
	if err != nil {
		t.Fatalf("read LICENSE: %v", err)
	}
	changed := append([]byte{}, original...)
	changed = append(changed, []byte("changed\n")...)
	if err := os.WriteFile(licensePath, changed, 0o644); err != nil {
		t.Fatalf("rewrite LICENSE: %v", err)
	}
	if _, err := CheckInstalled(dir, "acme/minimal/0.1.0"); err == nil {
		t.Fatalf("CheckInstalled: want error, got nil")
	}
}

func TestCheckInstalledRejectsThemeAPI(t *testing.T) {
	cases := []int{themecore.ThemeAPICurrent() + 1}
	if themecore.ThemeAPIMin() > 1 {
		cases = append(cases, themecore.ThemeAPIMin()-1)
	}
	for _, v := range cases {
		v := v
		t.Run(fmt.Sprintf("version=%d", v), func(t *testing.T) {
			dir := writeTheme(t, func(files map[string][]byte) {
				files["built.toml"] = bytes.Replace(files["built.toml"], []byte("theme-api-version = 1"), []byte(fmt.Sprintf("theme-api-version = %d", v)), 1)
			})
			if _, err := CheckInstalled(dir, "acme/minimal/0.1.0"); !errors.Is(err, ErrThemeAPI) {
				t.Fatalf("CheckInstalled: errors.Is(err, ErrThemeAPI) = false; err = %v", err)
			}
		})
	}
}

func TestCheckInstalledRunsTheStaticCheck(t *testing.T) {
	dir := writeTheme(t, func(files map[string][]byte) {
		delete(files, "assets/theme.css")
	})
	_, err := CheckInstalled(dir, "acme/minimal/0.1.0")
	if err == nil {
		t.Fatalf("CheckInstalled: want error, got nil")
	}
	if errors.Is(err, ErrThemeAPI) {
		t.Fatalf("CheckInstalled: errors.Is(err, ErrThemeAPI) = true; err = %v", err)
	}
}

func TestCheckInstalledRejectsInvalidPath(t *testing.T) {
	dir := writeTheme(t, nil)
	paths := []string{
		"../acme/minimal/0.1.0",
		"acme/minimal",
		"acme/minimal/0.1.0/assets",
		"Acme/minimal/0.1.0",
		"acme/minimal/v0.1.0",
		"acme/minimal/0.1",
		"/acme/minimal/0.1.0",
		"acme//minimal/0.1.0",
		"",
	}
	for _, p := range paths {
		p := p
		t.Run(fmt.Sprintf("path=%q", p), func(t *testing.T) {
			if _, err := CheckInstalled(dir, p); !errors.Is(err, ErrInvalidThemePath) {
				t.Fatalf("CheckInstalled(%q): errors.Is(err, ErrInvalidThemePath) = false; err = %v", p, err)
			}
		})
	}
}

func TestCheckInstalledMissingTheme(t *testing.T) {
	dir := writeTheme(t, nil)
	if _, err := CheckInstalled(dir, "acme/minimal/0.2.0"); err == nil {
		t.Fatalf("CheckInstalled: want error, got nil")
	}
}
