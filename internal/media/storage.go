package media

import (
	"context"
	"io"
)

type ObjectMeta struct {
	ContentType string
	Size        int64
}

type Storage interface {
	Name() string
	Put(ctx context.Context, key string, r io.Reader, meta ObjectMeta) error
	Get(ctx context.Context, key string) (io.ReadCloser, ObjectMeta, error)
	Delete(ctx context.Context, key string) error
	PublicURL(ctx context.Context, key string) (string, error)
}
