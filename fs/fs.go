package fs

import (
	"io"
	"os"
)

// FileSystem is an interface that groups the common functions
// Open and Stat of the os package.
type FileSystem interface {
	Open(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
}

// File is an interface that groups several read interfaces of
// the io package together with the Stat function of the os package
// to make reading astractions of the filesystem possible.
type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	Stat() (os.FileInfo, error)
}
