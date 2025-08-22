package session

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Session represents session state.
type Session struct {
	Token    string `json:"token"`
	Template string `json:"template"`
}

// Store handles persistence in state directory.
type Store struct {
	path string
}

// NewStore creates store with default path.
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".local", "state", "seidor-cloud")
	os.MkdirAll(dir, 0o755)
	return &Store{path: filepath.Join(dir, "session.json")}, nil
}

// Load reads session if exists.
func (s *Store) Load() (*Session, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Session{}, nil
		}
		return nil, err
	}
	sess := &Session{}
	if err := json.Unmarshal(b, sess); err != nil {
		return nil, err
	}
	return sess, nil
}

// Save writes session to disk.
func (s *Store) Save(sess *Session) error {
	b, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o600)
}
