//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

// adminPage is a page in its own browser context. It records the errors the
// page reports and fails the test at cleanup for each error that was not
// allowed with allowError.
type adminPage struct {
	playwright.Page
	t       *testing.T
	mu      sync.Mutex
	errors  []string
	allowed []string
}

// htmxErrorScript reports the htmx error events as console errors.
const htmxErrorScript = `["htmx:targetError", "htmx:responseError"].forEach(function (name) {
  document.addEventListener(name, function () { console.error(name); });
})`

func newPage(t *testing.T) *adminPage {
	t.Helper()
	ctx, err := browser.NewContext(playwright.BrowserNewContextOptions{BaseURL: playwright.String(baseURL)})
	if err != nil {
		t.Fatalf("new context: %v", err)
	}
	if err := ctx.AddInitScript(playwright.Script{Content: playwright.String(htmxErrorScript)}); err != nil {
		t.Fatalf("add init script: %v", err)
	}
	page, err := ctx.NewPage()
	if err != nil {
		t.Fatalf("new page: %v", err)
	}
	p := &adminPage{Page: page, t: t}
	page.OnConsole(func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			p.record("console: " + msg.Text())
		}
	})
	page.OnPageError(func(err error) {
		p.record("pageerror: " + err.Error())
	})
	page.OnResponse(func(res playwright.Response) {
		if res.Status() >= 500 {
			p.record(fmt.Sprintf("response: %d %s", res.Status(), res.URL()))
		}
	})
	t.Cleanup(func() {
		_ = ctx.Close()
		p.mu.Lock()
		errs := append([]string(nil), p.errors...)
		allowed := append([]string(nil), p.allowed...)
		p.mu.Unlock()
		for _, e := range errs {
			ok := false
			for _, a := range allowed {
				if strings.Contains(e, a) {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf("page error: %s", e)
			}
		}
	})
	return p
}

func (p *adminPage) record(e string) {
	p.mu.Lock()
	p.errors = append(p.errors, e)
	p.mu.Unlock()
}

// allowError lets the test expect the errors that contain substr.
func (p *adminPage) allowError(substr string) {
	p.mu.Lock()
	p.allowed = append(p.allowed, substr)
	p.mu.Unlock()
}

// login signs in from /admin/login and waits for the admin panel.
func (p *adminPage) login() {
	p.t.Helper()
	if _, err := p.Goto("/admin/login"); err != nil {
		p.t.Fatalf("goto /admin/login: %v", err)
	}
	if err := p.Locator("#username").Fill(testUsername); err != nil {
		p.t.Fatalf("fill username: %v", err)
	}
	if err := p.Locator("#password").Fill(testPassword); err != nil {
		p.t.Fatalf("fill password: %v", err)
	}
	if err := p.Locator("#submitBtn").Click(); err != nil {
		p.t.Fatalf("click submit: %v", err)
	}
	if err := p.Locator("#logout-button").WaitFor(); err != nil {
		p.t.Fatalf("wait for the admin panel: %v", err)
	}
}
