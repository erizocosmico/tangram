package source

import (
	"strings"
	"testing"

	"github.com/elm-tangram/tangram/token"

	"github.com/stretchr/testify/require"
)

const sourceFixture = `
map : (a -> b) -> List a -> List b
map f xs =
  foldr (\x acc -> f x :: acc) [] xs

foldr : (a -> b -> b) -> b -> List a -> b
foldr =
	Native.List.foldr
`

const expectedMap = `map : (a -> b) -> List a -> List b
map f xs =
  foldr (\x acc -> f x :: acc) [] xs`

const expectedType = `      (a -> b) -> List a -> List b`

const expectedWithTab = `    Native.List.foldr`

func TestSource(t *testing.T) {
	s, err := NewSource("foo", strings.NewReader(sourceFixture))
	require.NoError(t, err)

	t.Run("line index", func(t *testing.T) {
		require := require.New(t)

		cases := []lineInfo{
			{0, 1},
			{1, 36},
			{36, 47},
			{47, 84},
			{84, 85},
			{85, 127},
			{127, 135},
			{135, 154},
		}

		require.Len(s.lineIndex, len(cases))
		for i, c := range cases {
			require.Equal(c.start, s.lineIndex[i].start, "start of line %d", i+1)
			require.Equal(c.end, s.lineIndex[i].end, "end of line %d", i+1)
		}
	})

	t.Run("findLineStart", func(t *testing.T) {
		require := require.New(t)

		cases := []struct {
			pos   token.Pos
			start token.Pos
			line  int
		}{
			{20, 1, 2},
			{84, 84, 5},
			{100, 85, 6},
		}

		for _, c := range cases {
			start, line := s.findLineStart(c.pos)
			require.Equal(c.start, start, "start of offset %d", c.pos)
			require.Equal(c.line, line, "line of offset %d", c.pos)
		}
	})

	t.Run("LinePos", func(t *testing.T) {
		require := require.New(t)

		cases := []struct {
			pos  token.Pos
			col  int
			line int
		}{
			{3, 3, 2},
			{1, 1, 2},
			{50, 4, 4},
		}

		for _, c := range cases {
			p, err := s.LinePos(c.pos)
			require.NoError(err, "should not error, offset %d", c.pos)
			require.Equal(c.col, p.Col, "col of offset %d", c.pos)
			require.Equal(c.line, p.Line, "line of offset %d", c.pos)
		}
	})

	t.Run("Region", func(t *testing.T) {
		require := require.New(t)

		cases := []struct {
			name     string
			start    token.Pos
			end      token.Pos
			line     int
			expected string
		}{
			{"at start", 1, 84, 2, expectedMap},
			{"in the middle of the line", 7, 36, 2, expectedType},
			{"with tabs", 135, 154, 8, expectedWithTab},
		}

		for _, c := range cases {
			snippet, err := s.Region(c.start, c.end)
			require.NoError(err, c.name)

			lines := strings.Split(c.expected, "\n")
			require.Equal(lines, snippet.Lines, c.name)
			require.Equal(c.line, snippet.Start, c.name)
		}
	})
}
