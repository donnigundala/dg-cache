package memory

import (
	"sync"
	"testing"
)

func TestMetrics_RecordHit(t *testing.T) {
	m := newMetrics()

	m.RecordHit()
	m.RecordHit()
	m.RecordHit()

	stats := m.Stats()
	if stats.Hits != 3 {
		t.Errorf("Expected 3 hits, got %d", stats.Hits)
	}
}

func TestMetrics_RecordMiss(t *testing.T) {
	m := newMetrics()

	m.RecordMiss()
	m.RecordMiss()

	stats := m.Stats()
	if stats.Misses != 2 {
		t.Errorf("Expected 2 misses, got %d", stats.Misses)
	}
}

func TestMetrics_HitRate(t *testing.T) {
	m := newMetrics()

	// 3 hits, 1 miss = 75% hit rate
	m.RecordHit()
	m.RecordHit()
	m.RecordHit()
	m.RecordMiss()

	stats := m.Stats()
	expectedRate := 0.75
	if stats.HitRate != expectedRate {
		t.Errorf("Expected hit rate %.2f, got %.2f", expectedRate, stats.HitRate)
	}
}

func TestMetrics_RecordSet(t *testing.T) {
	m := newMetrics()

	m.RecordSet(100)
	m.RecordSet(200)

	stats := m.Stats()
	if stats.Sets != 2 {
		t.Errorf("Expected 2 sets, got %d", stats.Sets)
	}
	if stats.ItemCount != 2 {
		t.Errorf("Expected 2 items, got %d", stats.ItemCount)
	}
	if stats.BytesUsed != 300 {
		t.Errorf("Expected 300 bytes, got %d", stats.BytesUsed)
	}
}

func TestMetrics_RecordUpdate(t *testing.T) {
	m := newMetrics()

	m.RecordSet(100)
	m.RecordUpdate(100, 200)

	stats := m.Stats()
	if stats.Sets != 2 {
		t.Errorf("Expected 2 sets, got %d", stats.Sets)
	}
	if stats.ItemCount != 1 {
		t.Errorf("Expected 1 item, got %d", stats.ItemCount)
	}
	if stats.BytesUsed != 200 {
		t.Errorf("Expected 200 bytes, got %d", stats.BytesUsed)
	}
}

func TestMetrics_RecordDelete(t *testing.T) {
	m := newMetrics()

	m.RecordSet(100)
	m.RecordSet(200)
	m.RecordDelete(100)

	stats := m.Stats()
	if stats.Deletes != 1 {
		t.Errorf("Expected 1 delete, got %d", stats.Deletes)
	}
	if stats.ItemCount != 1 {
		t.Errorf("Expected 1 item, got %d", stats.ItemCount)
	}
	if stats.BytesUsed != 200 {
		t.Errorf("Expected 200 bytes, got %d", stats.BytesUsed)
	}
}

func TestMetrics_RecordEviction(t *testing.T) {
	m := newMetrics()

	m.RecordSet(100)
	m.RecordEviction(100)

	stats := m.Stats()
	if stats.Evictions != 1 {
		t.Errorf("Expected 1 eviction, got %d", stats.Evictions)
	}
	if stats.ItemCount != 0 {
		t.Errorf("Expected 0 items, got %d", stats.ItemCount)
	}
	if stats.BytesUsed != 0 {
		t.Errorf("Expected 0 bytes, got %d", stats.BytesUsed)
	}
}

func TestMetrics_Reset(t *testing.T) {
	m := newMetrics()

	m.RecordHit()
	m.RecordMiss()
	m.RecordSet(100)

	m.Reset()

	stats := m.Stats()
	if stats.Hits != 0 || stats.Misses != 0 || stats.Sets != 0 {
		t.Error("Expected all counters to be reset to 0")
	}
	if stats.ItemCount != 0 || stats.BytesUsed != 0 {
		t.Error("Expected size metrics to be reset to 0")
	}
}

func TestMetrics_Concurrency(t *testing.T) {
	m := newMetrics()
	var wg sync.WaitGroup

	// Concurrent hits
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.RecordHit()
		}()
	}

	// Concurrent misses
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.RecordMiss()
		}()
	}

	wg.Wait()

	stats := m.Stats()
	if stats.Hits != 100 {
		t.Errorf("Expected 100 hits, got %d", stats.Hits)
	}
	if stats.Misses != 50 {
		t.Errorf("Expected 50 misses, got %d", stats.Misses)
	}
}

func TestMetrics_ZeroHitRate(t *testing.T) {
	m := newMetrics()

	stats := m.Stats()
	if stats.HitRate != 0.0 {
		t.Errorf("Expected 0.0 hit rate with no operations, got %.2f", stats.HitRate)
	}
}
