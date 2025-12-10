package serializer

import (
	"testing"

	"github.com/donnigundala/dg-cache/compression"
	"github.com/stretchr/testify/assert"
)

func TestCompressedSerializer(t *testing.T) {
	inner := NewJSONSerializer()
	comp := compression.NewGzipCompressor(compression.DefaultCompression)
	serializer := NewCompressedSerializer(inner, comp)

	data := map[string]string{"foo": "bar"}

	// Marshal
	bytes, err := serializer.Marshal(data)
	assert.NoError(t, err)
	assert.NotNil(t, bytes)

	// Unmarshal
	var result map[string]string
	err = serializer.Unmarshal(bytes, &result)
	assert.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestCompressedSerializer_Name(t *testing.T) {
	inner := NewJSONSerializer()
	comp := compression.NewGzipCompressor(compression.DefaultCompression)
	serializer := NewCompressedSerializer(inner, comp)

	assert.Equal(t, "json", serializer.Name())
}
