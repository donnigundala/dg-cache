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
	// 1. Try to unmarshal as an Envelope first
	// We use a temporary struct with RawMessage to defer unmarshaling of the value
	type tempEnvelope struct {
		Type  string          `json:"type"`
		Value json.RawMessage `json:"value"`
	}

	var temp tempEnvelope
	if err := json.Unmarshal(data, &temp); err == nil && temp.Type != "" {
		// It's a valid envelope, unmarshal the inner value into v
		return json.Unmarshal(temp.Value, v)
	}

	// 2. Fallback: Unmarshal directly (for simple types or backward compatibility)
	return json.Unmarshal(data, v)
}

// Name returns the serializer name.
func (s *JSONSerializer) Name() string {
	return "json"
}
