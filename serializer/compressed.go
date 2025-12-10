package serializer

import (
	"github.com/donnigundala/dg-cache/compression"
)

// CompressedSerializer wraps another serializer and compresses the output.
type CompressedSerializer struct {
	inner      Serializer
	compressor compression.Compressor
}

// NewCompressedSerializer creates a new CompressSerializer.
func NewCompressedSerializer(inner Serializer, compressor compression.Compressor) *CompressedSerializer {
	return &CompressedSerializer{
		inner:      inner,
		compressor: compressor,
	}
}

// Marshal marshals the value using the inner serializer and then compresses it.
func (s *CompressedSerializer) Marshal(v interface{}) ([]byte, error) {
	data, err := s.inner.Marshal(v)
	if err != nil {
		return nil, err
	}
	return s.compressor.Compress(data)
}

// Unmarshal decompresses the data and then unmarshals it using the inner serializer.
func (s *CompressedSerializer) Unmarshal(data []byte, v interface{}) error {
	uncompressed, err := s.compressor.Decompress(data)
	if err != nil {
		return err
	}
	return s.inner.Unmarshal(uncompressed, v)
}

// Name returns the name of the inner serializer combined with "compressed".
func (s *CompressedSerializer) Name() string {
	return s.inner.Name()
}
