package csp

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestNonce(t *testing.T) {
	n1, err := Nonce()
	if err != nil {
		t.Fatalf("Nonce() first call returned error: %v", err)
	}
	n2, err := Nonce()
	if err != nil {
		t.Fatalf("Nonce() second call returned error: %v", err)
	}
	if n1 == n2 {
		t.Fatalf("Nonce() returned equal values: got %q, want distinct value", n1)
	}
	decoded, err := base64.StdEncoding.DecodeString(n1)
	if err != nil {
		t.Fatalf("base64.StdEncoding.DecodeString(%q) returned error: %v", n1, err)
	}
	if got, want := len(decoded), 16; got != want {
		t.Fatalf("base64.StdEncoding.DecodeString(%q) length: got %d, want %d", n1, got, want)
	}
}

func TestImageOrigins(t *testing.T) {
	self := "https://example.com"
	urls := []string{
		"/uploads/a.png",
		"https://example.com/x.png",
		"https://cdn.example.net/a.png?x=1#f",
		"http://cdn.example.net:8080/a.png",
		"https://user:pass@img.example.org/a.png",
		"//cdn2.example.net/a.png",
		"ftp://files.example.net/a.png",
		"data:image/png;base64,xx",
		"https://*.example.net/a.png",
		"https://a,b.example/x.png",
		"https://a;b.example/x.png",
		"https://cdn.example.net/b.png",
		"HTTPS://CDN.Example.NET/c.png",
		"",
	}
	want := []string{"https://cdn.example.net", "http://cdn.example.net:8080", "https://img.example.org", "cdn2.example.net"}
	got := ImageOrigins(urls, self)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ImageOrigins(urls, %q) = %#v, want %#v", self, got, want)
	}

	if got := ImageOrigins(nil, self); len(got) != 0 {
		t.Fatalf("ImageOrigins(nil, %q) length: got %d, want %d", self, len(got), 0)
	}

	got = ImageOrigins([]string{"https://example.com/x.png"}, "")
	want = []string{"https://example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ImageOrigins([\"https://example.com/x.png\"], \"\") = %#v, want %#v", got, want)
	}
}

func TestHeader(t *testing.T) {
	got := Header("abc", []string{"https://cdn.example.net", "cdn2.example.net"})
	want := "default-src 'none'; script-src 'nonce-abc'; style-src 'self'; img-src 'self' https://cdn.example.net cdn2.example.net; font-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"
	if got != want {
		t.Fatalf("Header(\"abc\", [\"https://cdn.example.net\", \"cdn2.example.net\"]) = %q, want %q", got, want)
	}

	got = Header("abc", nil)
	want = "default-src 'none'; script-src 'nonce-abc'; style-src 'self'; img-src 'self'; font-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"
	if got != want {
		t.Fatalf("Header(\"abc\", nil) = %q, want %q", got, want)
	}
}
