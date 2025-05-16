package compressor

import (
	"fmt"
	"p2compression/internal/compressor/huffman"
	"p2compression/internal/compressor/paq"
	"p2compression/internal/compressor/sevenzip"
)

// Compressor defines the interface that all compression algorithms must implement.
type Compressor interface {
	Compress(inputFile string, outputFile string, params map[string]interface{}) error
}

// GetCompressor returns a Compressor implementation based on the algorithm name.
// This function will be expanded as more algorithms are implemented.
func GetCompressor(algorithm string) (Compressor, error) {
	switch algorithm {
	// Cases for different algorithms will be added here
	case "7zip":
		return &sevenzip.SevenZipCompressor{}, nil
	case "huffman":
		return &huffman.HuffmanCompressor{}, nil
	case "paq":
		return &paq.PAQCompressor{}, nil
	default:
		return nil, fmt.Errorf("unsupported compression algorithm: %s", algorithm)
	}
}
