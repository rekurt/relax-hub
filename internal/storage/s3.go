package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
)

// S3Storage implements FileStorage using S3-compatible object storage (MinIO, AWS S3).
type S3Storage struct {
	client *minio.Client
	bucket string
	region string
	useSSL bool
	logger *logger.Logger
}

func NewS3Storage(cfg *config.Config, log *logger.Logger) (*S3Storage, error) {
	client, err := minio.New(cfg.Storage.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.AccessKey, cfg.Storage.SecretKey, ""),
		Secure: cfg.Storage.UseSSL,
		Region: cfg.Storage.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("create s3 client: %w", err)
	}

	s := &S3Storage{
		client: client,
		bucket: cfg.Storage.Bucket,
		region: cfg.Storage.Region,
		useSSL: cfg.Storage.UseSSL,
		logger: log,
	}

	if err := s.ensureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("ensure bucket: %w", err)
	}

	return s, nil
}

func (s *S3Storage) Upload(ctx context.Context, filename string, data io.Reader, contentType string) (string, error) {
	if s == nil {
		return "", nil
	}
	buf, err := io.ReadAll(data)
	if err != nil {
		return "", fmt.Errorf("read data: %w", err)
	}

	_, err = s.client.PutObject(ctx, s.bucket, filename, bytes.NewReader(buf), int64(len(buf)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload to s3: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", s.endpointURL(), s.bucket, filename)
	return url, nil
}

func (s *S3Storage) Delete(ctx context.Context, filename string) error {
	if s == nil {
		return nil
	}
	err := s.client.RemoveObject(ctx, s.bucket, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete from s3: %w", err)
	}
	return nil
}

func (s *S3Storage) endpointURL() string {
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, s.client.EndpointURL().Host)
}

func (s *S3Storage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{
			Region: s.region,
		})
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		s.logger.Info("created storage bucket", "bucket", s.bucket)
	}
	return nil
}

// AvatarPath returns the S3 key for an avatar file.
func AvatarPath(userID, suffix, ext string) string {
	return path.Join("avatars", userID, suffix+ext)
}
