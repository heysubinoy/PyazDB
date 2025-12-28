package api

import (
	"encoding/json"
	"net/http"

	"github.com/hashicorp/raft"
	"github.com/heysubinoy/pyazdb/pkg/kv"
)

// Server wraps a kv.Store and exposes HTTP endpoints for KV operations.
// The Store can be replaced with a Raft-backed implementation later.
type Server struct {
	Store kv.Store
	Raft  *raft.Raft
}

// NewServer creates a new HTTP server with the given store.
func NewServer(store kv.Store, raftNode *raft.Raft) *Server {
	return &Server{
		Store: store,
		Raft:  raftNode,
	}
}

// RegisterRoutes registers all HTTP handlers on the given mux.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/get", s.handleGet)
	mux.HandleFunc("/set", s.handleSet)
	mux.HandleFunc("/delete", s.handleDelete)
}

// handleGet handles GET /get?key=foo requests.
// Returns the value as plain text or appropriate error codes.
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Raft != nil && s.Raft.State() != raft.Leader {
		leader := s.Raft.Leader()
		if leader == "" {
			http.Error(w, "Not leader and no leader known", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Location", "http://"+string(leader)+":8080/get")
		http.Error(w, "Not leader. Redirect to leader.", http.StatusTemporaryRedirect)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	value, ok := s.Store.Get(key)
	if !ok {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(value))
}

// handleSet handles POST /set requests with JSON body.
// Expects: {"key": "foo", "value": "bar"}
func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Raft != nil && s.Raft.State() != raft.Leader {
		leader := s.Raft.Leader()
		if leader == "" {
			http.Error(w, "Not leader and no leader known", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Location", "http://"+string(leader)+":8080/set")
		http.Error(w, "Not leader. Redirect to leader.", http.StatusTemporaryRedirect)
		return
	}

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "Missing key field", http.StatusBadRequest)
		return
	}

	if err := s.Store.Set(req.Key, req.Value); err != nil {
		http.Error(w, "Failed to set key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleDelete handles POST /delete requests with JSON body.
// Expects: {"key": "foo"}
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Raft != nil && s.Raft.State() != raft.Leader {
		leader := s.Raft.Leader()
		if leader == "" {
			http.Error(w, "Not leader and no leader known", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Location", "http://"+string(leader)+":8080/delete")
		http.Error(w, "Not leader. Redirect to leader.", http.StatusTemporaryRedirect)
		return
	}

	var req struct {
		Key string `json:"key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "Missing key field", http.StatusBadRequest)
		return
	}

	if err := s.Store.Delete(req.Key); err != nil {
		http.Error(w, "Failed to delete key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
