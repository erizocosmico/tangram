package parser

import (
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
	var f *ast.File
	func() {
		defer assertEOF(t, "TestParseFile_OnlyFixity", true)
		defer p.sess.Emit()
		f = p.parseFile()
	}()
	require.True(p.sess.IsOK(), "no errors should be returned")
	require.NotNil(f)
}
