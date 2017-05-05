package parser

import (
	"testing"

	"github.com/elm-tangram/tangram/ast"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	s := require.New(t)
	table := newOpTable()
	s.NoError(table.add("?", "foo", ast.Left, 0))
	s.Error(table.add("?", "foo", ast.Left, 0))
}
