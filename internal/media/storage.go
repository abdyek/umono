package media

import (
	"context"
	"errors"
	"io"
)

type ObjectMeta struct {
	ContentType string
	Size        int64
}

var ErrPresignUnsupported = errors.New("presign unsupported")

type Storage interface {
	Name() string
	Put(ctx context.Context, key string, r io.Reader, meta ObjectMeta) error
	Get(ctx context.Context, key string) (io.ReadCloser, ObjectMeta, error)
	Delete(ctx context.Context, key string) error
	PublicURL(ctx context.Context, key string) (string, error)
	PresignPut(ctx context.Context, key string, meta ObjectMeta) (url string, headers map[string]string, err error)
	PresignGet(ctx context.Context, key string) (url string, err error)
}
