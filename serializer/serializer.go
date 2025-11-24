package serializer

// Serializer handles marshaling and unmarshaling of cache values.
// Implementations must be thread-safe.
type Serializer interface {
	// Marshal converts a Go value to bytes for storage.
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal converts bytes back to a Go value.
	// The result is stored in the value pointed to by v.
	Unmarshal(data []byte, v interface{}) error

	// Name returns the serializer name (e.g., "json", "msgpack").
	Name() string
}

// Envelope wraps values with type information for safe deserialization.
// This allows the cache to store the type alongside the value.
type Envelope struct {
	Type  string      `json:"type" msgpack:"type"`
	Value interface{} `json:"value" msgpack:"value"`
}
