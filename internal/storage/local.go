package storage

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

type LocalStore struct {
	root string
}

func NewLocalStore(root string) (*LocalStore, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local store root: %w", err)
	}
	return &LocalStore{root: abs}, nil
}

func (l *LocalStore) GetObject(ctx context.Context, key string) ([]byte, error) {
	path := filepath.Join(l.root, filepath.FromSlash(key))
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", key, err)
	}
	return content, nil
}

func (l *LocalStore) GetObjectMeta(ctx context.Context, key string) (*ObjectMeta, error) {
	path := filepath.Join(l.root, filepath.FromSlash(key))
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", key, err)
	}
	return &ObjectMeta{
		Key:          key,
		LastModified: info.ModTime(),
		Size:         info.Size(),
	}, nil
}

func (l *LocalStore) ListObjects(ctx context.Context, prefix string) ([]ObjectMeta, error) {
	dir := filepath.Join(l.root, filepath.FromSlash(prefix))
	var objects []ObjectMeta
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(l.root, path)
		if err != nil {
			return err
		}
		objects = append(objects, ObjectMeta{
			Key:          filepath.ToSlash(rel),
			LastModified: info.ModTime(),
			Size:         info.Size(),
		})
		return nil
	})
	return objects, err
}

func (l *LocalStore) GetAsset(ctx context.Context, path string) ([]byte, string, error) {
	fullPath := filepath.Join(l.root, "posts", "assets", filepath.FromSlash(path))
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read asset %s: %w", path, err)
	}
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	return content, contentType, nil
}
