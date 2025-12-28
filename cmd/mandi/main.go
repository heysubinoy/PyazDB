package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

/*
mandi is a soft-state discovery service for Raft joins.
It is NOT authoritative and NOT part of Raft correctness.
*/

// -------------------- Config --------------------

const (
	leaderTTL      = 10 * time.Second
	joinRequestTTL = 30 * time.Second
	cleanupEvery   = 5 * time.Second
)

// -------------------- Types --------------------

type LeaderInfo struct {
	ID        string    `json:"id"`
	Addr      string    `json:"addr"`
	HTTPAddr  string    `json:"http_addr"`
	GRPCAddr  string    `json:"grpc_addr"`
	Term      uint64    `json:"term"`
	UpdatedAt time.Time `json:"updated_at"`
}

type JoinRequest struct {
	ID        string    `json:"id"`
	Addr      string    `json:"addr"`
	StartedAt time.Time `json:"started_at"`
}

// -------------------- In-memory Store --------------------

type Store struct {
	mu           sync.Mutex
	leader       *LeaderInfo
	joinRequests map[string]JoinRequest
}

func NewStore() *Store {
	return &Store{
		joinRequests: make(map[string]JoinRequest),
	}
}

// -------------------- HTTP Handlers --------------------

func (s *Store) getLeader(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.leader == nil || time.Since(s.leader.UpdatedAt) > leaderTTL {
		http.Error(w, "leader not available", http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(s.leader)
}

func (s *Store) putLeader(w http.ResponseWriter, r *http.Request) {
	var info LeaderInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	info.UpdatedAt = time.Now()

	s.mu.Lock()
	s.leader = &info
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (s *Store) postJoinRequest(w http.ResponseWriter, r *http.Request) {
	var jr JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&jr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jr.StartedAt = time.Now()

	s.mu.Lock()
	s.joinRequests[jr.ID] = jr
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (s *Store) listJoinRequests(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var list []JoinRequest
	for _, jr := range s.joinRequests {
		list = append(list, jr)
	}

	_ = json.NewEncoder(w).Encode(list)
}

func (s *Store) deleteJoinRequest(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	delete(s.joinRequests, id)
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// -------------------- Cleanup Loop --------------------

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(cleanupEvery)
	for range ticker.C {
		s.mu.Lock()

		// Expire leader
		if s.leader != nil && time.Since(s.leader.UpdatedAt) > leaderTTL {
			s.leader = nil
		}

		// Expire join requests
		for id, jr := range s.joinRequests {
			if time.Since(jr.StartedAt) > joinRequestTTL {
				delete(s.joinRequests, id)
			}
		}

		s.mu.Unlock()
	}
}

// -------------------- main --------------------

func main() {
	addr := ":7000"
	if v := os.Getenv("MANDI_ADDR"); v != "" {
		addr = v
	}

	store := NewStore()
	go store.cleanupLoop()

	mux := http.NewServeMux()

	mux.HandleFunc("/leader", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			store.getLeader(w, r)
		case http.MethodPut:
			store.putLeader(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/join-requests", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			store.postJoinRequest(w, r)
		case http.MethodGet:
			store.listJoinRequests(w, r)
		case http.MethodDelete:
			store.deleteJoinRequest(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("mandi listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
