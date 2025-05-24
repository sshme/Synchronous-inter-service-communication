package hash

import (
	"context"
	"io"
)

// Hasher defines the interface for computing file content hashes
type Hasher interface {
	// ComputeHash calculates the SHA256 hash of the given data
	ComputeHash(ctx context.Context, data io.Reader) (string, error)

	// ComputeHashFromFile calculates the SHA256 hash from file path
	ComputeHashFromFile(ctx context.Context, filePath string) (string, error)
}
