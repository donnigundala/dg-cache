package redis

import (
	"context"
	"time"

	cache "github.com/donnigundala/dg-cache"
	"github.com/redis/go-redis/v9"
)

// TaggedCache implements the TaggedStore interface for Redis.
type TaggedCache struct {
	*Driver
	tags []string
}

// Tags returns a new TaggedStore instance with the given tags.
func (d *Driver) Tags(tags ...string) cache.TaggedStore {
	return &TaggedCache{
		Driver: d,
		tags:   tags,
	}
}

// Tags adds more tags to the existing TaggedCache.
func (c *TaggedCache) Tags(tags ...string) cache.TaggedStore {
	return &TaggedCache{
		Driver: c.Driver,
		tags:   append(c.tags, tags...),
	}
}

// tagKey returns the Redis key for a tag set.
func (c *TaggedCache) tagKey(tag string) string {
	return c.prefix + ":tag:" + tag
}

// addTags adds the key to the tag sets.
func (c *TaggedCache) addTags(ctx context.Context, key string) error {
	if len(c.tags) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()
	prefixedKey := c.prefixKey(key)

	for _, tag := range c.tags {
		pipe.SAdd(ctx, c.tagKey(tag), prefixedKey)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Put stores a value in the cache and associates it with the tags.
func (c *TaggedCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Serialize the value
	data, err := c.serializer.Marshal(value)
	if err != nil {
		return err
	}

	// Use a pipeline to ensure both operations happen
	pipe := c.client.Pipeline()

	// Set the value
	pipe.Set(ctx, c.prefixKey(key), data, ttl)

	// Add to tag sets
	prefixedKey := c.prefixKey(key)
	for _, tag := range c.tags {
		pipe.SAdd(ctx, c.tagKey(tag), prefixedKey)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// PutMultiple stores multiple values and associates them with the tags.
func (c *TaggedCache) PutMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	pipe := c.client.Pipeline()

	for key, value := range items {
		// Serialize each value
		data, err := c.serializer.Marshal(value)
		if err != nil {
			return err
		}

		prefixedKey := c.prefixKey(key)
		pipe.Set(ctx, prefixedKey, data, ttl)

		for _, tag := range c.tags {
			pipe.SAdd(ctx, c.tagKey(tag), prefixedKey)
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Increment increments a value and associates it with the tags.
func (c *TaggedCache) Increment(ctx context.Context, key string, value int64) (int64, error) {
	// We can't easily pipeline the return value of IncrBy with SAdd if we want to return it immediately
	// But we can just run them sequentially or use a transaction.
	// For simplicity and performance, we'll use a pipeline but we need the result.

	pipe := c.client.Pipeline()
	incr := pipe.IncrBy(ctx, c.prefixKey(key), value)

	prefixedKey := c.prefixKey(key)
	for _, tag := range c.tags {
		pipe.SAdd(ctx, c.tagKey(tag), prefixedKey)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

// Decrement decrements a value and associates it with the tags.
func (c *TaggedCache) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	pipe := c.client.Pipeline()
	decr := pipe.DecrBy(ctx, c.prefixKey(key), value)

	prefixedKey := c.prefixKey(key)
	for _, tag := range c.tags {
		pipe.SAdd(ctx, c.tagKey(tag), prefixedKey)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return decr.Val(), nil
}

// Forever stores a value indefinitely and associates it with the tags.
func (c *TaggedCache) Forever(ctx context.Context, key string, value interface{}) error {
	return c.Put(ctx, key, value, 0)
}

// FlushTags removes all items associated with the given tags.
func (c *TaggedCache) FlushTags(ctx context.Context, tags ...string) error {
	if len(tags) == 0 {
		tags = c.tags
	}

	if len(tags) == 0 {
		return nil
	}

	// Load Lua script
	// In a real implementation, we should embed this or load it once
	script := redis.NewScript(`
		local prefix = ARGV[1]
		local keysToDelete = {}
		local tagsToDelete = {}

		for i, tagName in ipairs(KEYS) do
			local tagKey = prefix .. ":tag:" .. tagName
			table.insert(tagsToDelete, tagKey)
			
			local keys = redis.call("SMEMBERS", tagKey)
			for _, key in ipairs(keys) do
				table.insert(keysToDelete, key)
			end
		end

		if #keysToDelete > 0 then
			for i = 1, #keysToDelete, 1000 do
				local chunk = {}
				for j = i, math.min(i + 999, #keysToDelete) do
					table.insert(chunk, keysToDelete[j])
				end
				redis.call("DEL", unpack(chunk))
			end
		end

		if #tagsToDelete > 0 then
			redis.call("DEL", unpack(tagsToDelete))
		end

		return #keysToDelete
	`)

	return script.Run(ctx, c.client, tags, c.prefix).Err()
}
