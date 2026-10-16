package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/umono-cms/themecore"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func findGroup(t *testing.T, list ThemeList, g models.ContractGroup) ThemeGroup {
	t.Helper()
	for _, item := range list.Groups {
		if item.Group == g {
			return item
		}
	}
	t.Fatalf("group %q not found in list", g)
	return ThemeGroup{}
}

func TestThemeListInstalledTheme(t *testing.T) {
	svc, _, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Groups) != 1 {
		t.Fatalf("len(list.Groups) = %d, want 1", len(list.Groups))
	}
	blogGroup := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	g := findGroup(t, list, blogGroup)
	if g.PageCount != 0 {
		t.Fatalf("g.PageCount = %d, want 0", g.PageCount)
	}
	if g.HasBase {
		t.Fatalf("g.HasBase = true, want false")
	}
	if g.Active != nil {
		t.Fatalf("g.Active = %+v, want nil", g.Active)
	}
	if len(g.Records) != 0 {
		t.Fatalf("len(g.Records) = %d, want 0", len(g.Records))
	}
	if len(g.Themes) != 1 {
		t.Fatalf("len(g.Themes) = %d, want 1", len(g.Themes))
	}
	if g.Themes[0].Path != "acme/minimal/0.1.0" {
		t.Fatalf("g.Themes[0].Path = %q, want %q", g.Themes[0].Path, "acme/minimal/0.1.0")
	}
	if g.Themes[0].Contract.Minor != 0 {
		t.Fatalf("g.Themes[0].Contract.Minor = %d, want 0", g.Themes[0].Contract.Minor)
	}
	if g.Themes[0].Active {
		t.Fatalf("g.Themes[0].Active = true, want false")
	}
	if g.Themes[0].UncoveredPages != 0 {
		t.Fatalf("g.Themes[0].UncoveredPages = %d, want 0", g.Themes[0].UncoveredPages)
	}
	if g.Themes[0].ThemeAPI {
		t.Fatalf("g.Themes[0].ThemeAPI = true, want false")
	}
	if len(list.Pending) != 1 {
		t.Fatalf("len(list.Pending) = %d, want 1", len(list.Pending))
	}
	if list.Pending[0].Path != "acme/minimal/0.1.0" {
		t.Fatalf("list.Pending[0].Path = %q, want %q", list.Pending[0].Path, "acme/minimal/0.1.0")
	}
}

func TestThemeListActiveTheme(t *testing.T) {
	ctx := context.Background()
	svc, db, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)

	c, err := svc.Activate(ctx, "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}
	bindPage(t, db, c.Contract.ID, "post")

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	blogGroup := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	g := findGroup(t, list, blogGroup)
	if g.PageCount != 1 {
		t.Fatalf("g.PageCount = %d, want 1", g.PageCount)
	}
	if !g.HasBase {
		t.Fatalf("g.HasBase = false, want true")
	}
	if g.Base != 0 {
		t.Fatalf("g.Base = %d, want 0", g.Base)
	}
	if g.Active == nil {
		t.Fatalf("g.Active = nil, want non-nil")
	}
	if g.Active.Path != "acme/minimal/0.1.0" {
		t.Fatalf("g.Active.Path = %q, want %q", g.Active.Path, "acme/minimal/0.1.0")
	}
	if g.Active.ContractID != "example/blog@v1.0" {
		t.Fatalf("g.Active.ContractID = %q, want %q", g.Active.ContractID, "example/blog@v1.0")
	}
	if g.Active.Broken != nil {
		t.Fatalf("g.Active.Broken = %+v, want nil", g.Active.Broken)
	}
	if !g.Themes[0].Active {
		t.Fatalf("g.Themes[0].Active = false, want true")
	}
	if len(g.Records) != 1 {
		t.Fatalf("len(g.Records) = %d, want 1", len(g.Records))
	}
	if g.Records[0].Contract.ID != c.Contract.ID {
		t.Fatalf("g.Records[0].Contract.ID = %d, want %d", g.Records[0].Contract.ID, c.Contract.ID)
	}
	if g.Records[0].PageCount != 1 {
		t.Fatalf("g.Records[0].PageCount = %d, want 1", g.Records[0].PageCount)
	}
	if g.Records[0].CanDelete {
		t.Fatalf("g.Records[0].CanDelete = true, want false")
	}
}

func TestThemeListUncoveredTheme(t *testing.T) {
	svc, db, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)

	rec := models.Contract{
		Org:      "example",
		Name:     "blog",
		Major:    1,
		Minor:    1,
		Source:   "s",
		Checksum: "x",
	}
	if err := db.Create(&rec).Error; err != nil {
		t.Fatalf("create contract: %v", err)
	}
	bindPage(t, db, rec.ID, "post")
	layout := "post"
	page2 := models.SitePage{
		Name:       "P",
		Slug:       "p2",
		Enabled:    true,
		Content:    "",
		ContractID: &rec.ID,
		LayoutID:   &layout,
	}
	if err := db.Create(&page2).Error; err != nil {
		t.Fatalf("create page2: %v", err)
	}

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	blogGroup := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	g := findGroup(t, list, blogGroup)
	if !g.HasBase {
		t.Fatalf("g.HasBase = false, want true")
	}
	if g.Base != 1 {
		t.Fatalf("g.Base = %d, want 1", g.Base)
	}
	if g.PageCount != 2 {
		t.Fatalf("g.PageCount = %d, want 2", g.PageCount)
	}
	if g.Themes[0].UncoveredPages != 2 {
		t.Fatalf("g.Themes[0].UncoveredPages = %d, want 2", g.Themes[0].UncoveredPages)
	}
}

func TestThemeListThemeAPI(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Contract{}, &models.SitePage{}, &models.Option{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	themesDir := writeTestThemeWith(t, func(files map[string][]byte) {
		files["built.toml"] = bytes.Replace(files["built.toml"], []byte("theme-api-version = 1"), []byte(fmt.Sprintf("theme-api-version = %d", themecore.ThemeAPICurrent()+1)), 1)
	})
	svc := NewThemeService(themesDir, t.TempDir(), NewContractService(repository.NewContractRepository(db)), repository.NewThemeRepository(db), &fakeRegistry{})

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	blogGroup := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	g := findGroup(t, list, blogGroup)
	if !g.Themes[0].ThemeAPI {
		t.Fatalf("g.Themes[0].ThemeAPI = false, want true")
	}
}

func TestThemeListBrokenTheme(t *testing.T) {
	ctx := context.Background()
	db, themesDir, _ := activated(t)
	breakLicense(t, themesDir)
	svc, _ := restart(t, db, themesDir)
	if err := svc.LoadActive(ctx); err != nil {
		t.Fatalf("LoadActive: %v", err)
	}

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	blogGroup := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	g := findGroup(t, list, blogGroup)
	if g.Active == nil {
		t.Fatalf("g.Active = nil, want non-nil")
	}
	if g.Active.Broken == nil {
		t.Fatalf("g.Active.Broken = nil, want non-nil")
	}
	if g.Active.Broken.Err == nil {
		t.Fatalf("g.Active.Broken.Err = nil, want non-nil")
	}
}

func TestThemeListRecordWithoutTheme(t *testing.T) {
	svc, db, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)

	rec := models.Contract{
		Org:      "umono",
		Name:     "page",
		Major:    0,
		Minor:    1,
		Source:   "s",
		Checksum: "x",
	}
	if err := db.Create(&rec).Error; err != nil {
		t.Fatalf("create contract: %v", err)
	}

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Groups) != 2 {
		t.Fatalf("len(list.Groups) = %d, want 2", len(list.Groups))
	}
	blogGroup := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	if list.Groups[0].Group != blogGroup {
		t.Fatalf("list.Groups[0].Group = %q, want %q", list.Groups[0].Group, blogGroup)
	}
	pageGroup := models.Contract{Org: "umono", Name: "page", Major: 0, Minor: 1}.Group()
	p := findGroup(t, list, pageGroup)
	if len(p.Themes) != 0 {
		t.Fatalf("len(p.Themes) = %d, want 0", len(p.Themes))
	}
	if len(p.Records) != 1 {
		t.Fatalf("len(p.Records) = %d, want 1", len(p.Records))
	}
	if !p.Records[0].CanDelete {
		t.Fatalf("p.Records[0].CanDelete = false, want true")
	}
}

func TestThemeListUnreadableTheme(t *testing.T) {
	svc, _, themesDir, _ := newPendingTestService(t, &fakeRegistry{}, nil)

	dir := filepath.Join(themesDir, "acme", "broken", "1.0.0")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "built.toml"), []byte("not toml"), 0o644); err != nil {
		t.Fatalf("write built.toml: %v", err)
	}

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Unreadable) != 1 {
		t.Fatalf("len(list.Unreadable) = %d, want 1", len(list.Unreadable))
	}
	if list.Unreadable[0].Path != "acme/broken/1.0.0" {
		t.Fatalf("list.Unreadable[0].Path = %q, want %q", list.Unreadable[0].Path, "acme/broken/1.0.0")
	}
	if len(list.Groups) != 1 {
		t.Fatalf("len(list.Groups) = %d, want 1", len(list.Groups))
	}
}
