package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	g := NewGraph("a").
		Add("b", "a").
		Add("c", "a").
		Add("e", "b").
		Add("d", "b").
		Add("d", "c").
		Add("f", "e").
		Add("g", "f").
		Add("g", "d")

	nodes, err := g.Resolve()
	require.NoError(t, err)

	expected := []string{"g", "f", "e", "d", "b", "c", "a"}
	require.Equal(t, expected, nodes)

	nodes2, err := g.Resolve()
	require.NoError(t, err)

	require.Exactly(t, nodes, nodes2)
}

func TestCircularDep(t *testing.T) {
	g := NewGraph("a").
		Add("b", "a").
		Add("c", "a").
		Add("e", "b").
		Add("d", "b").
		Add("d", "c").
		Add("f", "e").
		Add("b", "f").
		Add("g", "f").
		Add("g", "d")

	nodes, err := g.Resolve()
	require.Error(t, err)
	circular, ok := err.(*CircularDependencyError)
	require.True(t, ok, "expected a CircularDependencyError")
	require.Equal(t, [2]string{"f", "b"}, circular.Modules)
	require.Nil(t, nodes)
}
