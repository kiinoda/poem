package errors

import "errors"

var (
	ErrPostNotFound     = errors.New("post not found")
	ErrAssetNotFound    = errors.New("asset not found")
	ErrStorageError     = errors.New("storage error")
	ErrInvalidFormat    = errors.New("invalid post format")
	ErrServiceNotReady  = errors.New("service not ready")
)
