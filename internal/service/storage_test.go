package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStorageServiceTestS3InputRunsPutGetDelete(t *testing.T) {
	var sequence []string
	objects := map[string]string{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/bucket/umono-test/") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		switch r.Method {
		case http.MethodPut:
			sequence = append(sequence, "put")
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read put body: %v", err)
			}
			objects[r.URL.Path] = string(body)
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			sequence = append(sequence, "get")
			body, ok := objects[r.URL.Path]
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte(body))
		case http.MethodDelete:
			sequence = append(sequence, "delete")
			delete(objects, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	}))
	defer server.Close()

	err := NewStorageService(nil).TestS3Input(context.Background(), StorageInput{
		Name:      "Object storage",
		Endpoint:  server.URL,
		Region:    "us-east-1",
		Bucket:    "bucket",
		AccessKey: "key",
		SecretKey: "secret",
	})
	if err != nil {
		t.Fatalf("test s3 input failed: %v", err)
	}

	wantSequence := []string{"put", "get", "delete"}
	if strings.Join(sequence, ",") != strings.Join(wantSequence, ",") {
		t.Fatalf("unexpected sequence: got %v want %v", sequence, wantSequence)
	}
	if len(objects) != 0 {
		t.Fatalf("test object was not deleted: %#v", objects)
	}
}

func TestStorageServiceTestS3InputStopsBeforeDeleteWhenGetFails(t *testing.T) {
	var deleteCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			http.Error(w, "get failed", http.StatusInternalServerError)
		case http.MethodDelete:
			deleteCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	}))
	defer server.Close()

	err := NewStorageService(nil).TestS3Input(context.Background(), StorageInput{
		Name:      "Object storage",
		Endpoint:  server.URL,
		Region:    "us-east-1",
		Bucket:    "bucket",
		AccessKey: "key",
		SecretKey: "secret",
	})
	if err == nil {
		t.Fatal("expected test error")
	}
	if deleteCalled {
		t.Fatal("delete should not run after get failure")
	}

	var testErr *StorageTestError
	if !errors.As(err, &testErr) {
		t.Fatalf("expected StorageTestError, got %T", err)
	}
	if testErr.Step != "get" {
		t.Fatalf("unexpected failed step: got %q want get", testErr.Step)
	}
}
