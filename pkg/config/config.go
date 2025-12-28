package config

import (
	"fmt"
	"os"

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

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}
