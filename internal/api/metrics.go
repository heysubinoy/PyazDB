package api

import (
	"encoding/json"
	"net/http"

	"github.com/heysubinoy/pyazdb/internal/store"
)

// MetricsHandler returns current store metrics as JSON.
// Only works if the server was initialized with an InstrumentedStore.
func MetricsHandler(instrumentedStore *store.InstrumentedStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		metrics := instrumentedStore.GetMetrics()

		response := map[string]interface{}{
			"operations": map[string]uint64{
				"get":    metrics.GetCount,
				"set":    metrics.SetCount,
				"delete": metrics.DeleteCount,
			},
			"avg_latency": map[string]string{
				"get":    metrics.GetAvgLatency.String(),
				"set":    metrics.SetAvgLatency.String(),
				"delete": metrics.DeleteAvgLatency.String(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
