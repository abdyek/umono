package service

import (
	"bytes"
	"strings"
	"testing"

	umonocrypto "github.com/umono-cms/crypto"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSecretSchemaUsesStrictTableAndColumns(t *testing.T) {
	db := newSecretTestDB(t)

	rows, err := db.Raw("PRAGMA table_info(secrets)").Rows()
	if err != nil {
		t.Fatalf("read secrets schema: %v", err)
	}
	defer rows.Close()

	columns := map[string]string{}
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, pk int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			t.Fatalf("scan secrets schema: %v", err)
		}
		columns[name] = strings.ToLower(columnType)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate secrets schema: %v", err)
	}

	if len(columns) != 2 {
		t.Fatalf("expected secrets table to have exactly 2 columns, got %d: %#v", len(columns), columns)
	}
	if columns["id"] != "text" {
		t.Fatalf("expected id to be text, got %q", columns["id"])
	}
	if columns["ciphertext"] != "blob" {
		t.Fatalf("expected ciphertext to be blob, got %q", columns["ciphertext"])
	}
}

func TestSecretServiceCreateDecryptAndUpdate(t *testing.T) {
	db := newSecretTestDB(t)
	svc := newSecretTestService(t, db)

	created, err := svc.Create([]byte("sensitive-value"))
	if err != nil {
		t.Fatalf("create secret: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected generated id")
	}
	if bytes.Equal(created.Ciphertext, []byte("sensitive-value")) {
		t.Fatal("expected ciphertext to differ from plaintext")
	}

	plaintext, err := svc.DecryptByID(created.ID)
	if err != nil {
		t.Fatalf("decrypt secret: %v", err)
	}
	if string(plaintext) != "sensitive-value" {
		t.Fatalf("expected decrypted plaintext, got %q", plaintext)
	}

	updated, err := svc.Update(created.ID, []byte("updated-sensitive-value"))
	if err != nil {
		t.Fatalf("update secret: %v", err)
	}
	if bytes.Equal(updated.Ciphertext, created.Ciphertext) {
		t.Fatal("expected ciphertext to change after update")
	}

	plaintext, err = svc.DecryptByID(created.ID)
	if err != nil {
		t.Fatalf("decrypt updated secret: %v", err)
	}
	if string(plaintext) != "updated-sensitive-value" {
		t.Fatalf("expected updated plaintext, got %q", plaintext)
	}
}

func newSecretTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Secret{}); err != nil {
		t.Fatalf("migrate secrets: %v", err)
	}

	return db
}

func newSecretTestService(t *testing.T, db *gorm.DB) *SecretService {
	t.Helper()

	key, err := umonocrypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	crypto, err := umonocrypto.New(key, []byte("umono-secrets"))
	if err != nil {
		t.Fatalf("create crypto secret: %v", err)
	}

	return NewSecretService(repository.NewSecretRepository(db), crypto)
}
