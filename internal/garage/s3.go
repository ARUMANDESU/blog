package garage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/arumandesu/blog/internal/app"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var _ app.ObjectStorage = (*S3)(nil)

type S3 struct {
	bucket string
	client *minio.Client
}

type Config struct {
	Endpoint   string
	Region     string
	Bucket     string
	CredId     string
	CredSecret string
	IsSecure   bool
}

func NewS3(cfg Config) (*S3, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.CredId, cfg.CredSecret, ""),
		Secure:       cfg.IsSecure,
		Region:       cfg.Region,
		BucketLookup: minio.BucketLookupPath, // best for garage i guess
	})
	if err != nil {
		return nil, err
	}

	return &S3{
		bucket: cfg.Bucket,
		client: client,
	}, nil
}

// PutObject implements [app.ObjectStorage].
func (s *S3) PutObject(ctx context.Context, r io.Reader, s3Key string, mime string, size int64) error {
	_, err := s.client.PutObject(ctx, s.bucket, s3Key, r, size, minio.PutObjectOptions{ContentType: mime})
	return err
}

// DeleteObjects implements [app.ObjectStorage].
func (s *S3) DeleteObjects(ctx context.Context, s3Keys []string) error {
	if len(s3Keys) == 0 {
		return nil
	}
	return removeKeys(ctx, s.client, s.bucket, s3Keys)
}

func removeKeys(ctx context.Context, c *minio.Client, bucket string, keys []string) error {
	ch := make(chan minio.ObjectInfo, len(keys))
	for _, k := range keys {
		ch <- minio.ObjectInfo{Key: k}
	}
	close(ch)

	var errs []error
	for e := range c.RemoveObjects(ctx, bucket, ch, minio.RemoveObjectsOptions{}) {
		errs = append(errs, fmt.Errorf("%s: %w", e.ObjectName, e.Err))
	}
	return errors.Join(errs...)
}
