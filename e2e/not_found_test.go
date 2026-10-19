//go:build e2e

package e2e

import (
	"strings"
	"testing"
)

func TestSaveNotFoundPage(t *testing.T) {
	p := newPage(t)
	p.login()
	p.acceptDialogs()
	if _, err := p.Goto("/admin/settings/404-page"); err != nil {
		t.Fatalf("goto /admin/settings/404-page: %v", err)
	}
	if err := p.Locator("#title").Fill("E2E Not Found"); err != nil {
		t.Fatalf("fill title: %v", err)
	}
	if err := p.Locator("#content").Fill("E2E missing page."); err != nil {
		t.Fatalf("fill content: %v", err)
	}
	if _, err := p.ExpectResponse("**/admin/options/not-found-page-option", func() error {
		return p.Locator("#editor-submit-button").Click()
	}); err != nil {
		t.Fatalf("save not found page: %v", err)
	}
	if _, err := p.Goto("/admin/settings/404-page"); err != nil {
		t.Fatalf("goto /admin/settings/404-page: %v", err)
	}
	if err := expect.Locator(p.Locator("#content")).ToHaveValue("E2E missing page."); err != nil {
		t.Fatalf("assert content saved: %v", err)
	}
	res, err := p.Request().Get(baseURL + "/e2e-no-such-page")
	if err != nil {
		t.Fatalf("get not found page: %v", err)
	}
	if res.Status() != 404 {
		t.Errorf("status = %d, want 404", res.Status())
	}
	body, err := res.Text()
	if err != nil {
		t.Fatalf("read not found page body: %v", err)
	}
	if !strings.Contains(body, "E2E missing page.") {
		t.Errorf("404 page does not show the saved content: %s", body)
	}
}
