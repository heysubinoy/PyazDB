package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"

	"github.com/heysubinoy/pyazdb/api/proto"
	"github.com/heysubinoy/pyazdb/internal/api"
	"github.com/heysubinoy/pyazdb/internal/store"
	"github.com/heysubinoy/pyazdb/pkg/config"

	"google.golang.org/grpc"
)

/* ---------------- Discovery Types ---------------- */

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

/* ---------------- Raft Setup ---------------- */

func setupRaft(mem *store.MemStore, nodeID, bindAddr, dataDir string, bootstrap bool) *store.RaftStore {
	_ = os.MkdirAll(dataDir, 0700)

	cfg := raft.DefaultConfig()
	cfg.LocalID = raft.ServerID(nodeID)

	// Safer defaults
	cfg.HeartbeatTimeout = 2 * time.Second
	cfg.ElectionTimeout = 3 * time.Second
	cfg.LeaderLeaseTimeout = 1 * time.Second
	cfg.CommitTimeout = 500 * time.Millisecond

	logStore, _ := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-log.bolt"))
	stableStore, _ := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-stable.bolt"))
	snapshots, _ := raft.NewFileSnapshotStore(dataDir, 1, os.Stdout)

	transport, err := raft.NewTCPTransport(bindAddr, nil, 3, 10*time.Second, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	fsm := store.NewRaftStore(mem, nil)
	r, err := raft.NewRaft(cfg, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		log.Fatal(err)
	}

	rs := store.NewRaftStore(mem, r)

	hasState, _ := raft.HasExistingState(logStore, stableStore, snapshots)

	if bootstrap && !hasState {
		r.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				{ID: raft.ServerID(nodeID), Address: raft.ServerAddress(bindAddr)},
			},
		})
		log.Println("Cluster bootstrapped")
	}

	return rs
}

/* ---------------- Discovery Helpers ---------------- */

func registerLeader(mandi, nodeID, addr, httpAddr, grpcAddr string, r *raft.Raft) error {
	// Extract hostname from raft addr (e.g., "pyazdb-node1:12000" -> "pyazdb-node1")
	hostname := "localhost"
	if idx := len(addr) - 1; idx >= 0 {
		for i := 0; i < len(addr); i++ {
			if addr[i] == ':' {
				hostname = addr[:i]
				break
			}
		}
	}

	// Prepend hostname to HTTP and gRPC addresses if they start with ":"
	fullHTTPAddr := httpAddr
	if len(httpAddr) > 0 && httpAddr[0] == ':' {
		fullHTTPAddr = hostname + httpAddr
	}

	fullGRPCAddr := grpcAddr
	if len(grpcAddr) > 0 && grpcAddr[0] == ':' {
		fullGRPCAddr = hostname + grpcAddr
	}

	info := LeaderInfo{
		ID:        nodeID,
		Addr:      addr,
		HTTPAddr:  fullHTTPAddr,
		GRPCAddr:  fullGRPCAddr,
		Term:      r.CurrentTerm(),
		UpdatedAt: time.Now(),
	}

	data, _ := json.Marshal(info)
	req, _ := http.NewRequest(http.MethodPut, mandi+"/leader", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func postJoin(mandi, nodeID, addr string) {
	j := JoinRequest{ID: nodeID, Addr: addr}
	b, _ := json.Marshal(j)
	http.Post(mandi+"/join-requests", "application/json", bytes.NewBuffer(b))
}

/* ---------------- Leader Loop ---------------- */

// monitorLeadership continuously monitors if this node becomes leader
// and runs the leader duties when it does
func monitorLeadership(mandi, nodeID, raftAddr, httpAddr, grpcAddr string, r *raft.Raft) {
	for {
		// Wait until we become leader
		for r.State() != raft.Leader {
			time.Sleep(500 * time.Millisecond)
		}

		log.Println("Became leader, starting leader duties")

		// Run leader duties until we lose leadership
		runLeaderDuties(mandi, nodeID, raftAddr, httpAddr, grpcAddr, r)

		log.Println("Lost leadership, waiting for next election")
	}
}

func runLeaderDuties(mandi, nodeID, raftAddr, httpAddr, grpcAddr string, r *raft.Raft) {
	leaderTicker := time.NewTicker(2 * time.Second)
	joinTicker := time.NewTicker(3 * time.Second)

	defer leaderTicker.Stop()
	defer joinTicker.Stop()

	for {
		if r.State() != raft.Leader {
			return
		}

		select {
		case <-leaderTicker.C:
			_ = registerLeader(mandi, nodeID, raftAddr, httpAddr, grpcAddr, r)

		case <-joinTicker.C:
			resp, err := http.Get(mandi + "/join-requests")
			if err != nil {
				continue
			}

			var joins []JoinRequest
			_ = json.NewDecoder(resp.Body).Decode(&joins)
			resp.Body.Close()

			for _, j := range joins {
				if j.ID == nodeID {
					continue
				}

				log.Printf("Adding non-voter %s", j.ID)

				f := r.AddNonvoter(
					raft.ServerID(j.ID),
					raft.ServerAddress(j.Addr),
					0,
					10*time.Second,
				)
				if f.Error() != nil {
					continue
				}

				time.Sleep(2 * time.Second)

				p := r.AddVoter(
					raft.ServerID(j.ID),
					raft.ServerAddress(j.Addr),
					0,
					10*time.Second,
				)
				if err := p.Error(); err != nil {
					log.Printf("Failed to promote %s to voter: %v", j.ID, err)
					continue
				}

				req, _ := http.NewRequest(
					http.MethodDelete,
					mandi+"/join-requests?id="+j.ID,
					nil,
				)
				http.DefaultClient.Do(req)

				log.Printf("Node %s promoted to voter", j.ID)
			}
		}
	}
}

/* ---------------- Non-Leader Loop ---------------- */

func nonLeaderLoop(mandi, nodeID, raftAddr string, r *raft.Raft) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		postJoin(mandi, nodeID, raftAddr)

		cfg := r.GetConfiguration()
		if cfg.Error() == nil {
			for _, s := range cfg.Configuration().Servers {
				if string(s.ID) == nodeID {
					log.Println("Joined cluster successfully")
					return
				}
			}
		}

		<-ticker.C
	}
}

/* ---------------- Main ---------------- */

func main() {
	mem := store.NewMemStore()

	// NODE_CONFIG is now optional - if not set, will use environment variables
	cfgPath := os.Getenv("NODE_CONFIG")

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	rs := setupRaft(mem, cfg.NodeID, cfg.RaftAddr, cfg.RaftData, cfg.RaftLeader)

	var r *raft.Raft
	if g, ok := interface{}(rs).(interface{ GetRaft() *raft.Raft }); ok {
		r = g.GetRaft()
	}

	// Always monitor for leadership changes - any node can become leader
	go monitorLeadership(cfg.MandiAddr, cfg.NodeID, cfg.RaftAddr, cfg.HTTPAddr, cfg.GRPCAddr, r)

	// Non-leader nodes should try to join the cluster
	if !cfg.RaftLeader {
		go nonLeaderLoop(cfg.MandiAddr, cfg.NodeID, cfg.RaftAddr, r)
	}

	go func() {
		lis, _ := net.Listen("tcp", cfg.GRPCAddr)
		s := grpc.NewServer()
		proto.RegisterKVServiceServer(s, api.NewGRPCServer(rs, r, cfg.GRPCAddr, cfg.MandiAddr))
		s.Serve(lis)
	}()

	httpSrv := api.NewServer(rs, r, cfg.MandiAddr, cfg.HTTPAddr)
	mux := http.NewServeMux()
	httpSrv.RegisterRoutes(mux)

	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, mux))
}
