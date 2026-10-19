package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"github.com/umono-cms/umono/internal/theme"
)

// writeDevTheme writes a theme project (root) and its dev directory,
// themes/dev/dev/0.0.0/, under themesDir as the CLI creates them. built.toml
// is the package fixture's with the id dev/dev and the version 0.0.0;
// contract.toml, attributes.toml and assets are symlinks into root. built,
// when not nil, changes built.toml before it is written.
func writeDevTheme(t *testing.T, themesDir string, built func(string) string) (root string) {
	t.Helper()
	root = t.TempDir()

	contractPath := filepath.Join(root, ".dev", "contracts", "example", "blog@v1.0.toml")
	if err := os.MkdirAll(filepath.Dir(contractPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(contractPath), err)
	}
	contractBytes, err := os.ReadFile(filepath.Join("..", "theme", "testdata", "package", "contract.toml"))
	if err != nil {
		t.Fatalf("read fixture contract.toml: %v", err)
	}
	if err := os.WriteFile(contractPath, contractBytes, 0o644); err != nil {
		t.Fatalf("write %s: %v", contractPath, err)
	}

	attributesPath := filepath.Join(root, "attributes.toml")
	attributesBytes, err := os.ReadFile(filepath.Join("..", "theme", "testdata", "package", "attributes.toml"))
	if err != nil {
		t.Fatalf("read fixture attributes.toml: %v", err)
	}
	if err := os.WriteFile(attributesPath, attributesBytes, 0o644); err != nil {
		t.Fatalf("write %s: %v", attributesPath, err)
	}

	builtBytes, err := os.ReadFile(filepath.Join("..", "theme", "testdata", "package", "built.toml"))
	if err != nil {
		t.Fatalf("read fixture built.toml: %v", err)
	}
	b := string(builtBytes)
	b = strings.Replace(b, "acme/minimal", "dev/dev", 1)
	b = strings.Replace(b, "0.1.0", "0.0.0", 1)
	if built != nil {
		b = built(b)
	}

	dir := filepath.Join(themesDir, "dev", "dev", "0.0.0")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	builtPath := filepath.Join(dir, "built.toml")
	if err := os.WriteFile(builtPath, []byte(b), 0o644); err != nil {
		t.Fatalf("write %s: %v", builtPath, err)
	}

	contractLink := filepath.Join(dir, "contract.toml")
	if err := os.Symlink(filepath.Join(root, ".dev", "contracts", "example", "blog@v1.0.toml"), contractLink); err != nil {
		t.Skipf("symlink: %v", err)
	}
	attributesLink := filepath.Join(dir, "attributes.toml")
	if err := os.Symlink(filepath.Join(root, "attributes.toml"), attributesLink); err != nil {
		t.Skipf("symlink: %v", err)
	}
	assetsLink := filepath.Join(dir, "assets")
	if err := os.Symlink(filepath.Join(root, "assets"), assetsLink); err != nil {
		t.Skipf("symlink: %v", err)
	}

	return root
}

// devRef is the contract of the dev theme's fixture.
func devRef(t *testing.T) models.Contract {
	t.Helper()
	ref, err := models.ParseContractRef("example/blog@v1.0")
	if err != nil {
		t.Fatalf("parse contract ref: %v", err)
	}
	return ref
}

func TestLoadDevAddsTheContractRecordWithoutTheRegistry(t *testing.T) {
	ctx := context.Background()
	reg := &fakeRegistry{}
	svc, db := newThemeTestService(t, reg)

	root := writeDevTheme(t, svc.themesDir, nil)
	svc.EnableDev(root)
	if err := svc.LoadDev(ctx); err != nil {
		t.Fatalf("LoadDev: %v", err)
	}

	if reg.calls != 0 {
		t.Fatalf("reg.calls = %d, want 0", reg.calls)
	}

	ref := devRef(t)
	record, ok := svc.contracts.Record(ref)
	if !ok {
		t.Fatalf("contracts.Record(devRef) ok = false, want true")
	}
	if record.Source != string(contractSource(t)) {
		t.Fatalf("record.Source = %q, want %q", record.Source, string(contractSource(t)))
	}

	l, ok := svc.Loaded(ref.Group())
	if !ok {
		t.Fatalf("Loaded(devRef.Group()) ok = false, want true")
	}
	if l.Path != theme.DevPath {
		t.Fatalf("l.Path = %q, want %q", l.Path, theme.DevPath)
	}
	if l.Release != theme.DevRelease {
		t.Fatalf("l.Release = %q, want %q", l.Release, theme.DevRelease)
	}
	if l.ContractID != record.ID {
		t.Fatalf("l.ContractID = %d, want %d", l.ContractID, record.ID)
	}

	active, err := repository.NewThemeRepository(db).GetActiveThemes()
	if err != nil {
		t.Fatalf("GetActiveThemes: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("len(active) = %d, want 0", len(active))
	}
}

func TestLoadDevThemeAPIOutOfRange(t *testing.T) {
	svc, _ := newThemeTestService(t, &fakeRegistry{})
	root := writeDevTheme(t, svc.themesDir, func(b string) string {
		return strings.Replace(b, "theme-api-version = 1", "theme-api-version = 2", 1)
	})
	svc.EnableDev(root)

	if err := svc.LoadDev(context.Background()); !errors.Is(err, theme.ErrThemeAPI) {
		t.Fatalf("LoadDev error = %v, want theme.ErrThemeAPI", err)
	}

	ref := devRef(t)
	if _, ok := svc.Loaded(ref.Group()); ok {
		t.Fatalf("Loaded(devRef.Group()) ok = true, want false")
	}
	if _, ok := svc.contracts.Record(ref); ok {
		t.Fatalf("contracts.Record(devRef) ok = true, want false")
	}
}

func TestLoadDevWithoutDevMode(t *testing.T) {
	svc, _ := newThemeTestService(t, &fakeRegistry{})
	writeDevTheme(t, svc.themesDir, nil)

	if err := svc.LoadDev(context.Background()); err != nil {
		t.Fatalf("LoadDev: %v", err)
	}

	ref := devRef(t)
	if _, ok := svc.Loaded(ref.Group()); ok {
		t.Fatalf("Loaded(devRef.Group()) ok = true, want false")
	}
	if _, ok := svc.contracts.Record(ref); ok {
		t.Fatalf("contracts.Record(devRef) ok = true, want false")
	}
}

func TestLoadDevChangedContract(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})
	root := writeDevTheme(t, svc.themesDir, nil)
	svc.EnableDev(root)

	if err := db.Create(&models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0, Source: "s", Checksum: "x"}).Error; err != nil {
		t.Fatalf("create contract: %v", err)
	}

	if err := svc.LoadDev(context.Background()); !errors.Is(err, ErrContractChanged) {
		t.Fatalf("LoadDev error = %v, want ErrContractChanged", err)
	}

	g := devRef(t).Group()
	b, ok := svc.Broken(g)
	if !ok {
		t.Fatalf("Broken(devRef.Group()) ok = false, want true")
	}
	if b.Path != theme.DevPath {
		t.Fatalf("b.Path = %q, want %q", b.Path, theme.DevPath)
	}
	if _, ok := svc.Loaded(g); ok {
		t.Fatalf("Loaded(devRef.Group()) ok = true, want false")
	}
}

func TestLoadDevNotCovered(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})
	root := writeDevTheme(t, svc.themesDir, nil)
	svc.EnableDev(root)

	newer := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 5, Source: "s", Checksum: "x"}
	if err := db.Create(&newer).Error; err != nil {
		t.Fatalf("create contract: %v", err)
	}

	layout := "post"
	if err := db.Create(&models.SitePage{Name: "A", Slug: "a", Enabled: true, ContractID: &newer.ID, LayoutID: &layout}).Error; err != nil {
		t.Fatalf("create page: %v", err)
	}

	if err := svc.LoadDev(context.Background()); !errors.Is(err, ErrNotCovered) {
		t.Fatalf("LoadDev error = %v, want ErrNotCovered", err)
	}
	if _, ok := svc.Loaded(devRef(t).Group()); ok {
		t.Fatalf("Loaded(devRef.Group()) ok = true, want false")
	}
}

func TestDevGroupRefusesActivateDeactivateAndRetry(t *testing.T) {
	ctx := context.Background()
	svc, _ := newThemeTestService(t, &fakeRegistry{})
	root := writeDevTheme(t, svc.themesDir, nil)
	svc.EnableDev(root)
	if err := svc.LoadDev(ctx); err != nil {
		t.Fatalf("LoadDev: %v", err)
	}

	g := devRef(t).Group()
	if _, err := svc.Activate(ctx, "acme/minimal/0.1.0"); !errors.Is(err, ErrDevTheme) {
		t.Fatalf("Activate error = %v, want ErrDevTheme", err)
	}
	if err := svc.Deactivate(g); !errors.Is(err, ErrDevTheme) {
		t.Fatalf("Deactivate error = %v, want ErrDevTheme", err)
	}
	if err := svc.Retry(ctx, g); !errors.Is(err, ErrDevTheme) {
		t.Fatalf("Retry error = %v, want ErrDevTheme", err)
	}

	l, ok := svc.Loaded(g)
	if !ok {
		t.Fatalf("Loaded(devRef.Group()) ok = false, want true")
	}
	if l.Path != theme.DevPath {
		t.Fatalf("l.Path = %q, want %q", l.Path, theme.DevPath)
	}
}

func TestDevThemeWinsOverActiveThemes(t *testing.T) {
	ctx := context.Background()
	svc, db := newThemeTestService(t, &fakeRegistry{})

	if _, err := svc.Activate(ctx, "acme/minimal/0.1.0"); err != nil {
		t.Fatalf("Activate: %v", err)
	}

	root := writeDevTheme(t, svc.themesDir, nil)
	svc2, _ := restart(t, db, svc.themesDir)
	if err := svc2.LoadActive(ctx); err != nil {
		t.Fatalf("LoadActive: %v", err)
	}
	svc2.EnableDev(root)
	if err := svc2.LoadDev(ctx); err != nil {
		t.Fatalf("LoadDev: %v", err)
	}

	l, ok := svc2.Loaded(devRef(t).Group())
	if !ok {
		t.Fatalf("Loaded(devRef.Group()) ok = false, want true")
	}
	if l.Path != theme.DevPath {
		t.Fatalf("l.Path = %q, want %q", l.Path, theme.DevPath)
	}

	active, err := repository.NewThemeRepository(db).GetActiveThemes()
	if err != nil {
		t.Fatalf("GetActiveThemes: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("len(active) = %d, want 1", len(active))
	}
	if active[0].Path != "acme/minimal/0.1.0" {
		t.Fatalf("active[0].Path = %q, want %q", active[0].Path, "acme/minimal/0.1.0")
	}
}

func TestActivateDevWithoutDevModeFails(t *testing.T) {
	svc, _ := newThemeTestService(t, &fakeRegistry{})
	writeDevTheme(t, svc.themesDir, nil)

	if _, err := svc.Activate(context.Background(), theme.DevPath); err == nil {
		t.Fatalf("Activate error = nil, want non-nil")
	}
}
