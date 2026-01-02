	package store

import (
	"sync/atomic"
	"time"

	"github.com/heysubinoy/pyazdb/pkg/kv"
)

// Metrics holds timing statistics for store operations.
// Uses atomic operations for thread-safe updates without locks.
type Metrics struct {
	GetCount    atomic.Uint64
	SetCount    atomic.Uint64
	DeleteCount atomic.Uint64
	
	// Cumulative latencies in nanoseconds
	GetLatencyNs    atomic.Uint64
	SetLatencyNs    atomic.Uint64
	DeleteLatencyNs atomic.Uint64
}

// InstrumentedStore wraps any kv.Store implementation with timing metrics.
// This pattern works for both in-memory and Raft-backed stores.
type InstrumentedStore struct {
	store   kv.Store
	metrics *Metrics
}

// Compile-time check to ensure InstrumentedStore implements kv.Store.
var _ kv.Store = (*InstrumentedStore)(nil)

// NewInstrumentedStore wraps a store with instrumentation.
func NewInstrumentedStore(store kv.Store) *InstrumentedStore {
	return &InstrumentedStore{
		store:   store,
		metrics: &Metrics{},
	}
}

// Get delegates to the wrapped store and records timing.
func (s *InstrumentedStore) Get(key string) (string, bool) {
	start := time.Now()
	value, found := s.store.Get(key)
	elapsed := time.Since(start).Nanoseconds()
	
	s.metrics.GetCount.Add(1)
	s.metrics.GetLatencyNs.Add(uint64(elapsed))
	
	return value, found
}

// Set delegates to the wrapped store and records timing.
func (s *InstrumentedStore) Set(key, value string) error {
	start := time.Now()
	err := s.store.Set(key, value)
	elapsed := time.Since(start).Nanoseconds()
	
	s.metrics.SetCount.Add(1)
	s.metrics.SetLatencyNs.Add(uint64(elapsed))
	
	return err
}

// Delete delegates to the wrapped store and records timing.
func (s *InstrumentedStore) Delete(key string) error {
	start := time.Now()
	err := s.store.Delete(key)
	elapsed := time.Since(start).Nanoseconds()
	
	s.metrics.DeleteCount.Add(1)
	s.metrics.DeleteLatencyNs.Add(uint64(elapsed))
	
	return err
}

// GetMetrics returns a snapshot of current metrics.
func (s *InstrumentedStore) GetMetrics() MetricsSnapshot {
	getCount := s.metrics.GetCount.Load()
	setCount := s.metrics.SetCount.Load()
	deleteCount := s.metrics.DeleteCount.Load()
	
	return MetricsSnapshot{
		GetCount:       getCount,
		SetCount:       setCount,
		DeleteCount:    deleteCount,
		GetAvgLatency:  s.avgLatency(s.metrics.GetLatencyNs.Load(), getCount),
		SetAvgLatency:  s.avgLatency(s.metrics.SetLatencyNs.Load(), setCount),
		DeleteAvgLatency: s.avgLatency(s.metrics.DeleteLatencyNs.Load(), deleteCount),
	}
}

// ResetMetrics clears all metrics counters.
func (s *InstrumentedStore) ResetMetrics() {
	s.metrics.GetCount.Store(0)
	s.metrics.SetCount.Store(0)
	s.metrics.DeleteCount.Store(0)
	s.metrics.GetLatencyNs.Store(0)
	s.metrics.SetLatencyNs.Store(0)
	s.metrics.DeleteLatencyNs.Store(0)
}

func (s *InstrumentedStore) avgLatency(totalNs, count uint64) time.Duration {
	if count == 0 {
		return 0
	}
	return time.Duration(totalNs / count)
}

// MetricsSnapshot is a point-in-time view of metrics.
type MetricsSnapshot struct {
	GetCount         uint64
	SetCount         uint64
	DeleteCount      uint64
	GetAvgLatency    time.Duration
	SetAvgLatency    time.Duration
	DeleteAvgLatency time.Duration
}
