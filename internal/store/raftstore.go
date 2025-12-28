package store

import (
	"encoding/json"
	"io"

	"github.com/hashicorp/raft"
)

// GetRaft returns the underlying raft.Raft pointer (for API layer leader checks)
func (rs *RaftStore) GetRaft() *raft.Raft {
	return rs.raft
}

// RaftCommand represents a set/delete operation to be applied via Raft.
type RaftCommand struct {
	Op    string // "set" or "delete"
	Key   string
	Value string // only for set
}

// RaftStore wraps a Store and applies changes via Raft consensus.
type RaftStore struct {
	store *MemStore
	raft  *raft.Raft
}

func NewRaftStore(store *MemStore, r *raft.Raft) *RaftStore {
	return &RaftStore{store: store, raft: r}
}

// Apply applies a Raft log entry to the local store.
func (rs *RaftStore) Apply(log *raft.Log) interface{} {
	var cmd RaftCommand
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return err
	}
	switch cmd.Op {
	case "set":
		rs.store.Set(cmd.Key, cmd.Value)
	case "delete":
		rs.store.Delete(cmd.Key)
	}
	return nil
}

// Snapshot and Restore are required for raft.FSM, but can be no-ops for in-memory.
func (rs *RaftStore) Snapshot() (raft.FSMSnapshot, error) { return &noopSnapshot{}, nil }
func (rs *RaftStore) Restore(io.ReadCloser) error         { return nil }

type noopSnapshot struct{}

func (n *noopSnapshot) Persist(sink raft.SnapshotSink) error { return sink.Close() }
func (n *noopSnapshot) Release()                             {}

// Set submits a set command to Raft.
func (rs *RaftStore) Set(key, value string) error {
	cmd := RaftCommand{Op: "set", Key: key, Value: value}
	data, _ := json.Marshal(cmd)
	f := rs.raft.Apply(data, 0)
	return f.Error()
}

// Delete submits a delete command to Raft.
func (rs *RaftStore) Delete(key string) error {
	cmd := RaftCommand{Op: "delete", Key: key}
	data, _ := json.Marshal(cmd)
	f := rs.raft.Apply(data, 0)
	return f.Error()
}

// Get reads directly from the local store.
func (rs *RaftStore) Get(key string) (string, bool) {
	return rs.store.Get(key)
}
