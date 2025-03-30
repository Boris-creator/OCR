package s3

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"tele/internal/config"
)

type Storage struct {
	minio.Client
	cfg config.S3Config
}

func New(client minio.Client, cfg config.S3Config) *Storage {
	return &Storage{client, cfg}
}

func (storage *Storage) UploadFromLocal(ctx context.Context, localFilePath, objectName string) error {
	_, err := storage.FPutObject(ctx, storage.cfg.BucketName, objectName, localFilePath, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("client.FPutObject: %w", err)
	}
	return nil
}
