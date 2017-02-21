package parser

import (
	"errors"
	"io"
	"strings"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/diagnostic"
	"github.com/erizocosmico/elmo/scanner"
	"github.com/erizocosmico/elmo/source"
)

// ParseMode specifies the type of mode in which the parser will be run.
// ParseMode can be used to only parse certain parts of a file.
type ParseMode int

const (
	// FullParse parses completely the source file.
	FullParse ParseMode = iota
	// OnlyImports parses only package definition and imports.
	OnlyImports
	// ImportsAndFixity parses only package definition, imports and
	// fixity declarations.
	ImportsAndFixity
)

// Session represents the current parsing session.
type Session struct {
	*diagnostic.Diagnoser
	*source.CodeMap
}

// NewSession creates a new parsing session with a way of diagnosing errors
// and a code map.
func NewSession(d *diagnostic.Diagnoser, cm *source.CodeMap) *Session {
	return &Session{d, cm}
}

// ParseFile returns the AST representation of the given file.
func ParseFile(fileName string, src io.Reader, mode ParseMode) (f *ast.File, err error) {
	// TODO(erizocosmico): correctly set root
	cm := source.NewCodeMap(source.NewFsLoader("."))
	sess := NewSession(
		diagnostic.NewDiagnoser(cm, diagnostic.Stderr(true, true)),
		cm,
	)
	p := newParser(sess)
	s := scanner.New(fileName, src)

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(bailout); !ok {
				panic(r)
			}
		}

		if len(p.errors) > 0 {
			var errs []string
			for _, e := range p.errors {
				errs = append(errs, e.Error())
			}
			err = errors.New(strings.Join(errs, "\n"))
		}
	}()

	go s.Run()
	p.init(fileName, s, mode)
	f = p.parseFile()
	return
}
