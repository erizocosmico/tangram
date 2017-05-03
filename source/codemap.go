package source

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode/utf8"

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
	// Src is the source code of the file. It can never be asumed
	// that Src will be at offset 0, before using, seek to the start.
	Src       io.ReadSeeker
	lineIndex []lineInfo
	scanner   *scanner.Scanner
}

type lineInfo struct {
	start token.Pos
	end   token.Pos
}

type LinePos struct {
	Col  int
	Line int
}

func (li lineInfo) inLine(pos token.Pos) bool {
	return pos >= li.start && pos < li.end
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

func (s *Source) findLineStart(pos token.Pos) (token.Pos, int) {
	start, end := 0, len(s.lineIndex)

	for start < end {
		h := start + (end-start)/2
		li := s.lineIndex[h]
		if li.inLine(pos) {
			return li.start, h + 1
		} else if pos < li.start {
			end = h
		} else {
			start = h + 1
		}
	}

	if start >= len(s.lineIndex) {
		return s.lineIndex[len(s.lineIndex)-1].end, start + 1
	}
	return s.lineIndex[start].start, start + 1
}

// LinePos returns the column and line of an offset in the source.
func (s *Source) LinePos(pos token.Pos) (lp LinePos, err error) {
	start, lineNo := s.findLineStart(pos)
	if _, err = s.Src.Seek(int64(start), io.SeekStart); err != nil {
		return
	}

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, io.LimitReader(s.Src, int64(pos-start))); err != nil {
		return
	}

	line := strings.Replace(buf.String(), "\t", tab, -1)
	lp.Col = utf8.RuneCountInString(line) + 1
	lp.Line = lineNo
	return
}

type Snippet struct {
	Start int
	Lines []string
}

// tab represents the replacement of a tab for a fixed quantity of spaces
const tab = "    "

// Region returns a region of the source code beginning at the start position
// and ending at the end of the given region.
func (s *Source) Region(start, end token.Pos) (*Snippet, error) {
	lineStart, lineNo := s.findLineStart(start)
	if _, err := s.Src.Seek(int64(lineStart), io.SeekStart); err != nil {
		return nil, err
	}

	var (
		buf bytes.Buffer
		r   = bufio.NewReader(io.LimitReader(s.Src, int64(end-start)))
	)

	var prefixBuf bytes.Buffer
	if _, err := io.Copy(&prefixBuf, io.LimitReader(s.Src, int64(start-lineStart))); err != nil {
		return nil, err
	}

	for _, r := range prefixBuf.String() {
		if r == '\t' {
			buf.WriteString(tab)
		} else {
			buf.WriteRune(' ')
		}
	}

	for {
		l, _, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		buf.Write(l)
		buf.WriteRune('\n')
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n\r"), "\n")
	snippet := Snippet{lineNo, make([]string, len(lines))}
	for i, line := range lines {
		snippet.Lines[i] = strings.Replace(line, "\t", tab, -1)
	}

	return &snippet, nil
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
