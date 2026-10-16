package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

const fullPageCall = "{{ UMONO_LAYOUT site-name = \"S\" home-url = \"/\" menu-label = \"M\" menu = [] content = BODY footer = FOOT }}\n\n~ BODY\nHi\n\n~ FOOT\nBye\n"

func pageContractSource(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "themeapi", "testdata", "dom", "scenarios", "contracts", "umono", "page", "v0.1", "page-block.txtar"))
	if err != nil {
		t.Fatalf("read page contract fixture: %v", err)
	}
	src := string(b)
	startMarker := "-- contract.toml --\n"
	startIdx := strings.Index(src, startMarker)
	if startIdx < 0 {
		t.Fatalf("start marker %q not found in page-block.txtar", startMarker)
	}
	start := startIdx + len(startMarker)
	endMarker := "-- page.comp --"
	endIdx := strings.Index(src, endMarker)
	if endIdx < 0 {
		t.Fatalf("end marker %q not found in page-block.txtar", endMarker)
	}
	return src[start:endIdx]
}

func storeContract(t *testing.T, db *gorm.DB, org, name string, major, minor uint, source string) models.Contract {
	t.Helper()
	rec := models.Contract{
		Org:      org,
		Name:     name,
		Major:    major,
		Minor:    minor,
		Source:   source,
		Checksum: "x",
	}
	if err := db.Create(&rec).Error; err != nil {
		t.Fatalf("create contract: %v", err)
	}
	return rec
}

func bindingErr(t *testing.T, err error) *BindingError {
	t.Helper()
	if err == nil {
		t.Fatalf("expected *BindingError, got nil")
	}
	var be *BindingError
	if !errors.As(err, &be) {
		t.Fatalf("expected *BindingError, got %v", err)
	}
	return be
}

func sameNames(a, b []string) bool {
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

func strp(s string) *string {
	return &s
}

func TestCheckBindingAcceptsAValidPage(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})

	rec := storeContract(t, db, "umono", "page", 0, 1, pageContractSource(t))
	page := models.SitePage{ContractID: &rec.ID, LayoutID: strp("page"), Content: fullPageCall}

	if err := svc.CheckBinding(page, nil); err != nil {
		t.Fatalf("CheckBinding error = %v, want nil", err)
	}
}

func TestCheckBindingUnboundPage(t *testing.T) {
	svc, _ := newThemeTestService(t, &fakeRegistry{})

	if err := svc.CheckBinding(models.SitePage{Content: "# Hi"}, nil); err != nil {
		t.Fatalf("CheckBinding (# Hi) error = %v, want nil", err)
	}

	e := bindingErr(t, svc.CheckBinding(models.SitePage{Content: "{{ UMONO_LAYOUT }}\n"}, nil))
	if e.Kind != BindingUnboundCall {
		t.Fatalf("e.Kind = %q, want %q", e.Kind, BindingUnboundCall)
	}
}

func TestCheckBindingRootRule(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})
	rec := storeContract(t, db, "umono", "page", 0, 1, pageContractSource(t))

	emptySkeleton := "{{ UMONO_LAYOUT site-name = \"\" home-url = \"\" menu-label = \"\" menu = [] content = CONTENT footer = FOOTER }}"

	cases := []struct {
		name    string
		content string
	}{
		{name: "empty", content: ""},
		{name: "markdown prefix", content: "# Title\n\n" + fullPageCall},
		{name: "two layouts", content: fullPageCall + "\n" + "{{ UMONO_LAYOUT site-name = \"S\" home-url = \"/\" menu-label = \"M\" menu = [] content = BODY footer = FOOT }}\n"},
		{name: "text prefix", content: "Hello {{ UMONO_LAYOUT site-name = \"S\" home-url = \"/\" menu-label = \"M\" menu = [] content = BODY footer = FOOT }}\n\n~ BODY\nHi\n\n~ FOOT\nBye\n"},
		{name: "only markdown", content: "# Only markdown\n"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			page := models.SitePage{ContractID: &rec.ID, LayoutID: strp("page"), Content: c.content}
			e := bindingErr(t, svc.CheckBinding(page, nil))
			if e.Kind != BindingRoot {
				t.Fatalf("e.Kind = %q, want %q", e.Kind, BindingRoot)
			}
			if c.name == "empty" {
				if e.Skeleton != emptySkeleton {
					t.Fatalf("e.Skeleton = %q, want %q", e.Skeleton, emptySkeleton)
				}
			}
		})
	}
}

func TestCheckBindingLocalCall(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})
	rec := storeContract(t, db, "umono", "page", 0, 1, pageContractSource(t))

	content := fullPageCall + "\n~ X\n{{ UMONO_LAYOUT }}\n"
	page := models.SitePage{ContractID: &rec.ID, LayoutID: strp("page"), Content: content}

	e := bindingErr(t, svc.CheckBinding(page, nil))
	if e.Kind != BindingLocalCall {
		t.Fatalf("e.Kind = %q, want %q", e.Kind, BindingLocalCall)
	}
}

func TestCheckBindingUnknownParameter(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})
	rec := storeContract(t, db, "umono", "page", 0, 1, pageContractSource(t))

	content := "{{ UMONO_LAYOUT site-name = \"S\" home-url = \"/\" menu-label = \"M\" menu = [] content = BODY footer = FOOT extra = \"x\" }}\n\n~ BODY\nHi\n\n~ FOOT\nBye\n"
	page := models.SitePage{ContractID: &rec.ID, LayoutID: strp("page"), Content: content}

	e := bindingErr(t, svc.CheckBinding(page, nil))
	if e.Kind != BindingUnknownParameter {
		t.Fatalf("e.Kind = %q, want %q", e.Kind, BindingUnknownParameter)
	}
	if !sameNames(e.Names, []string{"extra"}) {
		t.Fatalf("e.Names = %v, want %v", e.Names, []string{"extra"})
	}
	if len(e.Offered) != 0 {
		t.Fatalf("len(e.Offered) = %d, want 0; e.Offered = %v", len(e.Offered), e.Offered)
	}
}

func TestCheckBindingMissingParameter(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})
	rec := storeContract(t, db, "umono", "page", 0, 1, pageContractSource(t))

	content := "{{ UMONO_LAYOUT site-name = \"S\" home-url = \"/\" menu = [] content = BODY }}\n\n~ BODY\nHi\n"
	page := models.SitePage{ContractID: &rec.ID, LayoutID: strp("page"), Content: content}

	e := bindingErr(t, svc.CheckBinding(page, nil))
	if e.Kind != BindingMissingParameter {
		t.Fatalf("e.Kind = %q, want %q", e.Kind, BindingMissingParameter)
	}
	if !sameNames(e.Names, []string{"menu-label", "footer"}) {
		t.Fatalf("e.Names = %v, want %v", e.Names, []string{"menu-label", "footer"})
	}
}

func TestCheckBindingLayoutNotFound(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})
	rec := storeContract(t, db, "umono", "page", 0, 1, pageContractSource(t))

	page := models.SitePage{ContractID: &rec.ID, LayoutID: strp("nope"), Content: fullPageCall}

	e := bindingErr(t, svc.CheckBinding(page, nil))
	if e.Kind != BindingLayoutNotFound {
		t.Fatalf("e.Kind = %q, want %q", e.Kind, BindingLayoutNotFound)
	}
}

func TestCheckBindingNotOffered(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})
	old := storeBlogContract(t, db)
	newer := storeContract(t, db, "example", "blog", 1, 1, string(contractSource(t)))

	page := models.SitePage{ContractID: &old.ID, LayoutID: strp("post"), Content: "{{ UMONO_LAYOUT title = \"T\" }}\n"}

	e := bindingErr(t, svc.CheckBinding(page, nil))
	if e.Kind != BindingNotOffered {
		t.Fatalf("e.Kind = %q, want %q", e.Kind, BindingNotOffered)
	}

	previous := models.SitePage{ContractID: &old.ID, LayoutID: strp("post")}
	if err := svc.CheckBinding(page, &previous); err != nil {
		t.Fatalf("CheckBinding (with previous) error = %v, want nil", err)
	}

	page.ContractID = &newer.ID
	if err := svc.CheckBinding(page, nil); err != nil {
		t.Fatalf("CheckBinding (newer) error = %v, want nil", err)
	}
}

func TestCheckBindingParameterOfTheOfferedRecord(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})
	old := storeBlogContract(t, db)
	newSource := strings.Replace(string(contractSource(t)), "~ title = \"\" body = NO_MATTER", "~ title = \"\" subtitle = \"\" body = NO_MATTER", 1)
	storeContract(t, db, "example", "blog", 1, 1, newSource)

	page := models.SitePage{ContractID: &old.ID, LayoutID: strp("post"), Content: "{{ UMONO_LAYOUT title = \"T\" subtitle = \"S\" }}\n"}
	previous := models.SitePage{ContractID: &old.ID, LayoutID: strp("post")}

	e := bindingErr(t, svc.CheckBinding(page, &previous))
	if e.Kind != BindingUnknownParameter {
		t.Fatalf("e.Kind = %q, want %q", e.Kind, BindingUnknownParameter)
	}
	if !sameNames(e.Names, []string{"subtitle"}) {
		t.Fatalf("e.Names = %v, want %v", e.Names, []string{"subtitle"})
	}
	if !sameNames(e.Offered, []string{"subtitle"}) {
		t.Fatalf("e.Offered = %v, want %v", e.Offered, []string{"subtitle"})
	}
}
