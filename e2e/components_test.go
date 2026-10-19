//go:build e2e

package e2e

import (
	"regexp"
	"testing"
)

func TestCreateGlobalAndUseItOnPage(t *testing.T) {
	p := newPage(t)
	p.login()
	if _, err := p.Goto("/admin/components/new"); err != nil {
		t.Fatalf("goto /admin/components/new: %v", err)
	}
	if err := p.Locator("#name").Fill("E2E_GREETING"); err != nil {
		t.Fatalf("fill name: %v", err)
	}
	if err := p.Locator("#content").Fill("Hello from the e2e global."); err != nil {
		t.Fatalf("fill content: %v", err)
	}
	if err := p.Locator("#editor-submit-button").Click(); err != nil {
		t.Fatalf("click submit: %v", err)
	}
	if err := p.WaitForURL(regexp.MustCompile(`/admin/components/\d+$`)); err != nil {
		t.Fatalf("wait for new component url: %v", err)
	}
	p.createPage("E2E Global Page", "e2e-global-page", "{{ E2E_GREETING }}", true)
	if _, err := p.Goto("/e2e-global-page"); err != nil {
		t.Fatalf("goto public page: %v", err)
	}
	if err := expect.Locator(p.Locator("body")).ToContainText("Hello from the e2e global."); err != nil {
		t.Fatalf("assert public page contains greeting: %v", err)
	}
	// A global saved in the admin UI used to break the render of every page
	// after the next requests.
	p.createPage("E2E After Global", "e2e-after-global", "After the global.", true)
	if _, err := p.Goto("/e2e-after-global"); err != nil {
		t.Fatalf("goto /e2e-after-global: %v", err)
	}
	if err := expect.Locator(p.Locator("body")).ToContainText("After the global."); err != nil {
		t.Fatalf("assert /e2e-after-global contains body text: %v", err)
	}
	if _, err := p.Goto("/e2e-global-page"); err != nil {
		t.Fatalf("goto /e2e-global-page: %v", err)
	}
	if err := expect.Locator(p.Locator("body")).ToContainText("Hello from the e2e global."); err != nil {
		t.Fatalf("assert /e2e-global-page still contains greeting: %v", err)
	}
}

func TestNewComponentButton(t *testing.T) {
	p := newPage(t)
	p.login()
	if err := p.Locator(`aside a[href="/admin/components/new"]`).Click(); err != nil {
		t.Fatalf("click new component button: %v", err)
	}
	if err := expect.Locator(p.Locator("#editor-area h2")).ToHaveText("Create Component"); err != nil {
		t.Fatalf("assert create component heading: %v", err)
	}
	if err := expect.Locator(p.Locator("#preview-content")).ToBeVisible(); err != nil {
		t.Fatalf("assert preview visible: %v", err)
	}
}

func TestEditAndDeleteGlobal(t *testing.T) {
	p := newPage(t)
	p.login()
	p.acceptDialogs()
	if _, err := p.Goto("/admin/components/new"); err != nil {
		t.Fatalf("goto /admin/components/new: %v", err)
	}
	if err := p.Locator("#name").Fill("E2E_EDITED"); err != nil {
		t.Fatalf("fill name: %v", err)
	}
	if err := p.Locator("#content").Fill("Before edit."); err != nil {
		t.Fatalf("fill content: %v", err)
	}
	if err := p.Locator("#editor-submit-button").Click(); err != nil {
		t.Fatalf("click submit: %v", err)
	}
	if err := p.WaitForURL(regexp.MustCompile(`/admin/components/\d+$`)); err != nil {
		t.Fatalf("wait for new component url: %v", err)
	}
	url := p.URL()
	if _, err := p.Goto("/admin/site-pages/new"); err != nil {
		t.Fatalf("goto /admin/site-pages/new: %v", err)
	}
	if err := p.Locator(`aside a:has-text("E2E_EDITED")`).Click(); err != nil {
		t.Fatalf("click sidebar entry for E2E_EDITED: %v", err)
	}
	if err := expect.Locator(p.Locator("#editor-area h2")).ToHaveText("E2E_EDITED"); err != nil {
		t.Fatalf("assert editor heading is E2E_EDITED: %v", err)
	}
	if err := expect.Locator(p.Locator("#preview-content")).ToBeVisible(); err != nil {
		t.Fatalf("assert preview visible: %v", err)
	}
	// htmx keeps the value it sees when it sets the keyup trigger up. A full
	// page load sets it up before the test types, so the key press below is
	// a change.
	if _, err := p.Goto(url); err != nil {
		t.Fatalf("goto editor url: %v", err)
	}
	if err := p.Locator("#content").Fill("After edit"); err != nil {
		t.Fatalf("fill content: %v", err)
	}
	// The live preview has hx-indicator="false", which htmx logs as a
	// selector without matches.
	p.allowError(`The selector "false" on hx-indicator returned no matches!`)
	if _, err := p.ExpectResponse("**/admin/components/preview", func() error {
		return p.Locator("#content").Press(".")
	}); err != nil {
		t.Fatalf("press . to trigger preview: %v", err)
	}
	if _, err := p.ExpectResponse(regexp.MustCompile(`/admin/components/\d+$`), func() error {
		return p.Locator("#editor-submit-button").Click()
	}); err != nil {
		t.Fatalf("submit edited component: %v", err)
	}
	if _, err := p.Goto(url); err != nil {
		t.Fatalf("goto editor url: %v", err)
	}
	if err := expect.Locator(p.Locator("#content")).ToHaveValue("After edit."); err != nil {
		t.Fatalf("assert content saved: %v", err)
	}
	if err := expect.Locator(p.Locator("#preview-content")).ToBeVisible(); err != nil {
		t.Fatalf("assert preview visible after edit: %v", err)
	}
	if err := p.Locator("#editor-delete-button").Click(); err != nil {
		t.Fatalf("click delete: %v", err)
	}
	if err := p.WaitForURL(baseURL + "/admin/components/new"); err != nil {
		t.Fatalf("wait for redirect to new: %v", err)
	}
	if err := expect.Locator(p.Locator("aside")).Not().ToContainText("E2E_EDITED"); err != nil {
		t.Fatalf("assert sidebar omits deleted component: %v", err)
	}
}
