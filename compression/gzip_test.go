package compression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzipCompressor(t *testing.T) {
	compressor := NewGzipCompressor(DefaultCompression)

	data := []byte("hello world hello world hello world")

	// Compress
	compressed, err := compressor.Compress(data)
	assert.NoError(t, err)
	assert.NotNil(t, compressed)
	assert.True(t, len(compressed) < len(data) || len(data) < 50, "compressed data should be smaller for large inputs")

	// Decompress
	decompressed, err := compressor.Decompress(compressed)
	assert.NoError(t, err)
	assert.Equal(t, data, decompressed)
}
