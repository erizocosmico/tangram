package source

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/erizocosmico/elmo/token"
)

// CodeMap contains a set of source code files.
type CodeMap struct {
	loader Loader
	files  map[string]*Source
}

// NewCodeMap returns a new code map.
func NewCodeMap(loader Loader) *CodeMap {
	return &CodeMap{loader, make(map[string]*Source)}
}

// Add includes a new file in the codemap. The path given must be a relative
// path in the project.
func (cm *CodeMap) Add(path string) error {
	if _, ok := cm.files[path]; ok {
		return nil
	}

	src, err := cm.loader.Load(path)
	if err != nil {
		return err
	}

	cm.files[path] = &Source{path, src}
	return nil
}

// Close closes all the source files that implement io.Closer.
func (cm *CodeMap) Close() error {
	for _, f := range cm.files {
		if f, ok := f.Src.(io.Closer); ok {
			if err := f.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Source returns the source for the given path.
func (cm *CodeMap) Source(path string) *Source {
	return cm.files[path]
}

// Source represents a single source file of code.
type Source struct {
	// Path is the absolute path of the file.
	Path string
	// Src is the source code of the file.
	Src io.ReadSeeker
}

// Region returns a region of the source code beginning at the start position
// and ending at the first line ending after the end position or eof.
func (s *Source) Region(start, end token.Pos) ([]string, error) {
	if _, err := s.Src.Seek(int64(start), io.SeekStart); err != nil {
		return nil, err
	}

	var (
		buf  bytes.Buffer
		size = int(end - start)
		r    = bufio.NewReader(s.Src)
	)

	for {
		l, _, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		buf.Write(l)
		if buf.Len() >= size {
			break
		}
	}

	return strings.Split(buf.String(), "\n"), nil
}
