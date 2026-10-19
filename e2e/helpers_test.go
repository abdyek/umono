//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

// artifactsDir is where a failing test saves its trace and screenshot. It is
// empty unless E2E_ARTIFACTS is set.
var artifactsDir = os.Getenv("E2E_ARTIFACTS")

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
	if artifactsDir != "" {
		if err := ctx.Tracing().Start(playwright.TracingStartOptions{Screenshots: playwright.Bool(true), Snapshots: playwright.Bool(true)}); err != nil {
			t.Fatalf("start tracing: %v", err)
		}
	}
	page, err := ctx.NewPage()
	if err != nil {
		t.Fatalf("new page: %v", err)
	}
	p := &adminPage{Page: page, t: t}
	page.OnConsole(func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			p.record(fmt.Sprintf("console: %s (%s)", msg.Text(), msg.Location().URL))
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
		if artifactsDir != "" {
			p.saveArtifacts(ctx)
		}
		_ = ctx.Close()
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

// saveArtifacts saves the trace and a screenshot of a failing test to
// artifactsDir. The trace of a passing test is dropped.
func (p *adminPage) saveArtifacts(ctx playwright.BrowserContext) {
	if !p.t.Failed() {
		_ = ctx.Tracing().Stop()
		return
	}
	name := strings.ReplaceAll(p.t.Name(), "/", "_")
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		p.t.Logf("save artifacts: %v", err)
		return
	}
	if _, err := p.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(filepath.Join(artifactsDir, name+".png")), FullPage: playwright.Bool(true)}); err != nil {
		p.t.Logf("save screenshot: %v", err)
	}
	if err := ctx.Tracing().Stop(filepath.Join(artifactsDir, name+"-trace.zip")); err != nil {
		p.t.Logf("save trace: %v", err)
	}
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

// acceptDialogs accepts the confirm dialogs of hx-confirm.
func (p *adminPage) acceptDialogs() {
	p.OnDialog(func(d playwright.Dialog) { _ = d.Accept() })
}

// createPage creates a page in the page editor and returns the URL of its
// editor.
func (p *adminPage) createPage(name, slug, content string, enabled bool) string {
	p.t.Helper()
	if _, err := p.Goto("/admin/site-pages/new"); err != nil {
		p.t.Fatalf("goto /admin/site-pages/new: %v", err)
	}
	if err := p.Locator("#name").Fill(name); err != nil {
		p.t.Fatalf("fill name: %v", err)
	}
	if err := p.Locator("#slug").Fill(slug); err != nil {
		p.t.Fatalf("fill slug: %v", err)
	}
	if err := p.Locator("#content").Fill(content); err != nil {
		p.t.Fatalf("fill content: %v", err)
	}
	if enabled {
		if err := p.Locator("#enabled-checkbox").SetChecked(true, playwright.LocatorSetCheckedOptions{Force: playwright.Bool(true)}); err != nil {
			p.t.Fatalf("set enabled: %v", err)
		}
	}
	if err := p.Locator("#editor-submit-button").Click(); err != nil {
		p.t.Fatalf("click submit: %v", err)
	}
	if err := p.WaitForURL(regexp.MustCompile(`/admin/site-pages/\d+$`)); err != nil {
		p.t.Fatalf("wait for new page url: %v", err)
	}
	return p.URL()
}
