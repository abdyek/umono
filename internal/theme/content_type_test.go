package theme

import "testing"

func TestContentType(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"theme.css", "text/css; charset=utf-8"},
		{"fonts/a.woff2", "font/woff2"},
		{"a.woff", "font/woff"},
		{"img/a.png", "image/png"},
		{"a.jpeg", "image/jpeg"},
		{"a.jpg", "image/jpeg"},
		{"a.webp", "image/webp"},
		{"a.avif", "image/avif"},
		{"a.gif", "image/gif"},
		{"other.css", ""},
		{"a.svg", ""},
		{"a.PNG", ""},
		{"LICENSE", ""},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			if got := ContentType(c.name); got != c.want {
				t.Errorf("ContentType(%q) = %q, want %q", c.name, got, c.want)
			}
		})
	}
}
