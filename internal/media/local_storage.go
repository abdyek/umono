package media

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) *LocalStorage {
	return &LocalStorage{root: root}
}

func (*LocalStorage) Name() string {
	return "local"
}

func (s *LocalStorage) Put(_ context.Context, key string, r io.Reader, _ ObjectMeta) error {
	path, err := s.resolvePath(key)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	return err
}

func (s *LocalStorage) Get(_ context.Context, key string) (io.ReadCloser, ObjectMeta, error) {
	path, err := s.resolvePath(key)
	if err != nil {
		return nil, ObjectMeta{}, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, ObjectMeta{}, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, ObjectMeta{}, err
	}

	return file, ObjectMeta{Size: info.Size()}, nil
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	path, err := s.resolvePath(key)
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (*LocalStorage) PublicURL(_ context.Context, key string) (string, error) {
	clean := strings.TrimPrefix(normalizeLocalKey(key), "uploads/")
	escaped := make([]string, 0)
	for _, segment := range strings.Split(clean, "/") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		escaped = append(escaped, url.PathEscape(segment))
	}
	return "/uploads/" + strings.Join(escaped, "/"), nil
}

func (*LocalStorage) PresignPut(_ context.Context, _ string, _ ObjectMeta) (string, map[string]string, error) {
	return "", nil, ErrPresignUnsupported
}

func (*LocalStorage) PresignGet(_ context.Context, _ string) (string, error) {
	return "", ErrPresignUnsupported
}

func (s *LocalStorage) resolvePath(key string) (string, error) {
	clean := filepath.Clean(normalizeLocalKey(key))
	if clean == "." || clean == "" || filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", os.ErrPermission
	}

	return filepath.Join(s.root, clean), nil
}

func normalizeLocalKey(key string) string {
	key = filepath.ToSlash(strings.TrimSpace(key))
	return strings.TrimPrefix(key, "uploads/")
}
