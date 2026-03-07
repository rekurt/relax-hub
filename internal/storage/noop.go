package storage

import (
	"context"
	"io"

	"github.com/nikitaaldaev/bani/internal/logger"
)

// NoopStorage is a FileStorage that does nothing. Used in dev when S3 is unavailable.
type NoopStorage struct {
	logger *logger.Logger
}

func NewNoopStorage(log *logger.Logger) *NoopStorage {
	return &NoopStorage{logger: log}
}

func (n *NoopStorage) Upload(_ context.Context, filename string, _ io.Reader, _ string) (string, error) {
	n.logger.Warn("noop storage: upload skipped", "filename", filename)
	return "noop://" + filename, nil
}

func (n *NoopStorage) Delete(_ context.Context, filename string) error {
	n.logger.Warn("noop storage: delete skipped", "filename", filename)
	return nil
}
