package kv

// Store defines the interface for a key-value store.
// Implementations of this interface can be swapped out,
// allowing for different storage backends (e.g., in-memory, Raft-replicated).
type Store interface {
	// Get retrieves the value associated with the given key.
	// Returns the value and true if the key exists, or empty string and false if not.
	Get(key string) (string, bool)

	// Set stores a key-value pair.
	// Returns an error if the operation fails.
	Set(key, value string) error

	// Delete removes a key from the store.
	// Returns an error if the operation fails.
	Delete(key string) error
}
