# PyazDB

**PyazDB** is a distributed key-value database built in Go, featuring Raft consensus for data replication and fault tolerance. It provides both HTTP and gRPC APIs for interacting with the database.

## 🌟 Features

- **Distributed Architecture**: Multi-node cluster with automatic leader election
- **Raft Consensus**: Strong consistency using HashiCorp Raft implementation
- **Multiple APIs**: Both HTTP REST and gRPC interfaces
- **Service Discovery**: Built-in discovery service (Mandi) for cluster coordination
- **Automatic Failover**: Leader forwarding ensures requests reach the correct node
- **Persistent Storage**: BoltDB-backed Raft log and stable storage

## 🏗️ Architecture

PyazDB consists of the following services:

```
┌─────────────────────────────────────────────────────────────┐
│                        Clients                               │
│                  (HTTP / gRPC / CLI)                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Mandi (Discovery)                         │
│              Tracks leader & join requests                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     KV Nodes (Raft Cluster)                  │
│   ┌─────────┐    ┌─────────┐    ┌─────────┐                 │
│   │  Node1  │◄──►│  Node2  │◄──►│  Node3  │                 │
│   │(Leader) │    │(Follower)│    │(Follower)│                │
│   └─────────┘    └─────────┘    └─────────┘                 │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Services

### KV-Single (Database Node)

The core database node that stores key-value pairs using Raft consensus.

**Features:**
- In-memory key-value store with Raft replication
- HTTP API for REST operations
- gRPC API for high-performance access
- Automatic leader election and failover
- Request forwarding to leader when needed

**Endpoints:**
- HTTP: `:8080` (configurable)
- gRPC: `:9090` (configurable)
- Raft: `:12000` (configurable)

**Environment Variables:**
| Variable | Description | Default |
|----------|-------------|---------|
| `NODE_ID` | Unique identifier for the node | Required |
| `RAFT_ADDR` | Address for Raft communication | Required |
| `RAFT_DATA` | Directory for Raft data persistence | Required |
| `RAFT_LEADER` | Bootstrap as leader (first node only) | `false` |
| `GRPC_ADDR` | gRPC server address | `:9090` |
| `HTTP_ADDR` | HTTP server address | `:8080` |
| `MANDI_ADDR` | Mandi discovery service address | `http://127.0.0.1:7000` |

### Mandi (Discovery Service)

A lightweight discovery service that helps nodes find the current leader and coordinate cluster joins. It maintains soft-state and is **not** part of Raft correctness.

**Features:**
- Leader tracking with TTL expiration
- Join request management
- Automatic cleanup of stale entries

**Endpoints:**
- `GET /leader` - Get current leader information
- `PUT /leader` - Register/update leader (called by leader node)
- `POST /join-requests` - Submit a join request (called by new nodes)
- `GET /join-requests` - List pending join requests
- `DELETE /join-requests?id=<node_id>` - Remove a join request

**Environment Variables:**
| Variable | Description | Default |
|----------|-------------|---------|
| `MANDI_ADDR` | Listen address | `:7000` |

### KV-CLI (Command Line Interface)

A command-line tool for interacting with the PyazDB cluster.

**Usage:**
```bash
# Set a value
kv-cli set <key> <value>

# Get a value
kv-cli get <key>

# Delete a value
kv-cli delete <key>
```

**Environment Variables:**
| Variable | Description | Default |
|----------|-------------|---------|
| `MANDI_ADDR` | Mandi discovery service address | `http://127.0.0.1:7000` |

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose (for containerized deployment)

### Building

```bash
# Build all binaries
go build -o bin/kv-single ./cmd/kv-single
go build -o bin/mandi ./cmd/mandi
go build -o bin/kv-cli ./cmd/kv-cli
```

### Running with Docker Compose

The easiest way to run PyazDB is using Docker Compose:

```bash
# Build and start the cluster
docker-compose up --build

# This starts:
# - mandi (discovery service) on port 7000
# - node1 (bootstrap leader) on ports 8080 (HTTP), 9090 (gRPC), 12000 (Raft)
```

### Running Manually

**1. Start the Mandi discovery service:**
```bash
MANDI_ADDR=:7000 ./bin/mandi
```

**2. Start the first node (bootstrap leader):**
```bash
NODE_ID=node1 \
RAFT_ADDR=127.0.0.1:12000 \
RAFT_DATA=./data/node1 \
RAFT_LEADER=true \
GRPC_ADDR=:9090 \
HTTP_ADDR=:8080 \
MANDI_ADDR=http://127.0.0.1:7000 \
./bin/kv-single
```

**3. Start additional nodes:**
```bash
NODE_ID=node2 \
RAFT_ADDR=127.0.0.1:12001 \
RAFT_DATA=./data/node2 \
RAFT_LEADER=false \
GRPC_ADDR=:9091 \
HTTP_ADDR=:8081 \
MANDI_ADDR=http://127.0.0.1:7000 \
./bin/kv-single
```

## 📡 API Reference

### HTTP API

**Get a value:**
```bash
curl "http://localhost:8080/get?key=mykey"
```

**Set a value:**
```bash
curl -X POST "http://localhost:8080/set" \
  -H "Content-Type: application/json" \
  -d '{"key": "mykey", "value": "myvalue"}'
```

**Delete a value:**
```bash
curl -X DELETE "http://localhost:8080/delete?key=mykey"
```

### gRPC API

The gRPC service is defined in `api/proto/kv.proto`:

```protobuf
service KVService {
  rpc Get(GetRequest) returns (GetResponse);
  rpc Set(SetRequest) returns (SetResponse);
  rpc Delete(DeleteRequest) returns (DeleteResponse);
}
```

**Using the CLI:**
```bash
# Set environment variable for discovery
export MANDI_ADDR=http://127.0.0.1:7000

# Operations
./bin/kv-cli set hello world
./bin/kv-cli get hello
./bin/kv-cli delete hello
```

## 📁 Project Structure

```
PyazDB/
├── api/
│   └── proto/           # gRPC protocol definitions
├── cmd/
│   ├── kv-cli/          # Command-line client
│   ├── kv-single/       # Database node
│   └── mandi/           # Discovery service
├── internal/
│   ├── api/             # HTTP and gRPC server implementations
│   └── store/           # Storage implementations (MemStore, RaftStore)
├── pkg/
│   ├── config/          # Configuration loading
│   └── kv/              # Store interface definition
├── docker-compose.yml   # Docker Compose configuration
├── Dockerfile.kv-single # Dockerfile for database nodes
└── Dockerfile.mandi     # Dockerfile for discovery service
```

## 🔄 How Raft Consensus Works in PyazDB

1. **Leader Election**: When the cluster starts, nodes elect a leader using Raft
2. **Write Path**: All writes go through the leader, which replicates to followers
3. **Read Path**: Reads can be served by any node (eventual consistency) or forwarded to leader
4. **Failover**: If the leader fails, remaining nodes elect a new leader automatically
5. **Join Process**: New nodes register with Mandi, leader adds them as non-voters, then promotes to voters

