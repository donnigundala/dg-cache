package dgcache

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	instrumentationName = "github.com/donnigundala/dg-cache"
)

// RegisterMetrics registers cache metrics with OpenTelemetry.
func (m *Manager) RegisterMetrics() error {
	meter := otel.GetMeterProvider().Meter(instrumentationName)

	var err error

	// Cumulative counters for operations
	m.metricHits, err = meter.Int64ObservableCounter(
		"cache.hits",
		metric.WithDescription("Total number of cache hits"),
	)
	if err != nil {
		return err
	}

	m.metricMisses, err = meter.Int64ObservableCounter(
		"cache.misses",
		metric.WithDescription("Total number of cache misses"),
	)
	if err != nil {
		return err
	}

	m.metricSets, err = meter.Int64ObservableCounter(
		"cache.sets",
		metric.WithDescription("Total number of cache set operations"),
	)
	if err != nil {
		return err
	}

	m.metricDeletes, err = meter.Int64ObservableCounter(
		"cache.deletes",
		metric.WithDescription("Total number of cache delete operations"),
	)
	if err != nil {
		return err
	}

	m.metricEvictions, err = meter.Int64ObservableCounter(
		"cache.evictions",
		metric.WithDescription("Total number of cache evictions"),
	)
	if err != nil {
		return err
	}

	// Gauges for current state
	m.metricItems, err = meter.Int64ObservableGauge(
		"cache.items",
		metric.WithDescription("Current number of items in cache"),
	)
	if err != nil {
		return err
	}

	m.metricBytes, err = meter.Int64ObservableGauge(
		"cache.bytes",
		metric.WithDescription("Current bytes used by cache"),
	)
	if err != nil {
		return err
	}

	// Register callback to collect metrics from all stores
	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		m.mu.RLock()
		defer m.mu.RUnlock()

		for name, store := range m.stores {
			stats := store.Stats()
			attrs := metric.WithAttributes(
				attribute.String("cache.store", name),
			)

			o.ObserveInt64(m.metricHits, stats.Hits, attrs)
			o.ObserveInt64(m.metricMisses, stats.Misses, attrs)
			o.ObserveInt64(m.metricSets, stats.Sets, attrs)
			o.ObserveInt64(m.metricDeletes, stats.Deletes, attrs)
			o.ObserveInt64(m.metricEvictions, stats.Evictions, attrs)
			o.ObserveInt64(m.metricItems, int64(stats.ItemCount), attrs)
			o.ObserveInt64(m.metricBytes, stats.BytesUsed, attrs)
		}
		return nil
	}, m.metricHits, m.metricMisses, m.metricSets, m.metricDeletes, m.metricEvictions, m.metricItems, m.metricBytes)

	return err
}
