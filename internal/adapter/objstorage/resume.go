package objstorage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"hr-helper/internal/inerrors"
)

type ResumeStorage struct {
	client *minio.Client
	bucket string
}

func NewResumeStorage(bucket string, client *minio.Client) *ResumeStorage {
	return &ResumeStorage{
		client: client,
		bucket: bucket,
	}
}

func (s *ResumeStorage) Download(ctx context.Context, candidateID int64, vacancyID uuid.UUID) ([]byte, error) {
	key := fmt.Sprintf("%d/%s", candidateID, vacancyID)

	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		var minioErr minio.ErrorResponse
		if errors.As(err, &minioErr) {
			if minioErr.Code == minio.NoSuchKey {
				return nil, inerrors.ErrNotFound
			}
		}
		return nil, fmt.Errorf("can't download object: %w", err)
	}
	defer object.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, object)
	if err != nil {
		return nil, fmt.Errorf("can't copy bytes: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *ResumeStorage) GetPresignedURL(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (string, error) {
	key := fmt.Sprintf("%d/%s", candidateID, vacancyID)

	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, key, time.Minute*20, nil)
	if err != nil {
		return "", fmt.Errorf("can't presign object: %w", err)
	}

	return presignedURL.String(), nil
}
