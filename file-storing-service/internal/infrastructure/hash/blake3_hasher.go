package hash

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/zeebo/blake3"
	"io"
	"os"
)

// BLAKE3Hasher implements the Hasher interface using BLAKE3 algorithm
type BLAKE3Hasher struct{}

// NewBLAKE3Hasher creates a new BLAKE3 hasher
func NewBLAKE3Hasher() *BLAKE3Hasher {
	return &BLAKE3Hasher{}
}

// ComputeHash calculates the BLAKE3 hash of the given data
func (h *BLAKE3Hasher) ComputeHash(ctx context.Context, data io.Reader) (string, error) {
	hasher := blake3.New()

	_, err := h.copyWithContext(ctx, hasher, data)
	if err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}

	hashBytes := hasher.Sum(nil)

	return hex.EncodeToString(hashBytes), nil
}

// ComputeHashFromFile calculates the BLAKE3 hash from file path
func (h *BLAKE3Hasher) ComputeHashFromFile(ctx context.Context, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	return h.ComputeHash(ctx, file)
}

// copyWithContext copies data with context cancellation support
func (h *BLAKE3Hasher) copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	buf := make([]byte, 32*1024)
	var written int64

	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
		}

		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = fmt.Errorf("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return written, er
			}
			break
		}
	}
	return written, nil
}
