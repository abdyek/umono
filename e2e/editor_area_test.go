//go:build e2e

package e2e

import (
	"testing"
)

// swapsOneEditorArea runs action, which swaps the editor, and fails the
// test unless the page has exactly one editor area afterwards. It marks the
// current editor area first and waits for an unmarked one, so it does not
// look before the swap.
func (p *adminPage) swapsOneEditorArea(name string, action func() error) {
	p.t.Helper()
	if _, err := p.Locator(`[id="editor-area"]`).First().Evaluate(`el => { el.dataset.stale = "1" }`, nil); err != nil {
		p.t.Fatalf("%s: mark editor area: %v", name, err)
	}
	if err := action(); err != nil {
		p.t.Fatalf("%s: %v", name, err)
	}
	if err := p.Locator(`[id="editor-area"]:not([data-stale])`).First().WaitFor(); err != nil {
		p.t.Fatalf("%s: wait for swapped editor area: %v", name, err)
	}
	if err := expect.Locator(p.Locator(`[id="editor-area"]`)).ToHaveCount(1); err != nil {
		p.t.Fatalf("%s: editor area is not single: %v", name, err)
	}
}

func TestSitePageEditorAreaIsSingle(t *testing.T) {
	p := newPage(t)
	p.login()
	if _, err := p.Goto("/admin/site-pages/new"); err != nil {
		t.Fatalf("goto new page: %v", err)
	}
	if err := p.Locator("#name").Fill("E2E Area Page"); err != nil {
		t.Fatalf("fill name: %v", err)
	}
	if err := p.Locator("#slug").Fill("e2e-area-page"); err != nil {
		t.Fatalf("fill slug: %v", err)
	}
	if err := p.Locator("#content").Fill("Area."); err != nil {
		t.Fatalf("fill content: %v", err)
	}
	p.swapsOneEditorArea("create", func() error { return p.Locator("#editor-submit-button").Click() })
	p.swapsOneEditorArea("save", func() error { return p.Locator("#editor-submit-button").Click() })
	p.swapsOneEditorArea("new page button", func() error { return p.Locator(`aside a[href="/admin/site-pages/new"]`).Click() })
	p.swapsOneEditorArea("open from list", func() error { return p.Locator(`aside a:has-text("E2E Area Page")`).Click() })
}

func TestComponentEditorAreaIsSingle(t *testing.T) {
	p := newPage(t)
	p.login()
	if _, err := p.Goto("/admin/components/new"); err != nil {
		t.Fatalf("goto new component: %v", err)
	}
	if err := p.Locator("#name").Fill("E2E_AREA"); err != nil {
		t.Fatalf("fill name: %v", err)
	}
	if err := p.Locator("#content").Fill("Area."); err != nil {
		t.Fatalf("fill content: %v", err)
	}
	p.swapsOneEditorArea("create", func() error { return p.Locator("#editor-submit-button").Click() })
	p.swapsOneEditorArea("save", func() error { return p.Locator("#editor-submit-button").Click() })
	p.swapsOneEditorArea("new component button", func() error { return p.Locator(`aside a[href="/admin/components/new"]`).Click() })
	p.swapsOneEditorArea("open from list", func() error { return p.Locator(`aside a:has-text("E2E_AREA")`).Click() })
}
