package parser

import (
	"errors"
	"io"
	"strings"

	"github.com/mvader/elmo/ast"
	"github.com/mvader/elmo/scanner"
)

// ParseFile returns the AST representation of the given file.
func ParseFile(fileName string, source io.Reader) (f *ast.File, err error) {
	var p parser
	s := scanner.New(fileName, source)

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
	p.init(fileName, s)
	f = p.parseFile()
	return
}
