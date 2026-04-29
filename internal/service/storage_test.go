package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

	err := NewStorageService(nil, nil).TestS3Input(context.Background(), StorageInput{
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

	err := NewStorageService(nil, nil).TestS3Input(context.Background(), StorageInput{
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

func TestStorageServiceCreateS3StoresCredentialsInSecret(t *testing.T) {
	db := newStorageTestDB(t)
	secrets := newSecretTestService(t, db)
	svc := NewStorageService(repository.NewStorageRepository(db), secrets)

	storage, err := svc.CreateS3(StorageInput{
		Name:      "Object storage",
		Endpoint:  "https://s3.example.com",
		Region:    "us-east-1",
		Bucket:    "bucket",
		AccessKey: "access",
		SecretKey: "secret",
	})
	if err != nil {
		t.Fatalf("create storage: %v", err)
	}

	if storage.Config["access_key"] != nil || storage.Config["secret_key"] != nil {
		t.Fatalf("storage config must not contain plaintext credentials: %#v", storage.Config)
	}

	credentialRef := StorageConfigValue(storage, "credential_ref")
	if credentialRef == "" {
		t.Fatal("expected credential_ref")
	}

	plaintext, err := secrets.DecryptByID(credentialRef)
	if err != nil {
		t.Fatalf("decrypt credentials: %v", err)
	}

	var credentials StorageS3Credentials
	if err := json.Unmarshal(plaintext, &credentials); err != nil {
		t.Fatalf("decode credentials json: %v", err)
	}
	if credentials.AccessKey != "access" || credentials.SecretKey != "secret" {
		t.Fatalf("unexpected credentials: %#v", credentials)
	}
}

func TestStorageServiceUpdateS3ReusesCredentialRef(t *testing.T) {
	db := newStorageTestDB(t)
	secrets := newSecretTestService(t, db)
	svc := NewStorageService(repository.NewStorageRepository(db), secrets)

	storage, err := svc.CreateS3(StorageInput{
		Name:      "Object storage",
		Endpoint:  "https://s3.example.com",
		Region:    "us-east-1",
		Bucket:    "bucket",
		AccessKey: "access",
		SecretKey: "secret",
	})
	if err != nil {
		t.Fatalf("create storage: %v", err)
	}
	credentialRef := StorageConfigValue(storage, "credential_ref")

	updated, err := svc.UpdateS3(storage.ID, StorageInput{
		Name:      "Updated storage",
		Endpoint:  "https://s3.example.com",
		Region:    "us-east-1",
		Bucket:    "updated-bucket",
		AccessKey: "updated-access",
		SecretKey: "updated-secret",
	})
	if err != nil {
		t.Fatalf("update storage: %v", err)
	}

	if got := StorageConfigValue(updated, "credential_ref"); got != credentialRef {
		t.Fatalf("expected credential_ref to be reused, got %q want %q", got, credentialRef)
	}

	credentials, err := svc.S3Credentials(updated)
	if err != nil {
		t.Fatalf("read updated credentials: %v", err)
	}
	if credentials.AccessKey != "updated-access" || credentials.SecretKey != "updated-secret" {
		t.Fatalf("unexpected updated credentials: %#v", credentials)
	}
}

func newStorageTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Storage{}, &models.Secret{}); err != nil {
		t.Fatalf("migrate storage and secrets: %v", err)
	}

	return db
}
