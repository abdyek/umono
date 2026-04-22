package media

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestLocalStorageStoresUploadsKeyUnderRootWithoutDuplicatingUploadsDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	storage := NewLocalStorage(root)

	if err := storage.Put(context.Background(), "uploads/test.webp", strings.NewReader("ok"), ObjectMeta{}); err != nil {
		t.Fatalf("put failed: %v", err)
	}

	reader, _, err := storage.Get(context.Background(), "uploads/test.webp")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(body) != "ok" {
		t.Fatalf("unexpected body: %q", string(body))
	}
}

func TestLocalStoragePublicURLTrimsStoragePrefix(t *testing.T) {
	t.Parallel()

	url, err := NewLocalStorage("uploads").PublicURL(context.Background(), "uploads/test.webp")
	if err != nil {
		t.Fatalf("public url failed: %v", err)
	}
	if url != "/uploads/test.webp" {
		t.Fatalf("unexpected url: %s", url)
	}
}
