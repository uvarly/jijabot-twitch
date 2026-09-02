package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FileStore struct {
	mu   sync.Mutex
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (fs *FileStore) Load(ctx context.Context) (Token, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := os.ReadFile(fs.path)
	if err != nil {
		return Token{}, fmt.Errorf("failed to read token file: %w", err)
	}

	var t Token
	if err := json.Unmarshal(data, &t); err != nil {
		return Token{}, fmt.Errorf("failed to unmarshal token file: %w", err)
	}

	return t, nil
}

func (fs *FileStore) Save(ctx context.Context, t Token) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	dir := filepath.Dir(fs.path)
	tmp, err := os.CreateTemp(dir, ".token-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp token file in %q: %w", dir, err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write temp token file %q: %w", tmpPath, err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp token file %q: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, fs.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp token file %q: %w", tmpPath, err)
	}

	return nil
}
