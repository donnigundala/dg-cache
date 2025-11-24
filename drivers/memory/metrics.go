package memory

import "sync"

// Metrics tracks cache statistics.
type Metrics struct {
	mu sync.RWMutex

	// Operation counters
	hits      int64
	misses    int64
	sets      int64
	deletes   int64
	evictions int64

	// Size tracking
	itemCount int
	bytesUsed int64
}

// newMetrics creates a new Metrics instance.
func newMetrics() *Metrics {
	return &Metrics{}
}

// RecordHit increments the hit counter.
func (m *Metrics) RecordHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hits++
}

// RecordMiss increments the miss counter.
func (m *Metrics) RecordMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.misses++
}

// RecordSet increments the set counter and updates size tracking.
func (m *Metrics) RecordSet(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sets++
	m.bytesUsed += bytes
	m.itemCount++
}

// RecordUpdate updates size tracking for an existing item.
func (m *Metrics) RecordUpdate(oldBytes, newBytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sets++
	m.bytesUsed = m.bytesUsed - oldBytes + newBytes
}

// RecordDelete decrements the item count and updates size tracking.
func (m *Metrics) RecordDelete(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deletes++
	m.bytesUsed -= bytes
	m.itemCount--
}

// RecordEviction increments the eviction counter and updates size tracking.
func (m *Metrics) RecordEviction(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evictions++
	m.bytesUsed -= bytes
	m.itemCount--
}

// Stats returns a snapshot of current cache statistics.
func (m *Metrics) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.hits + m.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(m.hits) / float64(total)
	}

	return Stats{
		Hits:      m.hits,
		Misses:    m.misses,
		Sets:      m.sets,
		Deletes:   m.deletes,
		Evictions: m.evictions,
		ItemCount: m.itemCount,
		BytesUsed: m.bytesUsed,
		HitRate:   hitRate,
	}
}

// Reset resets all metrics to zero.
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hits = 0
	m.misses = 0
	m.sets = 0
	m.deletes = 0
	m.evictions = 0
	m.itemCount = 0
	m.bytesUsed = 0
}

// Stats represents cache statistics at a point in time.
type Stats struct {
	// Hits is the number of cache hits.
	Hits int64

	// Misses is the number of cache misses.
	Misses int64

	// Sets is the number of set operations.
	Sets int64

	// Deletes is the number of delete operations.
	Deletes int64

	// Evictions is the number of evicted items.
	Evictions int64

	// ItemCount is the current number of items in the cache.
	ItemCount int

	// BytesUsed is the estimated total size of cached items in bytes.
	BytesUsed int64

	// HitRate is the cache hit rate (hits / (hits + misses)).
	HitRate float64
}
