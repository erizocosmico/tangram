package source

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/erizocosmico/elmo/scanner"
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

	source, err := NewSource(path, src)
	if err != nil {
		return err
	}

	cm.files[path] = source
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
	Src       io.ReadSeeker
	lineIndex []lineInfo
	scanner   *scanner.Scanner
}

type lineInfo struct {
	start token.Pos
	end   token.Pos
}

func (li lineInfo) inLine(pos token.Pos) bool {
	return pos >= li.start && pos <= li.end
}

func NewSource(path string, src io.ReadSeeker) (*Source, error) {
	s := &Source{path, src, nil, nil}
	if err := s.makeLineIndex(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Source) makeLineIndex() (err error) {
	var (
		reader = bufio.NewReader(s.Src)
		r      rune
		start  token.Pos
		pos    token.Pos
	)

	for {
		r, _, err = reader.ReadRune()
		if err == io.EOF {
			err = nil
			if start != pos {
				s.lineIndex = append(s.lineIndex, lineInfo{start, pos})
			}
			break
		}

		if err != nil {
			goto cleanup
		}

		pos++
		if r == '\n' || r == '\r' {
			s.lineIndex = append(s.lineIndex, lineInfo{start, pos})
			start = pos
		}
	}

cleanup:
	if _, err := s.Src.Seek(0, io.SeekStart); err != nil {
		return err
	}

	return
}

func (s *Source) findLineStart(pos token.Pos) token.Pos {
	start, end := 0, len(s.lineIndex)

	for start < end {
		h := start + (end-start)/2
		li := s.lineIndex[h]
		if li.inLine(pos) {
			return li.start
		} else if pos < li.start {
			end = h
		} else {
			start = h + 1
		}
	}

	return s.lineIndex[start].start
}

// Region returns a region of the source code beginning at the start position
// and ending at the first line ending after the end position or eof.
func (s *Source) Region(start, end token.Pos) ([]string, error) {
	start = s.findLineStart(start)
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
		buf.WriteRune('\n')
		if buf.Len() >= size {
			break
		}
	}

	return strings.Split(buf.String(), "\n"), nil
}

// Scanner returns a scanner for this source with all the tokens parsed.
func (s *Source) Scanner() *scanner.Scanner {
	if s.scanner == nil {
		s.scanner = scanner.New(s.Path, s.Src)
		s.scanner.Run()
	}

	s.scanner.Reset()
	return s.scanner
}
