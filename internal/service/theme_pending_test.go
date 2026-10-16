package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/umono-cms/themecore"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"github.com/umono-cms/umono/internal/theme"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPendingTestService(t *testing.T, reg *fakeRegistry, mutate func(files map[string][]byte)) (svc *ThemeService, db *gorm.DB, themesDir string, pendingDir string) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Contract{}, &models.SitePage{}, &models.Option{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	themesDir = writeTestTheme(t)
	pendingDir = writeTestThemeWith(t, mutate)
	svc = NewThemeService(themesDir, pendingDir, NewContractService(repository.NewContractRepository(db)), repository.NewThemeRepository(db), reg)
	t.Cleanup(func() {
		makeWritable(themesDir)
		makeWritable(pendingDir)
	})
	return svc, db, themesDir, pendingDir
}

func makeWritable(root string) {
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			_ = os.Chmod(path, 0o755)
		}
		return nil
	})
}

func readOnlyTree(t *testing.T, root string) {
	t.Helper()
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			_ = os.Chmod(path, 0o444)
		}
		return nil
	})
	var dirs []string
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})
	for _, d := range dirs {
		_ = os.Chmod(d, 0o555)
	}
}

func readThemeFile(t *testing.T, root, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "acme", "minimal", "0.1.0", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return b
}

func changeLicense(files map[string][]byte) {
	files["LICENSE"] = append(append([]byte{}, files["LICENSE"]...), "changed\n"...)
}

func contractVersion(v string) func(files map[string][]byte) {
	return func(files map[string][]byte) {
		files["contract.toml"] = bytes.Replace(files["contract.toml"], []byte(`version = "v1.0"`), []byte(`version = "`+v+`"`), 1)
		files["built.toml"] = bytes.Replace(files["built.toml"], []byte(`contract = "example/blog@v1.0"`), []byte(`contract = "example/blog@`+v+`"`), 1)
	}
}

func TestApplyPendingToAnInactiveTheme(t *testing.T) {
	ctx := context.Background()
	path := "acme/minimal/0.1.0"
	reg := &fakeRegistry{}
	svc, _, themesDir, pendingDir := newPendingTestService(t, reg, changeLicense)

	want := readThemeFile(t, pendingDir, "LICENSE")
	readOnlyTree(t, filepath.Join(themesDir, "acme", "minimal", "0.1.0"))
	readOnlyTree(t, filepath.Join(pendingDir, "acme", "minimal", "0.1.0"))

	if err := svc.ApplyPending(ctx, path); err != nil {
		t.Fatalf("ApplyPending: %v", err)
	}
	if !bytes.Equal(readThemeFile(t, themesDir, "LICENSE"), want) {
		t.Fatalf("themesDir/LICENSE = %q, want %q", readThemeFile(t, themesDir, "LICENSE"), want)
	}
	if reg.calls != 0 {
		t.Fatalf("reg.calls = %d, want 0", reg.calls)
	}
	if _, err := os.Stat(filepath.Join(pendingDir, "acme")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("os.Stat(pending/acme): err = %v, want fs.ErrNotExist", err)
	}
	entries, err := os.ReadDir(filepath.Join(themesDir, "acme", "minimal"))
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "0.1.0" {
		t.Fatalf("entries = %v, want only one named %q", entries, "0.1.0")
	}
}

func TestApplyPendingToTheActiveTheme(t *testing.T) {
	ctx := context.Background()
	path := "acme/minimal/0.1.0"
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	reg := &fakeRegistry{}
	svc, db, _, pendingDir := newPendingTestService(t, reg, changeLicense)

	pendingChecksums, err := os.ReadFile(filepath.Join(pendingDir, "acme", "minimal", "0.1.0", "checksums.json"))
	if err != nil {
		t.Fatalf("read pending checksums.json: %v", err)
	}
	sum := sha256.Sum256(pendingChecksums)

	calls := 0
	svc.OnRenderInputChange(func(ctx context.Context) {
		calls++
	})

	if _, err := svc.Activate(ctx, path); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	before := activeThemes(t, db)

	if err := svc.ApplyPending(ctx, path); err != nil {
		t.Fatalf("ApplyPending: %v", err)
	}

	l, ok := svc.Loaded(blog)
	if !ok {
		t.Fatalf("svc.Loaded(blog) = false, want true")
	}
	if l.Release != hex.EncodeToString(sum[:])[:12] {
		t.Fatalf("l.Release = %q, want %q", l.Release, hex.EncodeToString(sum[:])[:12])
	}
	if reg.calls != 1 {
		t.Fatalf("reg.calls = %d, want 1", reg.calls)
	}
	if got := activeThemes(t, db); got != before {
		t.Fatalf("activeThemes = %q, want %q", got, before)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestApplyPendingChangesTheContractMinor(t *testing.T) {
	ctx := context.Background()
	path := "acme/minimal/0.1.0"
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	reg := &fakeRegistry{}
	svc, db, _, _ := newPendingTestService(t, reg, contractVersion("v1.1"))

	calls := 0
	svc.OnRenderInputChange(func(ctx context.Context) {
		calls++
	})

	if _, err := svc.Activate(ctx, path); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if err := svc.ApplyPending(ctx, path); err != nil {
		t.Fatalf("ApplyPending: %v", err)
	}
	if reg.calls != 2 {
		t.Fatalf("reg.calls = %d, want 2", reg.calls)
	}
	if n := countContracts(t, db); n != 2 {
		t.Fatalf("countContracts = %d, want 2", n)
	}
	wantThemes := `[{"path":"acme/minimal/0.1.0","contract_id":"example/blog@v1.1"}]`
	if got := activeThemes(t, db); got != wantThemes {
		t.Fatalf("activeThemes = %q, want %q", got, wantThemes)
	}

	var stored models.Contract
	if err := db.Where(&models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 1}).First(&stored).Error; err != nil {
		t.Fatalf("load v1.1 contract: %v", err)
	}
	l, ok := svc.Loaded(blog)
	if !ok {
		t.Fatalf("svc.Loaded(blog) = false, want true")
	}
	if l.ContractID != stored.ID {
		t.Fatalf("l.ContractID = %d, want %d", l.ContractID, stored.ID)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestApplyPendingRejectsAnotherGroup(t *testing.T) {
	ctx := context.Background()
	path := "acme/minimal/0.1.0"
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	reg := &fakeRegistry{}
	svc, _, themesDir, pendingDir := newPendingTestService(t, reg, contractVersion("v2.0"))

	if _, err := svc.Activate(ctx, path); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	l0, ok := svc.Loaded(blog)
	if !ok {
		t.Fatalf("svc.Loaded(blog) after Activate = false, want true")
	}
	old := readThemeFile(t, themesDir, "contract.toml")

	err := svc.ApplyPending(ctx, path)
	if !errors.Is(err, ErrGroupChanged) {
		t.Fatalf("ApplyPending: err = %v, want ErrGroupChanged", err)
	}
	if !bytes.Equal(readThemeFile(t, themesDir, "contract.toml"), old) {
		t.Fatalf("themesDir/contract.toml changed after rejection")
	}
	if _, err := os.Stat(filepath.Join(pendingDir, "acme", "minimal", "0.1.0")); err != nil {
		t.Fatalf("os.Stat(pending/.../0.1.0): %v", err)
	}
	l1, ok := svc.Loaded(blog)
	if !ok {
		t.Fatalf("svc.Loaded(blog) after ApplyPending = false, want true")
	}
	if l1.Release != l0.Release {
		t.Fatalf("l1.Release = %q, want %q", l1.Release, l0.Release)
	}
}

func TestApplyPendingWithoutInstalledTheme(t *testing.T) {
	ctx := context.Background()
	path := "acme/minimal/0.1.0"
	reg := &fakeRegistry{}
	svc, _, themesDir, pendingDir := newPendingTestService(t, reg, changeLicense)

	if err := os.RemoveAll(filepath.Join(themesDir, "acme")); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}
	if err := svc.ApplyPending(ctx, path); !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("ApplyPending: err = %v, want ErrNotInstalled", err)
	}
	if _, err := os.Stat(filepath.Join(pendingDir, "acme", "minimal", "0.1.0")); err != nil {
		t.Fatalf("os.Stat(pending/.../0.1.0): %v", err)
	}
}

func TestApplyPendingRunsThePackageChecks(t *testing.T) {
	ctx := context.Background()
	path := "acme/minimal/0.1.0"
	reg := &fakeRegistry{}
	svc, _, themesDir, pendingDir := newPendingTestService(t, reg, func(files map[string][]byte) {
		files["built.toml"] = bytes.Replace(files["built.toml"], []byte("theme-api-version = 1"), []byte(fmt.Sprintf("theme-api-version = %d", themecore.ThemeAPICurrent()+1)), 1)
	})

	old := readThemeFile(t, themesDir, "built.toml")

	if err := svc.ApplyPending(ctx, path); !errors.Is(err, theme.ErrThemeAPI) {
		t.Fatalf("ApplyPending: err = %v, want theme.ErrThemeAPI", err)
	}
	if !bytes.Equal(readThemeFile(t, themesDir, "built.toml"), old) {
		t.Fatalf("themesDir/built.toml changed after rejection")
	}
	if _, err := os.Stat(filepath.Join(pendingDir, "acme", "minimal", "0.1.0")); err != nil {
		t.Fatalf("os.Stat(pending/.../0.1.0): %v", err)
	}
}
