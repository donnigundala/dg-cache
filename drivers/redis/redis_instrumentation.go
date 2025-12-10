package redis

import (
	"sync/atomic"

	cache "github.com/donnigundala/dg-cache"
)

// Stats returns the current cache statistics.
func (d *Driver) Stats() cache.Stats {
	return cache.Stats{
		Hits:    atomic.LoadInt64(&d.metrics.Hits),
		Misses:  atomic.LoadInt64(&d.metrics.Misses),
		Sets:    atomic.LoadInt64(&d.metrics.Sets),
		Deletes: atomic.LoadInt64(&d.metrics.Deletes),
	}
}

// recordHit increments the hit counter.
func (d *Driver) recordHit() {
	atomic.AddInt64(&d.metrics.Hits, 1)
}

// recordMiss increments the miss counter.
func (d *Driver) recordMiss() {
	atomic.AddInt64(&d.metrics.Misses, 1)
}

// recordSet increments the set counter.
func (d *Driver) recordSet() {
	atomic.AddInt64(&d.metrics.Sets, 1)
}

// recordDelete increments the delete counter.
func (d *Driver) recordDelete() {
	atomic.AddInt64(&d.metrics.Deletes, 1)
}
