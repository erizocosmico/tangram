package source

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// Loader finds the absolute path of files in the project and is able to
// load their source.
type Loader interface {
	// AbsPath returns the absolute path of the file.
	AbsPath(string) string
	// Load reads the source code of the file at the given relative project
	// path.
	Load(string) ([]byte, error)
}

// FsLoader is a loader from file system.
type FsLoader struct {
	// TODO(erizocosmico): root should be found by the loader itself.
	root string
}

// NewFsLoader creates a new filesystem loader with the given root.
func NewFsLoader(root string) *FsLoader {
	return &FsLoader{root}
}

// AbsPath returns the absolute path of the given path, which must be relative
// to the root of the loader.
func (l *FsLoader) AbsPath(path string) string {
	return filepath.Join(l.root, path)
}

// Load retrieves the source code of the file at the given path.
func (l *FsLoader) Load(path string) ([]byte, error) {
	p := l.AbsPath(path)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

// MemLoader is a loader that works in memory. It is intended for test
// purposes and not real use.
type MemLoader struct {
	files map[string]string
}

// NewMemLoader returns a new memory loader.
func NewMemLoader() *MemLoader {
	return &MemLoader{make(map[string]string)}
}

// Add inserts the content for the given path to the memory loader.
func (l *MemLoader) Add(path, content string) {
	l.files[path] = content
}

// AbsPath returns the absolute path of the given path.
func (l *MemLoader) AbsPath(path string) string {
	return path
}

// Load retrieves the content of the given path.
func (l *MemLoader) Load(path string) ([]byte, error) {
	if s, ok := l.files[path]; ok {
		return []byte(s), nil
	}

	return nil, os.ErrNotExist
}
