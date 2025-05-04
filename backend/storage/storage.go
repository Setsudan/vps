package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type StorageService struct {
	client *minio.Client
	bucket string
	url    string
}

func NewStorageService(client *minio.Client, bucket string) *StorageService {
	s3URL := os.Getenv("SEAWEEDFS_URL")
	if s3URL == "" {
		s3URL = "http://localhost:8333"
	}

	// Create a context with a timeout for the bucket operations.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Check if the bucket exists.
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		log.Fatalf("Error checking if bucket %q exists: %v", bucket, err)
	}

	// If it doesn't exist, create it.
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			log.Fatalf("Error creating bucket %q: %v", bucket, err)
		}
	}

	return &StorageService{
		client: client,
		bucket: bucket,
		url:    s3URL,
	}
}

// UploadFile uploads a file to S3 and returns its accessible URL.
func (s *StorageService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, path string) (string, error) {
	ext := filepath.Ext(header.Filename)
	fileID := uuid.New().String()
	objectName := fmt.Sprintf("%s/%s%s", path, fileID, ext)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	uploadCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	_, err := s.client.PutObject(uploadCtx, s.bucket, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	fileURL := fmt.Sprintf("%s/%s/%s", s.url, s.bucket, objectName)
	return fileURL, nil
}
