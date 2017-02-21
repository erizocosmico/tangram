package parser

import (
	"fmt"
	"testing"

	"github.com/erizocosmico/elmo/ast"

	"github.com/stretchr/testify/require"
)

const parseFixture = `
module Foo

import Foo.Bar
import Foo.Bar.Baz exposing (bar, baz)

foo : Int
foo = 45

(?) : comparable -> comparable -> Bool
(?) cmp1 cmp2 = cmp1 > cmp2

infixr 6 ?
infixl 7 :>

(:>) : comparable -> comparable -> Bool
(:>) cmp1 cmp2 = cmp1 > cmp2
`

func TestParseFile_OnlyFixity(t *testing.T) {
	require := require.New(t)

	p := stringParser(parseFixture)
	p.mode = ImportsAndFixity
	var f *ast.File
	func() {
		defer assertEOF(t, "TestParseFile_OnlyFixity", true)
		defer p.sess.Emit()
		f = p.parseFile()

		require.Len(f.Imports, 2, "should have 2 imports")
		require.Equal("Foo", f.Module.Name.String(), "module name")

		require.Len(f.Decls, 2, "should have 2 decls")
		assertFixity(t, f.Decls[0], "?", 6, ast.RightAssoc)
		assertFixity(t, f.Decls[1], ":>", 7, ast.LeftAssoc)
	}()
	require.True(p.sess.IsOK(), "no errors should be returned")
	require.NotNil(f)
}

func assertFixity(t *testing.T, d ast.Decl, op string, precedence uint, assoc ast.Associativity) {
	decl, ok := d.(*ast.InfixDecl)
	require.True(t, ok, "should be InfixDecl")
	require.Equal(t, op, decl.Op.Name)
	require.Equal(t, fmt.Sprint(precedence), decl.Priority.Value)
	require.Equal(t, assoc, decl.Assoc)
}
