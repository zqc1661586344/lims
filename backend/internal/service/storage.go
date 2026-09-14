package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"lims-backend/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type StorageService interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, objectName string) (io.ReadCloser, *minio.ObjectInfo, error)
	Delete(ctx context.Context, objectName string) error
	GetPresignedURL(ctx context.Context, objectName string, expires time.Duration) (string, error)
	EnsureBucket(ctx context.Context) error
}

type minioStorage struct {
	client     *minio.Client
	bucketName string
	logger     *zap.Logger
}

func NewStorageService(cfg *config.MinIOConfig, logger *zap.Logger) (StorageService, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	return &minioStorage{
		client:     client,
		bucketName: cfg.BucketName,
		logger:     logger,
	}, nil
}

func (s *minioStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("check bucket failed: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket failed: %w", err)
		}
		s.logger.Info("MinIO bucket created", zap.String("bucket", s.bucketName))
	}
	return nil
}

func (s *minioStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	opts := minio.PutObjectOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, size, opts)
	if err != nil {
		return fmt.Errorf("minio upload failed: %w", err)
	}
	return nil
}

func (s *minioStorage) Download(ctx context.Context, objectName string) (io.ReadCloser, *minio.ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, s.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("stat object failed: %w", err)
	}
	reader, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("minio download failed: %w", err)
	}
	return reader, &info, nil
}

func (s *minioStorage) Delete(ctx context.Context, objectName string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio delete failed: %w", err)
	}
	return nil
}

func (s *minioStorage) GetPresignedURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	if expires <= 0 {
		expires = 15 * time.Minute
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucketName, objectName, expires, nil)
	if err != nil {
		return "", fmt.Errorf("presign failed: %w", err)
	}
	return u.String(), nil
}

var ErrStorageUnavailable = errors.New("storage service unavailable")

type noopStorage struct{}

func (noopStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	return ErrStorageUnavailable
}
func (noopStorage) Download(ctx context.Context, objectName string) (io.ReadCloser, *minio.ObjectInfo, error) {
	return nil, nil, ErrStorageUnavailable
}
func (noopStorage) Delete(ctx context.Context, objectName string) error {
	return ErrStorageUnavailable
}
func (noopStorage) GetPresignedURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	return "", ErrStorageUnavailable
}
func (noopStorage) EnsureBucket(ctx context.Context) error { return nil }

func NewNoopStorage() StorageService { return noopStorage{} }
