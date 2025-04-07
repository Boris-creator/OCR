package s3

import (
	"context"
	"fmt"
	"tele/internal/config"

	"github.com/minio/minio-go/v7"
)

type Storage struct {
	minio.Client
	cfg config.S3Config
}

func New(client minio.Client, cfg config.S3Config) *Storage {
	return &Storage{client, cfg}
}

func (storage *Storage) UploadFromLocal(ctx context.Context, localFilePath, objectName string) error {
	var options minio.PutObjectOptions

	_, err := storage.FPutObject(ctx, storage.cfg.BucketName, objectName, localFilePath, options)
	if err != nil {
		return fmt.Errorf("client.FPutObject: %w", err)
	}

	return nil
}
