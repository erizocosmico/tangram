package source

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// Loader finds the absolute path of files in the project and is able to
// load their source.
type Loader interface {
	// Exists reports whether the file exists or not.
	Exists(string) (bool, error)
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

func (l *FsLoader) Exists(path string) (bool, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	defer f.Close()

	return true, nil
}

func (l *FsLoader) AbsPath(path string) string {
	return filepath.Join(l.root, path)
}

func (l *FsLoader) Load(path string) ([]byte, error) {
	p := l.AbsPath(path)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

type MemLoader struct {
	files map[string]string
}

func NewMemLoader() *MemLoader {
	return &MemLoader{make(map[string]string)}
}

func (l *MemLoader) Add(path, content string) {
	l.files[path] = content
}

func (l *MemLoader) Exists(path string) (bool, error) {
	_, ok := l.files[path]
	return ok, nil
}

func (l *MemLoader) AbsPath(path string) string {
	return filepath.Join("/", path)
}

func (l *MemLoader) Load(path string) ([]byte, error) {
	if s, ok := l.files[path]; ok {
		return []byte(s), nil
	}

	return nil, os.ErrNotExist
}
