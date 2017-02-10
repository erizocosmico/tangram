package operator

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TableSuite struct {
	suite.Suite
	t *Table
}

func (s *TableSuite) SetupTest() {
	s.t = NewTable()
}

func (s *TableSuite) TestAdd() {
	s.NoError(s.t.Add("?", "foo", Left, 0))
	s.Error(s.t.Add("?", "foo", Left, 0))

	s.NoError(s.t.Add("?", "bar", Left, 0))
	s.NoError(s.t.AddBuiltin("+", Left, 0))
	s.Error(s.t.Add("+", "baz", Left, 0))
}

func (s *TableSuite) TestIsBuiltin() {
	s.NoError(s.t.Add("?", "foo", Left, 0))
	s.t.AddBuiltin("+", Left, 0)
	s.False(s.t.IsBuiltin("?"))
	s.True(s.t.IsBuiltin("+"))
}

func TestTable(t *testing.T) {
	suite.Run(t, new(TableSuite))
}
