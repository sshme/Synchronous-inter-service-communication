package file

import (
	"errors"
	"time"
)

// File represents a file entity in the domain
type File struct {
	ID          string
	Name        string
	Size        int64
	ContentType string
	Location    string
	Hash        string
	UploadedAt  time.Time
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

// NewFile creates a new File domain entity
func NewFile(name, contentType string, size int64) (*File, error) {
	if size > MaxFileSize {
		return nil, errors.New("file size exceeds maximum allowed limit")
	}
	if size <= 0 {
		return nil, errors.New("file size must be greater than zero")
	}
	if contentType != ContentType {
		return nil, errors.New("content type must be " + ContentType)
	}

	now := time.Now()
	return &File{
		Name:        name,
		Size:        size,
		ContentType: contentType,
		UploadedAt:  now,
		UpdatedAt:   now,
		CreatedAt:   now,
	}, nil
}

// SetHash sets the content hash for the file
func (f *File) SetHash(hash string) error {
	if hash == "" {
		return errors.New("hash cannot be empty")
	}
	f.Hash = hash
	f.UpdatedAt = time.Now()
	return nil
}
