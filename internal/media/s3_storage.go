package media

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

type S3Storage struct {
	endpoint  string
	bucket    string
	client    *s3.Client
	presigner *s3.PresignClient
	uploader  *manager.Uploader
}

func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(strings.TrimSpace(cfg.Region)),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			strings.TrimSpace(cfg.AccessKey),
			strings.TrimSpace(cfg.SecretKey),
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	return &S3Storage{
		endpoint:  endpoint,
		bucket:    strings.TrimSpace(cfg.Bucket),
		client:    client,
		presigner: s3.NewPresignClient(client),
		uploader:  manager.NewUploader(client),
	}, nil
}

func (*S3Storage) Name() string {
	return "s3"
}

func (s *S3Storage) Put(ctx context.Context, key string, r io.Reader, meta ObjectMeta) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   r,
	}
	if meta.ContentType != "" {
		input.ContentType = aws.String(meta.ContentType)
	}
	if meta.CacheControl != "" {
		input.CacheControl = aws.String(meta.CacheControl)
	}

	_, err := s.uploader.Upload(ctx, input)
	return err
}

func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, ObjectMeta, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, ObjectMeta{}, err
	}

	meta := ObjectMeta{
		Size: aws.ToInt64(output.ContentLength),
	}
	if output.ContentType != nil {
		meta.ContentType = *output.ContentType
	}

	return output.Body, meta, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Storage) PublicURL(ctx context.Context, key string) (string, error) {
	baseURL, err := url.Parse(s.endpoint)
	if err != nil {
		return "", err
	}

	basePath := baseURL.Path
	baseURL.Path = joinURLPath(basePath, s.bucket, key)
	baseURL.RawPath = joinEscapedURLPath(basePath, s.bucket, key)
	baseURL.RawQuery = ""
	baseURL.Fragment = ""

	return baseURL.String(), nil
}

func (s *S3Storage) PresignPut(ctx context.Context, key string, meta ObjectMeta) (string, map[string]string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	if meta.ContentType != "" {
		input.ContentType = aws.String(meta.ContentType)
	}
	if meta.CacheControl != "" {
		input.CacheControl = aws.String(meta.CacheControl)
	}
	if meta.Size > 0 {
		input.ContentLength = aws.Int64(meta.Size)
	}

	output, err := s.presigner.PresignPutObject(ctx, input)
	if err != nil {
		return "", nil, err
	}

	return output.URL, flattenHeaders(output.SignedHeader), nil
}

func (s *S3Storage) PresignGet(ctx context.Context, key string) (string, error) {
	output, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}

	return output.URL, nil
}

func flattenHeaders(headers http.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	out := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}
		out[key] = strings.Join(values, ", ")
	}

	return out
}

func (s *S3Storage) String() string {
	return fmt.Sprintf("s3:%s", s.bucket)
}

func joinURLPath(parts ...string) string {
	joined := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		for _, segment := range strings.Split(strings.Trim(part, "/"), "/") {
			if segment == "" {
				continue
			}
			joined = append(joined, segment)
		}
	}

	if len(joined) == 0 {
		return "/"
	}

	return "/" + path.Join(joined...)
}

func joinEscapedURLPath(parts ...string) string {
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		for _, segment := range strings.Split(strings.Trim(part, "/"), "/") {
			if segment == "" {
				continue
			}
			escaped = append(escaped, url.PathEscape(segment))
		}
	}

	if len(escaped) == 0 {
		return "/"
	}

	return "/" + path.Join(escaped...)
}
