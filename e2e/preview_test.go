//go:build e2e

package e2e

import (
	"testing"
)

// checkLivePreview types in #content and waits until the preview pane shows
// the text without a reload. previewURL is the glob of the preview request.
func (p *adminPage) checkLivePreview(previewURL string) {
	p.t.Helper()
	if err := p.Locator("#content").Fill("Live preview text"); err != nil {
		p.t.Fatalf("fill content: %v", err)
	}
	if _, err := p.ExpectResponse(previewURL, func() error {
		return p.Locator("#content").Press(".")
	}); err != nil {
		p.t.Fatalf("press . to trigger preview: %v", err)
	}
	if err := expect.Locator(p.Locator("#preview-content .preview-body")).ToContainText("Live preview text."); err != nil {
		p.t.Fatalf("assert preview body: %v", err)
	}
}

func TestLivePreviewOfPageEditor(t *testing.T) {
	p := newPage(t)
	p.login()
	if _, err := p.Goto("/admin/site-pages/new"); err != nil {
		t.Fatalf("goto /admin/site-pages/new: %v", err)
	}
	p.checkLivePreview("**/admin/site-pages/preview")
}

func TestLivePreviewOfGlobalEditor(t *testing.T) {
	p := newPage(t)
	p.login()
	if _, err := p.Goto("/admin/components/new"); err != nil {
		t.Fatalf("goto /admin/components/new: %v", err)
	}
	if err := p.Locator("#name").Fill("E2E_LIVE"); err != nil {
		t.Fatalf("fill name: %v", err)
	}
	p.checkLivePreview("**/admin/components/preview")
}

func TestLivePreviewOf404Editor(t *testing.T) {
	p := newPage(t)
	p.login()
	if _, err := p.Goto("/admin/settings/404-page"); err != nil {
		t.Fatalf("goto /admin/settings/404-page: %v", err)
	}
	p.checkLivePreview("**/admin/not-found-page/preview")
}
