package main

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func TestUpdateEnvFilePreservesUmonoSecret(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	const umonoSecret = "persisted-secret"
	if err := os.WriteFile(".env", []byte("UMONO_SECRET="+umonoSecret+"\nUSERNAME=admin\nPASSWORD=admin\n"), 0o666); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	t.Setenv("APP_ENV", "dev")
	t.Setenv("SESSION_DRIVER", "db")
	t.Setenv("PORT", "8999")
	t.Setenv("DSN", "umono.db")
	t.Setenv("USERNAME", "new-admin")
	t.Setenv("PASSWORD", "new-password")

	if err := updateEnvFile(); err != nil {
		t.Fatalf("update env file: %v", err)
	}

	envFile, err := godotenv.Read(".env")
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}

	if envFile["UMONO_SECRET"] != umonoSecret {
		t.Fatalf("expected UMONO_SECRET to be preserved, got %q", envFile["UMONO_SECRET"])
	}
	if os.Getenv("UMONO_SECRET") != "" {
		t.Fatal("expected UMONO_SECRET to stay out of process env")
	}
	if envFile["USERNAME"] != "" {
		t.Fatalf("expected USERNAME to be cleared, got %q", envFile["USERNAME"])
	}
	if envFile["PASSWORD"] != "" {
		t.Fatalf("expected PASSWORD to be cleared, got %q", envFile["PASSWORD"])
	}

	hashedUsername, err := base64.StdEncoding.DecodeString(envFile["HASHED_USERNAME"])
	if err != nil {
		t.Fatalf("decode hashed username: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword(hashedUsername, []byte("new-admin")); err != nil {
		t.Fatalf("expected username hash to match: %v", err)
	}

	hashedPassword, err := base64.StdEncoding.DecodeString(envFile["HASHED_PASSWORD"])
	if err != nil {
		t.Fatalf("decode hashed password: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte("new-password")); err != nil {
		t.Fatalf("expected password hash to match: %v", err)
	}
}
