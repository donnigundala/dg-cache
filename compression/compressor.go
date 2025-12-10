package compression

// Compressor handles compression and decompression of data.
// Implementations must be thread-safe.
type Compressor interface {
	// Compress compresses the given data.
	Compress(data []byte) ([]byte, error)

	// Decompress decompresses the given data.
	Decompress(data []byte) ([]byte, error)
}
