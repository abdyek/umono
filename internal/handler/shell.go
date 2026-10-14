package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// shellFile is a file of the shell with its content type and ETag.
type shellFile struct {
	contentType string
	body        []byte
	etag        string
}

// shellHandler serves the files of the shell under /_umono/. Their content
// changes with the Umono version, so they are served with no-cache and an
// ETag instead of the immutable cache of /static.
type shellHandler struct {
	files map[string]shellFile
}

// NewShellHandler reads umono-shell.css and umono-root.js from fsys.
func NewShellHandler(fsys fs.FS) (*shellHandler, error) {
	files := map[string]shellFile{}

	for _, entry := range []struct {
		name        string
		contentType string
	}{
		{"umono-shell.css", "text/css; charset=utf-8"},
		{"umono-root.js", "text/javascript; charset=utf-8"},
	} {
		body, err := fs.ReadFile(fsys, entry.name)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(body)
		files[entry.name] = shellFile{
			contentType: entry.contentType,
			body:        body,
			etag:        strconv.Quote(hex.EncodeToString(sum[:])),
		}
	}

	return &shellHandler{files: files}, nil
}

func (h *shellHandler) Serve(c *fiber.Ctx) error {
	file, ok := h.files[c.Params("file")]
	if !ok {
		return c.SendStatus(fiber.StatusNotFound)
	}

	c.Set("Content-Type", file.contentType)
	c.Set("Cache-Control", "no-cache")
	c.Set("ETag", file.etag)

	if etagMatches(c.Get("If-None-Match"), file.etag) {
		return c.SendStatus(fiber.StatusNotModified)
	}

	return c.Send(file.body)
}

// etagMatches reports whether the If-None-Match header matches etag with the
// weak comparison: "*" matches, and each listed tag matches when it equals
// etag after a leading W/ is removed from both.
func etagMatches(header, etag string) bool {
	if header == "" {
		return false
	}
	etag = strings.TrimPrefix(etag, "W/")
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if part == "*" || strings.TrimPrefix(part, "W/") == etag {
			return true
		}
	}
	return false
}
