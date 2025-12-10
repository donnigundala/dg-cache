package compression

import (
	"bytes"
	"compress/gzip"
	"io"
)

// GzipCompressor implements the Compressor interface using gzip.
type GzipCompressor struct {
	Level int
}

const DefaultCompression = gzip.DefaultCompression

// NewGzipCompressor creates a new GzipCompressor.
// Default level is gzip.DefaultCompression.
func NewGzipCompressor(level int) *GzipCompressor {
	return &GzipCompressor{
		Level: level,
	}
}

// Compress compresses the given data using gzip.
func (c *GzipCompressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, c.Level)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decompress decompresses the given data using gzip.
func (c *GzipCompressor) Decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
