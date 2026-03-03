package storage

import (
	"context"
	"io"
)

// FileStorage provides file upload and deletion capabilities.
type FileStorage interface {
	Upload(ctx context.Context, filename string, data io.Reader, contentType string) (url string, err error)
	Delete(ctx context.Context, filename string) error
}
