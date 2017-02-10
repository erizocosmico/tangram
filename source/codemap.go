package source

import (
	"bufio"
	"bytes"
	"io"
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

	content, err := cm.loader.Load(path)
	if err != nil {
		return err
	}

	cm.files[path] = &Source{path, content}
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
	Src []byte
}

type Line struct {
	Num     int64
	Content string
}

// Region returns a region of the source code.
func (s *Source) Region(start, end int64) ([]Line, error) {
	var (
		r     = bufio.NewReader(bytes.NewBuffer(s.Src))
		l     int64
		lines []Line
	)

	for {
		l++
		ln, _, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if l >= start && l <= end {
			lines = append(lines, Line{l, string(ln)})
		} else if l > end {
			break
		}
	}

	return lines, nil
}
