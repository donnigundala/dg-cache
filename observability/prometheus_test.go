package observability

import (
	"strings"
	"testing"

	cache "github.com/donnigundala/dg-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

// MockObservable implements cache.Observable for testing.
type MockObservable struct {
	stats cache.Stats
}

func (m *MockObservable) Stats() cache.Stats {
	return m.stats
}

func TestPrometheusCollector(t *testing.T) {
	mock := &MockObservable{
		stats: cache.Stats{
			Hits:      10,
			Misses:    5,
			Sets:      20,
			Deletes:   2,
			Evictions: 1,
			ItemCount: 100,
			BytesUsed: 1024,
		},
	}

	collector := NewPrometheusCollector(mock, "myapp", "cache")

	// Create a registry and register the collector
	reg := prometheus.NewPedanticRegistry()
	err := reg.Register(collector)
	assert.NoError(t, err)

	// Gather metrics
	metrics, err := reg.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)

	// Verify metrics output using testutil (helper to lint and match)
	// We'll check for one metric text representation
	expected := `
		# HELP myapp_cache_hits_total Total number of cache hits
		# TYPE myapp_cache_hits_total counter
		myapp_cache_hits_total{driver="default"} 10
	`
	// Clean up formatting for comparison
	err = testutil.CollectAndCompare(collector, strings.NewReader(expected), "myapp_cache_hits_total")
	assert.NoError(t, err)
}
