package storage

import (
	"context"
	"fmt"
	"launay-dot-one/utils"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type StorageService struct {
	client  *minio.Client
	bucket  string
	baseURL string
}

func NewStorageService(c *minio.Client, bucket string) *StorageService {
	return &StorageService{
		client:  c,
		bucket:  bucket,
		baseURL: utils.GetEnv("STORAGE_PUBLIC_URL", "http://localhost:8080/storage"),
	}
}

func (s *StorageService) UploadFile(
	ctx context.Context,
	file multipart.File,
	header *multipart.FileHeader,
	path string,
) (string, error) {
	ext := filepath.Ext(header.Filename)
	id := uuid.New().String()
	object := fmt.Sprintf("%s/%s%s", path, id, ext)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	uploadCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	_, err := s.client.PutObject(
		uploadCtx,
		s.bucket,
		object,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/%s/%s", s.baseURL, s.bucket, object)
	return url, nil
}
