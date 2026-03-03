package storage

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// MockStorage is an in-memory FileStorage for testing.
type MockStorage struct {
	mu    sync.RWMutex
	files map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		files: make(map[string][]byte),
	}
}

func (m *MockStorage) Upload(_ context.Context, filename string, data io.Reader, _ string) (string, error) {
	buf, err := io.ReadAll(data)
	if err != nil {
		return "", fmt.Errorf("read data: %w", err)
	}

	m.mu.Lock()
	m.files[filename] = buf
	m.mu.Unlock()

	return "http://mock-storage/" + filename, nil
}

func (m *MockStorage) Delete(_ context.Context, filename string) error {
	m.mu.Lock()
	delete(m.files, filename)
	m.mu.Unlock()
	return nil
}

// Has returns true if the file exists in mock storage.
func (m *MockStorage) Has(filename string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.files[filename]
	return ok
}

// Get returns the stored file data.
func (m *MockStorage) Get(filename string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.files[filename]
	return data, ok
}

// Len returns the number of stored files.
func (m *MockStorage) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.files)
}
