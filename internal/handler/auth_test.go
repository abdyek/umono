package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginRegeneratesSessionID(t *testing.T) {
	setAuthEnv(t, "admin", "secret")

	store := testSessionStore()
	auth := NewAuthHandler(store)
	app := fiber.New()
	app.Get("/seed", seedSession(store))
	app.Post("/login", auth.Login)

	oldCookie := requestSessionCookie(t, app, "/seed")
	oldID := oldCookie.Value

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("username=admin&password=secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(oldCookie)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	newCookie := findCookie(t, resp.Cookies(), "session_id")
	newID := newCookie.Value
	if newID == oldID {
		t.Fatal("expected login to regenerate session id")
	}

	rawOld, err := store.Storage.Get(oldID)
	if err != nil {
		t.Fatal(err)
	}
	if rawOld != nil {
		t.Fatal("expected old session id to be deleted from storage")
	}

	rawNew, err := store.Storage.Get(newID)
	if err != nil {
		t.Fatal(err)
	}
	if rawNew == nil {
		t.Fatal("expected regenerated session id to be saved")
	}
}

func TestLogoutDestroysSession(t *testing.T) {
	store := testSessionStore()
	auth := NewAuthHandler(store)
	app := fiber.New()
	app.Get("/seed", seedSession(store))
	app.Post("/logout", auth.Logout)

	cookie := requestSessionCookie(t, app, "/seed")

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(cookie)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", resp.StatusCode)
	}

	raw, err := store.Storage.Get(cookie.Value)
	if err != nil {
		t.Fatal(err)
	}
	if raw != nil {
		t.Fatal("expected logout to delete session from storage")
	}
}

func setAuthEnv(t *testing.T, username, password string) {
	t.Helper()

	hashedUsername, err := bcrypt.GenerateFromPassword([]byte(username), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("HASHED_USERNAME", base64.StdEncoding.EncodeToString(hashedUsername))
	t.Setenv("HASHED_PASSWORD", base64.StdEncoding.EncodeToString(hashedPassword))
}

func testSessionStore() *session.Store {
	var id int
	return session.New(session.Config{
		KeyGenerator: func() string {
			id++
			return fmt.Sprintf("session-%d", id)
		},
	})
}

func seedSession(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return err
		}
		sess.Set("username", "admin")
		if err := sess.Save(); err != nil {
			return err
		}
		return nil
	}
}

func requestSessionCookie(t *testing.T, app *fiber.App, path string) *http.Cookie {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	return findCookie(t, resp.Cookies(), "session_id")
}

func findCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}

	t.Fatalf("expected %q cookie", name)
	return nil
}
