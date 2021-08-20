package storage

import "io"

// Backend is the interface that wraps the basic file operations.
type Backend interface {
	// Name returns the name of the backend implementation.
	Name() string

	// Reader returns a ReadCloser of the file.
	Reader(container, object string) (io.ReadCloser, error)
	// Reader returns a WriteCloser of the file.
	Writer(container, object string) (io.WriteCloser, error)
	// Copy copies a file.
	Copy(sc, so, dc, do string) error

	// FilenamesFrom list all the object names from the given prefix: `container/prefix'.
	FilenamesFrom(prefix string) ([]string, error)

	// Remove deletes the given file.
	Remove(container, object string) error
	// RemoveAll deletes all the file and folders.
	RemoveAll(path string) error
	// Cleanup cleans useless artifacts in storage.
	Cleanup() error
}
