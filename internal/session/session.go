package session

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Session holds runtime session data.
type Session struct {
	Token         string `json:"token"`
	ActiveProfile string `json:"active_profile"`
}

// Store defines session persistence.
type Store interface {
	Load() (Session, error)
	Save(Session) error
	Clear() error
}

// FileStore persists session to disk.
type FileStore struct {
	path string
}

// NewFileStore creates a new FileStore using XDG state directory.
func NewFileStore() (*FileStore, error) {
	dir := filepath.Join(os.Getenv("HOME"), ".local", "state", "seidor-aws-cli")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &FileStore{path: filepath.Join(dir, "session.json")}, nil
}

// Load returns session from disk.
func (s *FileStore) Load() (Session, error) {
	var sess Session
	b, err := os.ReadFile(s.path)
	if err != nil {
		return sess, err
	}
	if err := json.Unmarshal(b, &sess); err != nil {
		return sess, err
	}
	return sess, nil
}

// Save writes session.
func (s *FileStore) Save(sess Session) error {
	b, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o600)
}

// Clear removes session file.
func (s *FileStore) Clear() error {
	return os.Remove(s.path)
}
