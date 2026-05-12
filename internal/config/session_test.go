package config

import "testing"

func TestNewSessionStoreCookieSecurity(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("SESSION_DRIVER", "")

	store := NewSessionStore()

	if store.CookieHTTPOnly != true {
		t.Fatal("expected session cookie to be HTTP-only")
	}
	if store.CookiePath != "/admin" {
		t.Fatalf("expected session cookie path /admin, got %q", store.CookiePath)
	}
	if store.CookieSameSite != "strict" {
		t.Fatalf("expected session cookie SameSite strict, got %q", store.CookieSameSite)
	}
	if store.CookieSecure != false {
		t.Fatal("expected session cookie secure to be false outside prod")
	}
}

func TestNewSessionStoreCookieSecureInProd(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("SESSION_DRIVER", "")

	store := NewSessionStore()

	if store.CookieSecure != true {
		t.Fatal("expected session cookie secure to be true in prod")
	}
}
