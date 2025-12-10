package memory

import (
	"context"
	"time"

	cache "github.com/donnigundala/dg-cache"
)

// taggedCache implements the TaggedStore interface for the memory driver.
type taggedCache struct {
	*Driver
	tags []string
}

// Tags returns a new TaggedStore instance with the given tags.
func (d *Driver) Tags(tags ...string) cache.TaggedStore {
	return &taggedCache{
		Driver: d,
		tags:   tags,
	}
}

// Tags extends the current tags with new ones.
func (t *taggedCache) Tags(tags ...string) cache.TaggedStore {
	return &taggedCache{
		Driver: t.Driver,
		tags:   append(t.tags, tags...),
	}
}

// Put stores a value in the cache with tags.
func (t *taggedCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	err := t.Driver.put(key, value, ttl)
	if err != nil {
		return err
	}

	t.Driver.addKeyTags(t.Driver.prefixKey(key), t.tags)
	return nil
}

// PutMultiple stores multiple values in the cache with tags.
func (t *taggedCache) PutMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Logic from Driver.PutMultiple but calling internal put (or implementing it here as putMultiple is not refactored yet)
	// Actually Driver.PutMultiple isn't refactored. Let's make it simple: loop and put.
	// Optimizing: we can just use the same logic as PutMultipe but adding tags.

	// Logic from Driver.PutMultiple but calling internal put

	for key, value := range items {
		// Use internal PUT logic for each item?
		// Or replicate PutMultiple logic to avoid overhead?
		// Replicating logic for batch efficiency (avoiding repeated eviction checks/metrics update if possible, but internal Put handles it)
		// For simplicity/correctness, let's just reuse D.put if we don't have putMultiple
		err := t.Driver.put(key, value, ttl)
		if err != nil {
			return err
		}
		t.Driver.addKeyTags(t.Driver.prefixKey(key), t.tags)
	}

	return nil
}

// Forever stores a value in the cache indefinitely with tags.
func (t *taggedCache) Forever(ctx context.Context, key string, value interface{}) error {
	return t.Put(ctx, key, value, 0)
}

// Flush removes all items associated with the current tags (or any of them).
// For tagged cache, Flush() usually means "flush the tags", i.e. remove all keys that have these tags.
func (t *taggedCache) Flush(ctx context.Context) error {
	return t.Driver.FlushTags(ctx, t.tags...)
}

// FlushTags removes all items associated with the given tags.
func (d *Driver) FlushTags(ctx context.Context, tags ...string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Collect all keys to remove to avoid modifying map while iterating
	keysToRemove := make(map[string]bool)

	for _, tag := range tags {
		if keys, ok := d.tags[tag]; ok {
			for key := range keys {
				keysToRemove[key] = true
			}
		}
	}

	// Remove keys
	for key := range keysToRemove {
		// key is already prefixed in d.tags
		// We need to call internal forget with PREFIXED key logic?
		// d.forget expects NON-prefixed key usually if it calls prefixKey.
		// Wait. In addKeyTags, we passed `t.Driver.prefixKey(key)`.
		// So `d.tags` stores PREFIXED keys.

		// d.forget calls `d.prefixKey(key)`.
		// If we pass a prefixed key to d.forget, it will double prefix it!

		// We need an internal method `forgetPrefixed(prefixedKey)` or `removeItem(prefixedKey)`.
		// Or we can just do the deletion logic here since we are inside the package.

		d.removeKeyTags(key) // key is prefixed
		delete(d.items, key)
		delete(d.nodes, key)
	}

	return nil
}

// We need to ensure we don't double-prefix when removing.
// Let's verify `d.forget` implementation from previous step.
// func (d *Driver) forget(key string) error {
// 	prefixedKey := d.prefixKey(key)
// 	...
// }
// So `forget` expects UN-prefixed key.
// But `d.tags` stores PREFIXED keys.
// So we cannot call `d.forget`.
// We must duplicate deletion logic or create `forgetItem(prefixedKey)`.
// Duplication is fine for now as it's just 3 lines: removeKeyTags, delete items, delete nodes.
