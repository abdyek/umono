//go:build e2e

package e2e

import (
	"regexp"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

// devLayout is the label of the layout of the theme in development in the
// layout field.
const devLayout = "example/blog v1.0 · post"

// layoutCall is the content of a page that renders the layout of the theme in
// development. body keeps its default: it takes a component, not a string.
const layoutCall = `{{ UMONO_LAYOUT title = "Layout title" }}`

func TestThemesPageShowsTheThemeInDevelopment(t *testing.T) {
	p := newDevPage(t)
	p.login()
	if _, err := p.Goto("/admin/settings/themes"); err != nil {
		t.Fatalf("goto /admin/settings/themes: %v", err)
	}
	if err := expect.Locator(p.Locator("#dev-theme")).ToContainText("Theme in development"); err != nil {
		t.Fatalf("assert dev theme heading: %v", err)
	}
	if err := expect.Locator(p.Locator("#dev-theme")).ToContainText("Minimal"); err != nil {
		t.Fatalf("assert dev theme name: %v", err)
	}
	if err := expect.Locator(p.Locator("#dev-theme")).ToContainText("example/blog@v1.0"); err != nil {
		t.Fatalf("assert dev theme id: %v", err)
	}
	if err := expect.Locator(p.Locator("body")).Not().ToContainText("Themes that cannot be read"); err != nil {
		t.Fatalf("assert no unread themes: %v", err)
	}
	if err := expect.Locator(p.Locator(`form[hx-post="/admin/settings/themes/deactivate"]`)).ToHaveCount(0); err != nil {
		t.Fatalf("assert no deactivate form: %v", err)
	}
}

func TestLayoutChoiceUpdatesThePreview(t *testing.T) {
	p := newDevPage(t)
	p.login()
	if _, err := p.Goto("/admin/site-pages/new"); err != nil {
		t.Fatalf("goto /admin/site-pages/new: %v", err)
	}
	if err := p.Locator("#content").Fill(layoutCall); err != nil {
		t.Fatalf("fill content: %v", err)
	}
	if _, err := p.ExpectResponse("**/admin/site-pages/preview", func() error {
		_, err := p.Locator("#layout").SelectOption(playwright.SelectOptionValues{Labels: playwright.StringSlice(devLayout)})
		return err
	}); err != nil {
		t.Fatalf("select layout: %v", err)
	}
	if err := expect.Locator(p.Locator("#preview-content .preview-body")).ToContainText("Layout title"); err != nil {
		t.Fatalf("assert preview body: %v", err)
	}
}

func TestBoundPageRendersWithTheThemeInDevelopment(t *testing.T) {
	p := newDevPage(t)
	p.login()
	if _, err := p.Goto("/admin/site-pages/new"); err != nil {
		t.Fatalf("goto /admin/site-pages/new: %v", err)
	}
	if err := p.Locator("#name").Fill("E2E Dev Post"); err != nil {
		t.Fatalf("fill name: %v", err)
	}
	if err := p.Locator("#slug").Fill("e2e-dev-post"); err != nil {
		t.Fatalf("fill slug: %v", err)
	}
	if _, err := p.Locator("#layout").SelectOption(playwright.SelectOptionValues{Labels: playwright.StringSlice(devLayout)}); err != nil {
		t.Fatalf("select layout: %v", err)
	}
	if err := p.Locator("#content").Fill(layoutCall); err != nil {
		t.Fatalf("fill content: %v", err)
	}
	if err := p.Locator("#enabled-checkbox").SetChecked(true, playwright.LocatorSetCheckedOptions{Force: playwright.Bool(true)}); err != nil {
		t.Fatalf("set enabled: %v", err)
	}
	if err := p.Locator("#editor-submit-button").Click(); err != nil {
		t.Fatalf("click submit: %v", err)
	}
	if err := p.WaitForURL(regexp.MustCompile(`/admin/site-pages/\d+$`)); err != nil {
		t.Fatalf("wait for new page url: %v", err)
	}

	res, err := p.Request().Get(devBaseURL + "/e2e-dev-post")
	if err != nil {
		t.Fatalf("get page: %v", err)
	}
	if res.Status() != 200 {
		t.Fatalf("page status = %d, want 200", res.Status())
	}
	body, err := res.Text()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(body, `<umono-root class="layout-post">`) {
		t.Fatalf("body missing layout root: %s", body)
	}
	if !strings.Contains(body, "/_theme/dev/theme.css") {
		t.Fatalf("body missing dev theme css: %s", body)
	}

	css, err := p.Request().Get(devBaseURL + "/_theme/dev/theme.css")
	if err != nil {
		t.Fatalf("get dev theme css: %v", err)
	}
	if css.Status() != 200 {
		t.Fatalf("dev theme css status = %d, want 200", css.Status())
	}
	if css.Headers()["cache-control"] != "no-store" {
		t.Fatalf("dev theme css cache-control = %q, want %q", css.Headers()["cache-control"], "no-store")
	}
	cssBody, err := css.Text()
	if err != nil {
		t.Fatalf("read dev theme css: %v", err)
	}
	if !strings.Contains(cssBody, "rebeccapurple") {
		t.Fatalf("dev theme css missing rebeccapurple: %s", cssBody)
	}
}
