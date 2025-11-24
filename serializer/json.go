package serializer

import (
	"encoding/json"
	"reflect"
)

// JSONSerializer implements the Serializer interface using JSON encoding.
// It provides human-readable serialization with type preservation.
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSON serializer.
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Marshal converts a Go value to JSON bytes with type information.
func (s *JSONSerializer) Marshal(v interface{}) ([]byte, error) {
	// Handle nil values
	if v == nil {
		return json.Marshal(nil)
	}

	// For simple types (string, int, bool, etc.), store directly without envelope
	// This maintains backward compatibility and reduces overhead
	switch v.(type) {
	case string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return json.Marshal(v)
	}

	// For complex types, wrap with type information
	envelope := Envelope{
		Type:  reflect.TypeOf(v).String(),
		Value: v,
	}
	return json.Marshal(envelope)
}

// Unmarshal converts JSON bytes back to a Go value.
func (s *JSONSerializer) Unmarshal(data []byte, v interface{}) error {
	// Try to unmarshal directly first (for simple types and backward compatibility)
	if err := json.Unmarshal(data, v); err == nil {
		return nil
	}

	// If that fails, try to unmarshal as an envelope
	var envelope Envelope
	envelope.Value = v
	return json.Unmarshal(data, &envelope)
}

// Name returns the serializer name.
func (s *JSONSerializer) Name() string {
	return "json"
}
