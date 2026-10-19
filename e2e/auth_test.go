//go:build e2e

package e2e

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
)

var expect = playwright.NewPlaywrightAssertions()

func TestLoginRejectsWrongPassword(t *testing.T) {
	p := newPage(t)
	if _, err := p.Goto("/admin/login"); err != nil {
		t.Fatalf("goto /admin/login: %v", err)
	}
	if err := p.Locator("#username").Fill(testUsername); err != nil {
		t.Fatalf("fill username: %v", err)
	}
	if err := p.Locator("#password").Fill("wrong-password"); err != nil {
		t.Fatalf("fill password: %v", err)
	}
	if err := p.Locator("#submitBtn").Click(); err != nil {
		t.Fatalf("click submit: %v", err)
	}
	if err := expect.Locator(p.Locator("#login-result")).ToContainText("Invalid username or password."); err != nil {
		t.Fatalf("assert login result: %v", err)
	}
	if err := expect.Page(p).ToHaveURL(baseURL + "/admin/login"); err != nil {
		t.Fatalf("assert url: %v", err)
	}
}

func TestLoginAndLogout(t *testing.T) {
	p := newPage(t)
	p.login()
	if err := expect.Locator(p.Locator("#logout-button")).ToBeVisible(); err != nil {
		t.Fatalf("assert logout button visible: %v", err)
	}
	if err := p.Locator("#logout-button").Click(); err != nil {
		t.Fatalf("click logout: %v", err)
	}
	if err := p.WaitForURL(baseURL + "/admin/login"); err != nil {
		t.Fatalf("wait for /admin/login: %v", err)
	}
	if _, err := p.Goto("/admin"); err != nil {
		t.Fatalf("goto /admin: %v", err)
	}
	if err := expect.Page(p).ToHaveURL(baseURL + "/admin/login"); err != nil {
		t.Fatalf("assert url: %v", err)
	}
}
