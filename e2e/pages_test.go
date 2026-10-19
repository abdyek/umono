//go:build e2e

package e2e

import (
	"regexp"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

func TestCreatePageWithNewPageButton(t *testing.T) {
	p := newPage(t)
	p.login()
	if err := p.Locator(`aside a[href="/admin/site-pages/new"]`).Click(); err != nil {
		t.Fatalf("click new page button: %v", err)
	}
	if err := expect.Locator(p.Locator("#editor-area h2")).ToHaveText("Create Page"); err != nil {
		t.Fatalf("assert create page heading: %v", err)
	}
	if err := p.Locator("#name").Fill("E2E Button Page"); err != nil {
		t.Fatalf("fill name: %v", err)
	}
	if err := p.Locator("#slug").Fill("e2e-button-page"); err != nil {
		t.Fatalf("fill slug: %v", err)
	}
	if err := p.Locator("#content").Fill("Made with the new page button."); err != nil {
		t.Fatalf("fill content: %v", err)
	}
	if err := p.Locator("#editor-submit-button").Click(); err != nil {
		t.Fatalf("click submit: %v", err)
	}
	if err := p.WaitForURL(regexp.MustCompile(`/admin/site-pages/\d+$`)); err != nil {
		t.Fatalf("wait for new page url: %v", err)
	}
	if err := expect.Locator(p.Locator("aside")).ToContainText("E2E Button Page"); err != nil {
		t.Fatalf("assert sidebar contains new page: %v", err)
	}
}

func TestSaveAndEnablePage(t *testing.T) {
	p := newPage(t)
	p.login()
	url := p.createPage("E2E Save Page", "e2e-save-page", "First content.", false)
	if err := p.Locator("#content").Fill("Saved content."); err != nil {
		t.Fatalf("fill content: %v", err)
	}
	if err := p.Locator("#enabled-checkbox").SetChecked(true, playwright.LocatorSetCheckedOptions{Force: playwright.Bool(true)}); err != nil {
		t.Fatalf("set enabled: %v", err)
	}
	if _, err := p.ExpectResponse(regexp.MustCompile(`/admin/site-pages/\d+$`), func() error {
		return p.Locator("#editor-submit-button").Click()
	}); err != nil {
		t.Fatalf("save page: %v", err)
	}
	if _, err := p.Goto(url); err != nil {
		t.Fatalf("goto editor url: %v", err)
	}
	if err := expect.Locator(p.Locator("#content")).ToHaveValue("Saved content."); err != nil {
		t.Fatalf("assert content saved: %v", err)
	}
	if err := expect.Locator(p.Locator("#enabled-checkbox")).ToBeChecked(); err != nil {
		t.Fatalf("assert enabled checked: %v", err)
	}
	if _, err := p.Goto("/e2e-save-page"); err != nil {
		t.Fatalf("goto public page: %v", err)
	}
	if err := expect.Locator(p.Locator("body")).ToContainText("Saved content."); err != nil {
		t.Fatalf("assert public page content: %v", err)
	}
}

func TestDeletePage(t *testing.T) {
	p := newPage(t)
	p.login()
	p.acceptDialogs()
	p.createPage("E2E Delete Page", "e2e-delete-page", "Delete me.", true)
	if err := p.Locator("#editor-delete-button").Click(); err != nil {
		t.Fatalf("click delete: %v", err)
	}
	if err := p.WaitForURL(baseURL + "/admin/site-pages/new"); err != nil {
		t.Fatalf("wait for redirect to new: %v", err)
	}
	if err := expect.Locator(p.Locator("aside")).Not().ToContainText("E2E Delete Page"); err != nil {
		t.Fatalf("assert sidebar omits deleted page: %v", err)
	}
	res, err := p.Request().Get(baseURL + "/e2e-delete-page")
	if err != nil {
		t.Fatalf("get deleted page: %v", err)
	}
	if res.Status() != 404 {
		t.Errorf("deleted page status = %d, want 404", res.Status())
	}
}
