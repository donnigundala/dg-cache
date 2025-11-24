package serializer

import (
	"reflect"

	"github.com/vmihailenco/msgpack/v5"
)

// MsgpackSerializer implements the Serializer interface using MessagePack encoding.
// It provides faster, more compact serialization compared to JSON.
type MsgpackSerializer struct{}

// NewMsgpackSerializer creates a new msgpack serializer.
func NewMsgpackSerializer() *MsgpackSerializer {
	return &MsgpackSerializer{}
}

// Marshal converts a Go value to msgpack bytes with type information.
func (s *MsgpackSerializer) Marshal(v interface{}) ([]byte, error) {
	// Handle nil values
	if v == nil {
		return msgpack.Marshal(nil)
	}

	// For simple types, store directly without envelope
	switch v.(type) {
	case string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return msgpack.Marshal(v)
	}

	// For complex types, wrap with type information
	envelope := Envelope{
		Type:  reflect.TypeOf(v).String(),
		Value: v,
	}
	return msgpack.Marshal(envelope)
}

// Unmarshal converts msgpack bytes back to a Go value.
func (s *MsgpackSerializer) Unmarshal(data []byte, v interface{}) error {
	// Try to unmarshal directly first (for simple types)
	if err := msgpack.Unmarshal(data, v); err == nil {
		return nil
	}

	// If that fails, try to unmarshal as an envelope
	var envelope Envelope
	envelope.Value = v
	return msgpack.Unmarshal(data, &envelope)
}

// Name returns the serializer name.
func (s *MsgpackSerializer) Name() string {
	return "msgpack"
}
