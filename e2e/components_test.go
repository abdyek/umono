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
