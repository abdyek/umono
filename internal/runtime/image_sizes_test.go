package runtime

import (
	"strings"
	"testing"
)

func TestOptimizeHTMLAddsFullViewportSizesOutsideGrid(t *testing.T) {
	output, err := OptimizeHTML(`<compono-image><picture><source type="image/webp" srcset="/hero-640.webp 640w"><img src="/hero.jpg" alt="Hero" width="1280" height="720"></picture></compono-image>`)
	if err != nil {
		t.Fatalf("OptimizeHTML() error = %v", err)
	}

	if !strings.Contains(output, `sizes="100vw"`) {
		t.Fatalf("output = %q", output)
	}
}

func TestOptimizeHTMLAddsGridFractionSizes(t *testing.T) {
	input := `<compono-web-grid data-grid-template-columns="1fr 1fr" data-grid-template-rows="min-content" data-grid-template-areas='[["media","content"]]'><compono-web-grid-item data-grid-area="media"><compono-image><picture><source type="image/webp" srcset="/hero-640.webp 640w"><img src="/hero.jpg" alt="Hero" width="1280" height="720"></picture></compono-image></compono-web-grid-item><compono-web-grid-item data-grid-area="content"><p>Copy</p></compono-web-grid-item></compono-web-grid>`

	output, err := OptimizeHTML(input)
	if err != nil {
		t.Fatalf("OptimizeHTML() error = %v", err)
	}

	if !strings.Contains(output, `sizes="50vw"`) {
		t.Fatalf("output = %q", output)
	}
}

func TestOptimizeHTMLAddsResponsiveGridSizes(t *testing.T) {
	input := `<compono-web-grid data-grid-template-columns="1fr" data-grid-template-rows="min-content" data-grid-template-areas='[["media"]]' data-md-grid-template-columns="1fr 1fr" data-md-grid-template-rows="min-content" data-md-grid-template-areas='[["media","content"]]' data-xl-grid-template-columns="1fr 1fr 1fr" data-xl-grid-template-rows="min-content" data-xl-grid-template-areas='[["media","content","aside"]]'><compono-web-grid-item data-grid-area="media"><compono-image><picture><source type="image/webp" srcset="/hero-640.webp 640w"><img src="/hero.jpg" alt="Hero" width="1280" height="720"></picture></compono-image></compono-web-grid-item><compono-web-grid-item data-grid-area="content"><p>Copy</p></compono-web-grid-item><compono-web-grid-item data-grid-area="aside"><p>Aside</p></compono-web-grid-item></compono-web-grid>`

	output, err := OptimizeHTML(input)
	if err != nil {
		t.Fatalf("OptimizeHTML() error = %v", err)
	}

	want := `sizes="(min-width: 1200px) 33.33vw, (min-width: 768px) 50vw, 100vw"`
	if !strings.Contains(output, want) {
		t.Fatalf("output = %q", output)
	}
}

func TestGenerateGridCSSIncludesImageWidthRules(t *testing.T) {
	css, err := GenerateGridCSS(`<img src="/hero.jpg" alt="Hero" width="1280" height="720">`)
	if err != nil {
		t.Fatalf("GenerateGridCSS() error = %v", err)
	}

	for _, want := range []string{
		`compono-image,picture,img{box-sizing:border-box;max-width:100%;width:100%;}`,
		`compono-image,picture{display:block;}`,
		`compono-image>picture,compono-image>img,picture>img{display:block;height:auto;}`,
	} {
		if !strings.Contains(css, want) {
			t.Fatalf("css = %q", css)
		}
	}
}
