package theme

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/umono-cms/umono/internal/models"
)

func TestRegistryVerify(t *testing.T) {
	src, err := os.ReadFile("testdata/package/contract.toml")
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}

	var saved *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		saved = r.Clone(context.Background())
		_, _ = w.Write(src)
	}))
	defer srv.Close()

	ref := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}
	ctx := context.Background()

	if err := (Registry{BaseURL: srv.URL, Client: srv.Client()}).Verify(ctx, ref, src); err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if saved.Method != "GET" {
		t.Fatalf("method = %q, want GET", saved.Method)
	}
	if got, want := saved.URL.Path, "/registry/v1/contracts/example/blog/v1.0"; got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
	if saved.URL.RawQuery != "" {
		t.Fatalf("raw query = %q, want empty", saved.URL.RawQuery)
	}
	if got := saved.Header.Get("User-Agent"); got != "Umono" {
		t.Fatalf("User-Agent = %q, want Umono", got)
	}
	if got := saved.Header.Get("Cookie"); got != "" {
		t.Fatalf("Cookie = %q, want empty", got)
	}
	if got := saved.Header.Get("Referer"); got != "" {
		t.Fatalf("Referer = %q, want empty", got)
	}
}

func TestRegistryVerifyRejectsAnotherFile(t *testing.T) {
	src, err := os.ReadFile("testdata/package/contract.toml")
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(src)
		_, _ = w.Write([]byte("\n"))
	}))
	defer srv.Close()

	ref := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}
	ctx := context.Background()

	err = (Registry{BaseURL: srv.URL, Client: srv.Client()}).Verify(ctx, ref, src)
	if err == nil {
		t.Fatalf("Verify returned nil, want error")
	}
	if !errors.Is(err, ErrContractMismatch) {
		t.Fatalf("err = %v, want ErrContractMismatch", err)
	}
}

func TestRegistryVerifyNotPublished(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	ref := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}
	ctx := context.Background()

	src, err := os.ReadFile("testdata/package/contract.toml")
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}

	err = (Registry{BaseURL: srv.URL, Client: srv.Client()}).Verify(ctx, ref, src)
	if err == nil {
		t.Fatalf("Verify returned nil, want error")
	}
	if !errors.Is(err, ErrContractNotPublished) {
		t.Fatalf("err = %v, want ErrContractNotPublished", err)
	}
}

func TestRegistryVerifyServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ref := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}
	ctx := context.Background()

	src, err := os.ReadFile("testdata/package/contract.toml")
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}

	err = (Registry{BaseURL: srv.URL, Client: srv.Client()}).Verify(ctx, ref, src)
	if err == nil {
		t.Fatalf("Verify returned nil, want error")
	}
	if !errors.Is(err, ErrRegistryUnreachable) {
		t.Fatalf("err = %v, want ErrRegistryUnreachable", err)
	}
}

func TestRegistryVerifyUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	baseURL := srv.URL
	client := srv.Client()
	srv.Close()

	ref := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}
	ctx := context.Background()

	src, err := os.ReadFile("testdata/package/contract.toml")
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}

	err = (Registry{BaseURL: baseURL, Client: client}).Verify(ctx, ref, src)
	if err == nil {
		t.Fatalf("Verify returned nil, want error")
	}
	if !errors.Is(err, ErrRegistryUnreachable) {
		t.Fatalf("err = %v, want ErrRegistryUnreachable", err)
	}
}
