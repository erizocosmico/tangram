package operator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	s := require.New(t)
	table := NewTable()
	s.NoError(table.Add("?", "foo", Left, 0))
	s.Error(table.Add("?", "foo", Left, 0))
}
