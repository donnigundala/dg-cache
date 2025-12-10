package observability

import (
	"github.com/prometheus/client_golang/prometheus"

	cache "github.com/donnigundala/dg-cache"
)

// PrometheusCollector exports cache metrics to Prometheus.
type PrometheusCollector struct {
	driver    cache.Observable
	hits      *prometheus.Desc
	misses    *prometheus.Desc
	sets      *prometheus.Desc
	deletes   *prometheus.Desc
	evictions *prometheus.Desc
	items     *prometheus.Desc
	bytes     *prometheus.Desc
}

// NewPrometheusCollector creates a new PrometheusCollector.
// Namespace and subsystem are optional but recommended (e.g. "myapp", "cache").
func NewPrometheusCollector(driver cache.Observable, namespace, subsystem string) *PrometheusCollector {
	labels := []string{"driver"} // Could assume driver name if exposed, but kept simple

	return &PrometheusCollector{
		driver: driver,
		hits: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "hits_total"),
			"Total number of cache hits",
			labels, nil,
		),
		misses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "misses_total"),
			"Total number of cache misses",
			labels, nil,
		),
		sets: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "sets_total"),
			"Total number of cache set operations",
			labels, nil,
		),
		deletes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "deletes_total"),
			"Total number of cache delete operations",
			labels, nil,
		),
		evictions: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "evictions_total"),
			"Total number of cache evictions",
			labels, nil,
		),
		items: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "items"),
			"Current number of items in cache",
			labels, nil,
		),
		bytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "bytes"),
			"Current bytes used by cache",
			labels, nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.hits
	ch <- c.misses
	ch <- c.sets
	ch <- c.deletes
	ch <- c.evictions
	ch <- c.items
	ch <- c.bytes
}

// Collect implements prometheus.Collector.
func (c *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.driver.Stats()

	// We use "unknown" or pass driver name if meaningful.
	// For now, let's just use a constant or allow passing it in constructor?
	// Let's pass "cache" as label value or remove the label if not needed.
	// Actually, let's just iterate and use constant label value for simplicity
	labelValues := []string{"default"}

	ch <- prometheus.MustNewConstMetric(c.hits, prometheus.CounterValue, float64(stats.Hits), labelValues...)
	ch <- prometheus.MustNewConstMetric(c.misses, prometheus.CounterValue, float64(stats.Misses), labelValues...)
	ch <- prometheus.MustNewConstMetric(c.sets, prometheus.CounterValue, float64(stats.Sets), labelValues...)
	ch <- prometheus.MustNewConstMetric(c.deletes, prometheus.CounterValue, float64(stats.Deletes), labelValues...)
	ch <- prometheus.MustNewConstMetric(c.evictions, prometheus.CounterValue, float64(stats.Evictions), labelValues...)
	ch <- prometheus.MustNewConstMetric(c.items, prometheus.GaugeValue, float64(stats.ItemCount), labelValues...)
	ch <- prometheus.MustNewConstMetric(c.bytes, prometheus.GaugeValue, float64(stats.BytesUsed), labelValues...)
}
