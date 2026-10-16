package service

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/umono-cms/themecore"
)

func fixturePageRender(t *testing.T) PageRender {
	t.Helper()
	contract, err := os.ReadFile("../theme/testdata/package/contract.toml")
	if err != nil {
		t.Fatalf("os.ReadFile() contract error = %v", err)
	}
	attributes, err := os.ReadFile("../theme/testdata/package/attributes.toml")
	if err != nil {
		t.Fatalf("os.ReadFile() attributes error = %v", err)
	}
	c, err := themecore.ParseContract(contract)
	if err != nil {
		t.Fatalf("themecore.ParseContract() error = %v", err)
	}
	table, err := themecore.CompileAttributes(attributes, c)
	if err != nil {
		t.Fatalf("themecore.CompileAttributes() error = %v", err)
	}
	return PageRender{
		LayoutID:   "post",
		Layout:     c.Layouts[0].Source,
		Components: c.Components,
		Release:    "abc123abc123",
		Attributes: &table,
	}
}

func TestRenderRootWithoutLayout(t *testing.T) {
	ctx := context.Background()
	cc, err := NewContentCompiler(nil)
	if err != nil {
		t.Fatalf("NewContentCompiler: %v", err)
	}

	root, _, diagnostics, err := cc.RenderRoot(ctx, "# Hi", PageRender{})
	if err != nil {
		t.Fatalf("RenderRoot: %v", err)
	}
	if got, want := len(diagnostics), 0; got != want {
		t.Fatalf("len(diagnostics) = %d, want %d", got, want)
	}
	if got, want := strings.HasPrefix(root, "<umono-root>"), true; got != want {
		t.Fatalf("strings.HasPrefix(root, %q) = %v, want %v", "<umono-root>", got, want)
	}
	if got, want := strings.HasSuffix(root, "</umono-root>"), true; got != want {
		t.Fatalf("strings.HasSuffix(root, %q) = %v, want %v", "</umono-root>", got, want)
	}
	if got, want := strings.Contains(root, "Hi"), true; got != want {
		t.Fatalf("strings.Contains(root, %q) = %v, want %v", "Hi", got, want)
	}
}

func TestRenderRootWithLayout(t *testing.T) {
	ctx := context.Background()
	cc, err := NewContentCompiler(nil)
	if err != nil {
		t.Fatalf("NewContentCompiler: %v", err)
	}

	root, _, diagnostics, err := cc.RenderRoot(ctx, boundContent, fixturePageRender(t))
	if err != nil {
		t.Fatalf("RenderRoot: %v", err)
	}
	if got, want := len(diagnostics), 0; got != want {
		t.Fatalf("len(diagnostics) = %d, want %d", got, want)
	}
	if got, want := strings.HasPrefix(root, `<umono-root class="layout-post">`), true; got != want {
		t.Fatalf("strings.HasPrefix(root, %q) = %v, want %v; root = %q", `<umono-root class="layout-post">`, got, want, root)
	}
	if got, want := strings.Contains(root, "Hello"), true; got != want {
		t.Fatalf("strings.Contains(root, %q) = %v, want %v", "Hello", got, want)
	}
}

func TestRenderRootReturnsDiagnostics(t *testing.T) {
	ctx := context.Background()
	cc, err := NewContentCompiler(nil)
	if err != nil {
		t.Fatalf("NewContentCompiler: %v", err)
	}

	root, _, diagnostics, err := cc.RenderRoot(ctx, "{{ NOPE }}", PageRender{})
	if err != nil {
		t.Fatalf("RenderRoot: %v", err)
	}
	if got, want := len(diagnostics), 1; got != want {
		t.Fatalf("len(diagnostics) = %d, want %d", got, want)
	}
	if got, want := root, "<umono-root></umono-root>"; got != want {
		t.Fatalf("root = %q, want %q", got, want)
	}
}
