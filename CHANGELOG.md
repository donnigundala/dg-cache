# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.6.2] - 2025-12-10

### Added
- **Compression Support** - Gzip compression for large values
  - Configurable compression threshold (default: 1KB)
  - Automatic compression/decompression
- **Observability** - Standardized `Stats` and Prometheus exporter
  - Connection pool metrics
  - Hit/Miss ratio tracking
  - Latency storage
- **Reliability** - Enhanced error handling
  - Circuit Breaker pattern implementation
  - Enhanced retry logic with jitter

## [1.6.1] - 2025-12-08

### Added
- **Memory Driver Enhancements**
  - Tagged cache support for memory driver
  - Stress testing suite

## [1.6.0] - 2025-12-08

### Added
- **Container Integration** - Full integration with `dg-core` container
  - Auto-registration of named cache stores (e.g., `cache.redis`, `cache.memory`)
  - Lazy loading of lazily connection caches via internal manager
- **Cache Interface** - New implementation interface for better abstraction
- **Helper Functions** - Global helpers for easier container resolution
  - `Resolve()`, `MustResolve()` - Resolve main cache manager
  - `ResolveStore()`, `MustResolveStore()` - Resolve specific named stores
- **Injectable Pattern** - `Injectable` struct for clean dependency injection in services
- **Batch Deletion** - Added `ForgetMultiple` to Store interface and drivers

## [1.3.0] - 2025-11-24

### Added
- **Redis driver now included** - Merged dg-redis into `drivers/redis`
- Redis driver with full feature parity (tagged cache, serialization)
- `NewDriverWithClient()` for shared Redis connections
- Comprehensive Redis driver documentation

### Changed
- Redis driver import path: `github.com/donnigundala/dg-cache/drivers/redis`
- Package structure now mirrors memory driver pattern

### Deprecated
- `github.com/donnigundala/dg-redis` package (use `drivers/redis` instead)

### Migration
**Old:**
```go
import "github.com/donnigundala/dg-redis"
```

**New:**
```go
import "github.com/donnigundala/dg-cache/drivers/redis"
```

## [1.2.1] - 2025-11-24

### Fixed
- JSON serializer now correctly handles Envelope-wrapped values
- Improved Envelope detection using json.RawMessage

## [1.2.0] - 2025-11-24

### Added
- Typed helper methods for type-safe retrieval
  - `GetAs()` - Generic type-safe unmarshaling
  - `GetString()` - String value retrieval
  - `GetInt()` - Integer value retrieval
  - `GetInt64()` - Int64 value retrieval
  - `GetFloat64()` - Float64 value retrieval
  - `GetBool()` - Boolean value retrieval
- Fixed `GetMultiple` serialization in Redis driver
- Comprehensive documentation
  - API Reference (docs/API.md)
  - Serialization Guide (docs/SERIALIZATION.md)
  - Memory Driver Guide (docs/MEMORY_DRIVER.md)
- Updated README with extensive examples

### Changed
- Improved README with better organization and examples
- Enhanced error handling in typed helpers

## [1.1.0] - 2025-11-24

### Added
- Serialization support for complex Go types
  - JSON serializer (default, human-readable)
  - Msgpack serializer (2.6x faster unmarshal, binary format)
  - Automatic marshaling/unmarshaling of structs, slices, maps
  - Type preservation with envelope pattern
- Memory driver enhancements
  - LRU (Least Recently Used) eviction policy
  - Configurable size limits (max items and max bytes)
  - Metrics collection (hits, misses, evictions, size tracking)
  - Configurable cleanup intervals
  - Thread-safe operations
- Redis driver serialization integration
  - Automatic serialization in Get/Put operations
  - Backward compatible with raw string values
  - Configurable serializer choice (JSON/msgpack)

### Changed
- Memory driver now production-ready with size limits and eviction
- Redis driver updated to use serialization by default
- Improved performance with msgpack serializer option

### Performance
- Msgpack unmarshal: 2.6x faster than JSON (172ns vs 443ns)
- Msgpack payload: 30-50% smaller than JSON
- Memory driver: O(1) operations for Get/Put/Evict

## [1.0.0] - 2025-11-23

### Added
- Initial release of dg-cache
- Cache Manager with unified API
- Multiple store support
- Memory driver (in-memory caching)
- Redis driver integration (via dg-redis)
- Basic operations (Get, Put, Forget, Flush, Has, Missing, Pull)
- Batch operations (GetMultiple, PutMultiple)
- Atomic operations (Increment, Decrement)
- Remember pattern (cache-aside implementation)
- Tagged cache support (driver dependent)
- Fluent configuration API
- Service provider for dg-core integration

### Features
- Thread-safe operations
- Context support
- TTL (Time To Live) support
- Prefix support for key namespacing
- Error handling with typed errors
- Comprehensive test coverage

## [Unreleased]

### Planned
- Compression support for large values
- Additional serializers (protobuf, avro)
- Cache warming capabilities
- Distributed locking support
- Metrics interface for monitoring
- Prometheus integration
- Advanced eviction policies (LFU, FIFO)
- Cache tags for memory driver

---

## Version History

- **1.2.0** - Typed helpers and comprehensive documentation
- **1.1.0** - Serialization support and memory driver enhancements
- **1.0.0** - Initial release with basic caching functionality
