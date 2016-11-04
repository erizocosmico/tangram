package parser

import (
	"errors"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/mvader/elm-compiler/ast"
	"github.com/mvader/elm-compiler/lexer"
)

func ParseFile(fileName string, source io.Reader) (f *ast.File, err error) {
	var p parser
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(bailout); !ok {
				panic(r)
			}
		}

		if len(p.errors) > 0 {
			var errs []string
			for _, e := range p.errors {
				errs = append(errs, color.RedString("error: ")+e.Error())
			}
			err = errors.New(strings.Join(errs, "\n"))
		}
	}()

	l := lexer.New(fileName, source)
	go l.Run()

	p.init(fileName, l)
	f = p.parseFile()
	return
}
