package credentials

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	UsernameEnvKey       = "USERNAME"
	PasswordEnvKey       = "PASSWORD"
	HashedUsernameEnvKey = "HASHED_USERNAME"
	HashedPasswordEnvKey = "HASHED_PASSWORD"
)

var (
	ErrEmptyValue = errors.New("credential is empty")
	ErrHashNotSet = errors.New("credential hash is not set")
)

func IsEmpty(value string) bool {
	return strings.TrimSpace(value) == ""
}

func Hash(value string) (string, error) {
	if IsEmpty(value) {
		return "", ErrEmptyValue
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func Encode(hashed string) string {
	return base64.StdEncoding.EncodeToString([]byte(hashed))
}

func DecodeHash(encoded string) ([]byte, error) {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return nil, ErrHashNotSet
	}

	hashed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	if _, err := bcrypt.Cost(hashed); err != nil {
		return nil, err
	}

	return hashed, nil
}

func LoadHashes() (username, password []byte, err error) {
	username, err = DecodeHash(os.Getenv(HashedUsernameEnvKey))
	if err != nil {
		return nil, nil, err
	}

	password, err = DecodeHash(os.Getenv(HashedPasswordEnvKey))
	if err != nil {
		return nil, nil, err
	}

	return username, password, nil
}

func Matches(hashed []byte, value string) bool {
	return bcrypt.CompareHashAndPassword(hashed, []byte(value)) == nil
}
