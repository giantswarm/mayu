package fs

import (
	"bytes"
	"errors"
	"os"
	"time"
)

var (
	ErrFileOperationNotPermitted = errors.New("fake file does not exist and therefore this operation is not permitted.")
	ErrFileDoesNotExist          = errors.New("fake file does not exist")
)

// FakeFilesystem in memory implementations for the functions
// of the FileSystem interface to offer an implementation
// that can be used during in tests.
type FakeFilesystem struct {
	files map[string]FakeFile
}

// NewFakeFilesystemWithFiles creates a new FakeFilesystem
// instance bundled with a list of FakeFile instances that
// can be passed.
func NewFakeFilesystemWithFiles(fs []FakeFile) FakeFilesystem {
	fileMap := make(map[string]FakeFile)
	for _, f := range fs {
		fileMap[f.Name] = f
	}

	return FakeFilesystem{files: fileMap}
}

// Open searches in the internal map of FakeFile instances and
// returns it for reading. If successful, methods on the returned
// file can be used for reading. If the FakeFile instance cannot
// be found in the internal map an error of type *PathError will
// be returned.
func (ff FakeFilesystem) Open(name string) (File, error) {
	file, ok := ff.files[name]
	if !ok {
		return FakeFile{}, &os.PathError{"stat", name, ErrFileDoesNotExist}
	}

	return file, nil
}

// Stat returns the FakeFileInfo structure describing a FakeFile instance
// Found in the internal FakeFile map. If the FakeFile instance cannot
// be found in the internal map an error of type *PathError will
// be returned.
func (ff FakeFilesystem) Stat(name string) (os.FileInfo, error) {
	file, ok := ff.files[name]
	if !ok {
		return FakeFileInfo{}, &os.PathError{"stat", name, ErrFileDoesNotExist}
	}

	return file.Stat()
}

// Fake represents are readable in memory version of os.File which
// can be used for testing.
type FakeFile struct {
	Name    string
	Mode    os.FileMode
	ModTime time.Time
	Buffer  *bytes.Reader
}

// NewFakeFile creates a new FakeFile instances based on a file name
// and it's contents from strings. The content of the new instance will
// be stored in an internal bytes.Reader instance. The default file mode
// will be 0777 and the last modification time the moment when the
// function is called.
func NewFakeFile(name, content string) FakeFile {
	return FakeFile{
		Name:    name,
		Mode:    os.FileMode(0777),
		ModTime: time.Now(),
		Buffer:  bytes.NewReader([]byte(content)),
	}
}

// Close wraps io.Closer's functionality and does no operation.
func (f FakeFile) Close() error {
	return nil
}

// Read wraps io.Reader's functionality around the internal
// bytes.Reader instance.
func (f FakeFile) Read(p []byte) (n int, err error) {
	return f.Buffer.Read(p)
}

// ReadAt wraps io.ReaderAt's functionality around the internalt
// bytes.Reader instance.
func (f FakeFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.Buffer.ReadAt(p, off)
}

// Seek wraps io.Seeker's functionality around the internalt
// bytes.Reader instance.
func (f FakeFile) Seek(offset int64, whence int) (int64, error) {
	return f.Buffer.Seek(offset, whence)
}

// Stat returns the FakeFileInfo structure describing the FakeFile
// instance.
func (f FakeFile) Stat() (os.FileInfo, error) {
	return FakeFileInfo{File: f}, nil
}

// FakeFileInfo describes a wrapped FakeFile instance and is returned
// by FakeFile.Stat
type FakeFileInfo struct {
	File FakeFile
}

// Name returns the base name of the FakeFile instance.
func (fi FakeFileInfo) Name() string {
	return fi.File.Name
}

// Size returns the length in bytes of the file's
// internal bytes.Reader instance.
func (fi FakeFileInfo) Size() int64 {
	return int64(fi.File.Buffer.Len())
}

// Mode returns file mode bits of the FakeFile instance.
func (fi FakeFileInfo) Mode() os.FileMode {
	return fi.File.Mode
}

// ModTime returns the modification time of the FakeFile instance.
func (fi FakeFileInfo) ModTime() time.Time {
	return fi.File.ModTime
}

// IsDir always return false since it only uses FakeFile instances.
func (fi FakeFileInfo) IsDir() bool {
	return false
}

// Sys always returns nil to stay conformant to the os.FileInfo interface.
func (fi FakeFileInfo) Sys() interface{} {
	return nil
}
