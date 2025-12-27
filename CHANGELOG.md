# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-12-27

### Added
- Initial stable release of the `dg-cache` plugin.
- Cache Manager with unified API for multiple stores.
- **Memory Driver**: Production-ready with LRU eviction, size limits, and metrics.
- **Redis Driver**: Full feature parity with tagged cache and serialization support.
- **Serialization**: JSON and Msgpack serializers for complex Go types.
- **Type-Safe Helpers**: `GetAs()`, `GetString()`, `GetInt()`, `GetBool()`, etc.
- **Container Integration**: Auto-registration of named stores with Injectable pattern.
- **Observability**: OpenTelemetry metrics for hits, misses, and latency.
- **Compression**: Gzip compression for large values (configurable threshold).
- **Reliability**: Circuit breaker pattern and enhanced retry logic.

### Features
- Thread-safe operations with context support
- TTL (Time To Live) and prefix support for key namespacing
- Remember pattern (cache-aside implementation)
- Batch operations (GetMultiple, PutMultiple)
- Atomic operations (Increment, Decrement)
- Tagged cache support (driver dependent)
- Comprehensive test coverage

### Performance
- Msgpack unmarshal: 2.6x faster than JSON
- Memory driver: O(1) operations for Get/Put/Evict
- Connection pool metrics and hit/miss ratio tracking

---

## Development History

The following versions represent the development journey leading to v1.0.0:

### 2025-12-10
- Added compression support and observability enhancements
- Implemented circuit breaker pattern

### 2025-12-08
- Enhanced memory driver with tagged cache support

### 2025-12-08
- Full container integration with Injectable pattern
- Helper functions for easier resolution

### 2025-11-24
- Merged Redis driver into `drivers/redis`
- Deprecated standalone `dg-redis` package

### 2025-11-24
- Fixed JSON serializer envelope handling

### 2025-11-24
- Added typed helper methods for type-safe retrieval
- Comprehensive documentation

### 2025-11-24
- Serialization support (JSON/Msgpack)
- Memory driver LRU eviction and size limits

### 2025-11-23
- Initial beta release with basic caching functionality
