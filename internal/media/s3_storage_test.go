package media

import (
	"context"
	"testing"
)

func TestS3StoragePublicURLBuildsPathStyleURL(t *testing.T) {
	t.Parallel()

	storage, err := NewS3Storage(context.Background(), S3Config{
		Endpoint:  "https://s3.example.com",
		Region:    "eu-central-1",
		Bucket:    "umono-media",
		AccessKey: "key",
		SecretKey: "secret",
	})
	if err != nil {
		t.Fatalf("new s3 storage failed: %v", err)
	}

	got, err := storage.PublicURL(context.Background(), "uploads/hero image.webp")
	if err != nil {
		t.Fatalf("public url failed: %v", err)
	}

	want := "https://s3.example.com/umono-media/uploads/hero%20image.webp"
	if got != want {
		t.Fatalf("unexpected url: got %q want %q", got, want)
	}
}

func TestS3StoragePublicURLKeepsEndpointPathPrefix(t *testing.T) {
	t.Parallel()

	storage, err := NewS3Storage(context.Background(), S3Config{
		Endpoint:  "https://cdn.example.com/storage",
		Region:    "eu-central-1",
		Bucket:    "umono-media",
		AccessKey: "key",
		SecretKey: "secret",
	})
	if err != nil {
		t.Fatalf("new s3 storage failed: %v", err)
	}

	got, err := storage.PublicURL(context.Background(), "uploads/poster.webp")
	if err != nil {
		t.Fatalf("public url failed: %v", err)
	}

	want := "https://cdn.example.com/storage/umono-media/uploads/poster.webp"
	if got != want {
		t.Fatalf("unexpected url: got %q want %q", got, want)
	}
}
