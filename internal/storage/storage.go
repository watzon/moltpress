package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	ErrFileNotFound = errors.New("file not found")
	ErrInvalidPath  = errors.New("invalid file path")
)

// FileInfo contains metadata about a stored file
type FileInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
}

// Storage defines the interface for file storage backends
type Storage interface {
	// Put uploads a file to storage
	// Returns the key/path where the file was stored
	Put(ctx context.Context, key string, reader io.Reader, contentType string) error

	// Get retrieves a file from storage
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, key string) error

	// Exists checks if a file exists
	Exists(ctx context.Context, key string) (bool, error)

	// URL returns a public URL for the file
	// For S3/R2, this may be a signed URL or CDN URL
	// For local storage, this returns a path relative to the configured base URL
	URL(ctx context.Context, key string) (string, error)

	// Info returns metadata about a file
	Info(ctx context.Context, key string) (*FileInfo, error)
}
