package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Store struct {
	client     *s3.Client
	bucketName string
}

func NewS3Store(ctx context.Context, bucketName string) (*S3Store, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &S3Store{
		client:     s3.NewFromConfig(cfg),
		bucketName: bucketName,
	}, nil
}

func (s *S3Store) GetObject(ctx context.Context, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object content: %w", err)
	}

	return content, nil
}

func (s *S3Store) GetObjectMeta(ctx context.Context, key string) (*ObjectMeta, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata from S3: %w", err)
	}

	return &ObjectMeta{
		Key:          key,
		LastModified: aws.ToTime(result.LastModified),
		Size:         aws.ToInt64(result.ContentLength),
	}, nil
}

func (s *S3Store) ListObjects(ctx context.Context, prefix string) ([]ObjectMeta, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	}

	result, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects from S3: %w", err)
	}

	var objects []ObjectMeta
	for _, obj := range result.Contents {
		objects = append(objects, ObjectMeta{
			Key:          aws.ToString(obj.Key),
			LastModified: aws.ToTime(obj.LastModified),
			Size:         aws.ToInt64(obj.Size),
		})
	}

	return objects, nil
}

func (s *S3Store) GetAsset(ctx context.Context, path string) ([]byte, string, error) {
	key := fmt.Sprintf("posts/assets/%s", path)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get asset from S3: %w", err)
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read asset content: %w", err)
	}

	contentType := "application/octet-stream"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return content, contentType, nil
}
