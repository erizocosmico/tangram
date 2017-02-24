package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSelectorExpr(t *testing.T) {
	require := require.New(t)

	cases := []struct {
		name     string
		idents   []string
		isNil    bool
		expected string
	}{
		{"no idents", nil, true, ""},
		{"one ident", []string{"Foo"}, true, ""},
		{"two idents", []string{"foo", "bar"}, false, "foo.bar"},
		{"more than two", []string{"foo", "bar", "baz"}, false, "foo.bar.baz"},
	}

	for _, c := range cases {
		var idents []*Ident
		for _, i := range c.idents {
			idents = append(idents, &Ident{Name: i})
		}

		sel := NewSelectorExpr(idents...)
		if c.isNil {
			require.Nil(sel, c.name)
		} else {
			require.NotNil(sel, c.name)
			require.Equal(c.expected, sel.String())
		}
	}
}
