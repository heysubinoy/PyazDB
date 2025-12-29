package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NodeID     string `yaml:"node_id"`
	RaftAddr   string `yaml:"raft_addr"`
	RaftData   string `yaml:"raft_data"`
	RaftLeader bool   `yaml:"raft_leader"`
	GRPCAddr   string `yaml:"grpc_addr"`
	HTTPAddr   string `yaml:"http_addr"`
	MandiAddr  string `yaml:"mandi_addr"`
}

// LoadConfig loads configuration from a YAML file if path is provided,
// otherwise it falls back to environment variables.
func LoadConfig(path string) (*Config, error) {
	var cfg Config

	// If path is provided and file exists, load from YAML
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
			// Apply environment variable overrides
			applyEnvOverrides(&cfg)
			return &cfg, nil
		}
		// If path was explicitly provided but file doesn't exist, return error
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Load from environment variables
	cfg.NodeID = os.Getenv("NODE_ID")
	cfg.RaftAddr = os.Getenv("RAFT_ADDR")
	cfg.RaftData = os.Getenv("RAFT_DATA")
	cfg.GRPCAddr = os.Getenv("GRPC_ADDR")
	cfg.HTTPAddr = os.Getenv("HTTP_ADDR")
	cfg.MandiAddr = os.Getenv("MANDI_ADDR")

	// Parse RAFT_LEADER as boolean
	if leaderStr := os.Getenv("RAFT_LEADER"); leaderStr != "" {
		leader, err := strconv.ParseBool(leaderStr)
		if err != nil {
			return nil, fmt.Errorf("invalid RAFT_LEADER value: %w", err)
		}
		cfg.RaftLeader = leader
	}

	// Set defaults if not provided
	if cfg.RaftData == "" {
		cfg.RaftData = fmt.Sprintf("./pyaz/%s", cfg.NodeID)
	}
	if cfg.MandiAddr == "" {
		cfg.MandiAddr = "http://127.0.0.1:7000"
	}

	// Validate required fields
	if cfg.NodeID == "" {
		return nil, fmt.Errorf("NODE_ID is required (set via environment or config file)")
	}
	if cfg.RaftAddr == "" {
		return nil, fmt.Errorf("RAFT_ADDR is required (set via environment or config file)")
	}
	if cfg.GRPCAddr == "" {
		return nil, fmt.Errorf("GRPC_ADDR is required (set via environment or config file)")
	}
	if cfg.HTTPAddr == "" {
		return nil, fmt.Errorf("HTTP_ADDR is required (set via environment or config file)")
	}

	return &cfg, nil
}

// applyEnvOverrides allows environment variables to override YAML config values
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("NODE_ID"); v != "" {
		cfg.NodeID = v
	}
	if v := os.Getenv("RAFT_ADDR"); v != "" {
		cfg.RaftAddr = v
	}
	if v := os.Getenv("RAFT_DATA"); v != "" {
		cfg.RaftData = v
	}
	if v := os.Getenv("GRPC_ADDR"); v != "" {
		cfg.GRPCAddr = v
	}
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := os.Getenv("MANDI_ADDR"); v != "" {
		cfg.MandiAddr = v
	}
	if v := os.Getenv("RAFT_LEADER"); v != "" {
		if leader, err := strconv.ParseBool(v); err == nil {
			cfg.RaftLeader = leader
		}
	}
}
