package credentials

import (
	"encoding/base64"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashRejectsEmptyValues(t *testing.T) {
	for name, value := range map[string]string{
		"empty":      "",
		"spaces":     "   ",
		"whitespace": "\t\n ",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Hash(value); !errors.Is(err, ErrEmptyValue) {
				t.Fatalf("expected ErrEmptyValue, got %v", err)
			}
		})
	}
}

func TestHashRoundTrip(t *testing.T) {
	hashed, err := Hash("s3cret")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	decoded, err := DecodeHash(Encode(hashed))
	if err != nil {
		t.Fatalf("decode hash: %v", err)
	}

	if !Matches(decoded, "s3cret") {
		t.Fatal("expected hash to match the original value")
	}
	if Matches(decoded, "") {
		t.Fatal("expected hash not to match an empty value")
	}
}

func TestDecodeHashRejectsUnusableValues(t *testing.T) {
	notBcrypt := base64.StdEncoding.EncodeToString([]byte("not-a-bcrypt-hash"))

	for name, encoded := range map[string]string{
		"unset":          "",
		"blank":          "   ",
		"not base64":     "!!!not-base64!!!",
		"not bcrypt":     notBcrypt,
		"empty base64":   base64.StdEncoding.EncodeToString(nil),
		"truncated hash": base64.StdEncoding.EncodeToString([]byte("$2a$10$short")),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeHash(encoded); err == nil {
				t.Fatal("expected an error for an unusable hash")
			}
		})
	}
}

func TestLoadHashesRequiresBothHashes(t *testing.T) {
	hashed, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.StdEncoding.EncodeToString(hashed)

	cases := map[string]struct{ username, password string }{
		"both unset":      {"", ""},
		"username unset":  {"", encoded},
		"password unset":  {encoded, ""},
		"password broken": {encoded, "not-base64!!!"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Setenv(HashedUsernameEnvKey, tc.username)
			t.Setenv(HashedPasswordEnvKey, tc.password)

			if _, _, err := LoadHashes(); err == nil {
				t.Fatal("expected LoadHashes to fail")
			}
		})
	}

	t.Run("both set", func(t *testing.T) {
		t.Setenv(HashedUsernameEnvKey, encoded)
		t.Setenv(HashedPasswordEnvKey, encoded)

		if _, _, err := LoadHashes(); err != nil {
			t.Fatalf("expected LoadHashes to succeed, got %v", err)
		}
	})
}
