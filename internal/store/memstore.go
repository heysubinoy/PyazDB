package store

import (
	"sync"

	"github.com/heysubinoy/pyazdb/pkg/kv"
)

// MemStore is an in-memory implementation of the kv.Store interface.
// It uses a map protected by a RWMutex for thread-safe operations.
type MemStore struct {
	mu   sync.RWMutex
	data map[string]string
}

// Compile-time check to ensure MemStore implements kv.Store.
var _ kv.Store = (*MemStore)(nil)

// NewMemStore creates and returns a new MemStore instance.
func NewMemStore() *MemStore {
	return &MemStore{
		data: make(map[string]string),
	}
}

// Get retrieves a value by key from the store.
// Returns the value and true if found, empty string and false otherwise.
func (s *MemStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[key]
	return val, ok
}

// Set stores a key-value pair in the store.
// Always returns nil for in-memory operations.
func (s *MemStore) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
	return nil
}

// Delete removes a key from the store.
// Always returns nil, even if the key doesn't exist.
func (s *MemStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	return nil
}
