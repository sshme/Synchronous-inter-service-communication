package hash

import (
	"context"
	"io"
)

// Hasher defines the interface for computing file content hashes
type Hasher interface {
	ComputeHash(ctx context.Context, data io.Reader) (string, error)
	ComputeHashFromFile(ctx context.Context, filePath string) (string, error)
}
