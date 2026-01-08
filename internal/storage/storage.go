package storage

import (
	"context"
	"time"
)

type ObjectMeta struct {
	Key          string
	LastModified time.Time
	Size         int64
}

type ContentStore interface {
	GetObject(ctx context.Context, key string) ([]byte, error)
	GetObjectMeta(ctx context.Context, key string) (*ObjectMeta, error)
	ListObjects(ctx context.Context, prefix string) ([]ObjectMeta, error)
	GetAsset(ctx context.Context, path string) (content []byte, contentType string, err error)
}
