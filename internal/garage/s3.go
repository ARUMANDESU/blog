package garage

import (
	"context"
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
