package fs

import "os"

// DefaultFilesystem is the default implementation of
// the FileSystem interface. It uses OSFileSystem to make
// functions of the os package the default for users.
var DefaultFilesystem = OSFileSystem{}

// OSFileSystem wraps the functions of the FileSystem
// interface around functions of the os package to offer
// an abstraction for the golang library.
type OSFileSystem struct{}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file descriptor
// has mode O_RDONLY. If there is an error, it will be of type *os.PathError.
func (OSFileSystem) Open(name string) (File, error) {
	return os.Open(name)
}

// Stat returns the os.FileInfo structure describing file.
// If there is an error, it will be of type *os.PathError.
func (OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
