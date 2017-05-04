package scanner

import (
	"strings"
	"testing"

	"github.com/elm-tangram/tangram/token"
	"github.com/stretchr/testify/require"
)

func TestIsAllowedInIdentifier(t *testing.T) {
	allowed := "abc135fdcv_"
	notAllowed := ":.;,{}[]`|#%\\-+-?!&=/<>'^$"
	for _, r := range allowed {
		require.Equal(t, isAllowedInIdentifier(r), true)
	}

	for _, r := range notAllowed {
		require.Equal(t, isAllowedInIdentifier(r), false)
	}
}

func TestLexNumber(t *testing.T) {
	require := require.New(t)

	testLexState(t, "24 ", lexNumber, func(l *Scanner, tokens []*token.Token) {
		require.Equal(1, len(tokens))
		require.Equal(token.Int, tokens[0].Type)
		require.Equal("24", tokens[0].Value)
	})

	testLexState(t, "24.56 ", lexNumber, func(l *Scanner, tokens []*token.Token) {
		require.Equal(1, len(tokens))
		require.Equal(token.Float, tokens[0].Type)
		require.Equal("24.56", tokens[0].Value)
	})
}

func TestLexString(t *testing.T) {
	require := require.New(t)

	// note: first quote is already scanned
	testLexState(t, `foo bar\"baz\"" `, lexString, func(l *Scanner, tokens []*token.Token) {
		require.Equal(1, len(tokens))
		require.Equal(token.String, tokens[0].Type)
		// is missing the first quote because we did not actually scan it
		require.Equal(`foo bar\"baz\""`, tokens[0].Value)
	})

	testLexState(t, `""foo
	"" slfjlkdfj
	jfdlskfj " " "
	""" `, lexString, func(l *Scanner, tokens []*token.Token) {
		require.Equal(1, len(tokens))
		require.Equal(token.String, tokens[0].Type)
		require.Equal(`""foo
	"" slfjlkdfj
	jfdlskfj " " "
	"""`, tokens[0].Value)
	})
}

const testRecord = `
type alias Foo = 
	{ myInt : Int 
	, myFloat : Float
	}
`

func TestLexRecord(t *testing.T) {
	testLex(t, testRecord, []expectedToken{
		{"type", token.TypeDef},
		{"alias", token.Alias},
		{"Foo", token.Identifier},
		{"=", token.Assign},
		{"{", token.LeftBrace},
		{"myInt", token.Identifier},
		{":", token.Colon},
		{"Int", token.Identifier},
		{",", token.Comma},
		{"myFloat", token.Identifier},
		{":", token.Colon},
		{"Float", token.Identifier},
		{"}", token.RightBrace},
		{"\n", token.EOF},
	})
}

const textFuncDecl = `
foo : (Int -> Int) -> Int -> Int
foo fn n =
	fn n
`

func TestLexFuncDecl(t *testing.T) {
	testLex(t, textFuncDecl, []expectedToken{
		{"foo", token.Identifier},
		{":", token.Colon},
		{"(", token.LeftParen},
		{"Int", token.Identifier},
		{"->", token.Arrow},
		{"Int", token.Identifier},
		{")", token.RightParen},
		{"->", token.Arrow},
		{"Int", token.Identifier},
		{"->", token.Arrow},
		{"Int", token.Identifier},
		{"foo", token.Identifier},
		{"fn", token.Identifier},
		{"n", token.Identifier},
		{"=", token.Assign},
		{"fn", token.Identifier},
		{"n", token.Identifier},
		{"\n", token.EOF},
	})
}

const testRecordUpdate = `
{ model | foo = True }
`

func TestLexRecordUpdate(t *testing.T) {
	testLex(t, testRecordUpdate, []expectedToken{
		{"{", token.LeftBrace},
		{"model", token.Identifier},
		{"|", token.Pipe},
		{"foo", token.Identifier},
		{"=", token.Assign},
		{"True", token.True},
		{"}", token.RightBrace},
		{"\n", token.EOF},
	})
}

const testSumType = `
type Op
	= Sum
	| Div
	| Mul
	| Sub
`

func TestSumType(t *testing.T) {
	testLex(t, testSumType, []expectedToken{
		{"type", token.TypeDef},
		{"Op", token.Identifier},
		{"=", token.Assign},
		{"Sum", token.Identifier},
		{"|", token.Pipe},
		{"Div", token.Identifier},
		{"|", token.Pipe},
		{"Mul", token.Identifier},
		{"|", token.Pipe},
		{"Sub", token.Identifier},
		{"\n", token.EOF},
	})
}

const testString = `
tom = { name = "Tom", bar = "\t\"" }
`

func TestString(t *testing.T) {
	testLex(t, testString, []expectedToken{
		{"tom", token.Identifier},
		{"=", token.Assign},
		{"{", token.LeftBrace},
		{"name", token.Identifier},
		{"=", token.Assign},
		{`"Tom"`, token.String},
		{",", token.Comma},
		{"bar", token.Identifier},
		{"=", token.Assign},
		{`"\t\""`, token.String},
		{"}", token.RightBrace},
		{"\n", token.EOF},
	})
}

const testChar = `
tom = { initial = 'T', foo = '\\' }
`

func TestChar(t *testing.T) {
	testLex(t, testChar, []expectedToken{
		{"tom", token.Identifier},
		{"=", token.Assign},
		{"{", token.LeftBrace},
		{"initial", token.Identifier},
		{"=", token.Assign},
		{`'T'`, token.Char},
		{",", token.Comma},
		{"foo", token.Identifier},
		{"=", token.Assign},
		{`'\\'`, token.Char},
		{"}", token.RightBrace},
		{"\n", token.EOF},
	})
}

const testComment = `
-- comment
-- other comment
`

func TestComment(t *testing.T) {
	testLex(t, testComment, []expectedToken{
		{"-- comment", token.Comment},
		{"-- other comment", token.Comment},
		{"\n", token.EOF},
	})
}

const testMultiLineComment = `
{-|-}
{-| Extract the first element of a list.
    head [1,2,3] == Just 1
    head [] == Nothing
-}
`

func TestMultiLineComment(t *testing.T) {
	testLex(t, testMultiLineComment, []expectedToken{
		{"{-|-}", token.Comment},
		{`{-| Extract the first element of a list.
    head [1,2,3] == Just 1
    head [] == Nothing
-}`, token.Comment},
		{"\n", token.EOF},
	})
}

const testInfixOp = "theMax = 3 `max` 5"

func TestInfixOp(t *testing.T) {
	testLex(t, testInfixOp, []expectedToken{
		{"theMax", token.Identifier},
		{"=", token.Assign},
		{"3", token.Int},
		{"invalid syntax: \"`\"", token.Error},
	})
}

const testList = `
List.map fn [1, 2, 3]
`

func TestList(t *testing.T) {
	testLex(t, testList, []expectedToken{
		{"List", token.Identifier},
		{".", token.Dot},
		{"map", token.Identifier},
		{"fn", token.Identifier},
		{"[", token.LeftBracket},
		{"1", token.Int},
		{",", token.Comma},
		{"2", token.Int},
		{",", token.Comma},
		{"3", token.Int},
		{"]", token.RightBracket},
		{"\n", token.EOF},
	})
}

const testOp = `
a = [1] ++ [2]
`

func TestLexOp(t *testing.T) {
	testLex(t, testOp, []expectedToken{
		{"a", token.Identifier},
		{"=", token.Assign},
		{"[", token.LeftBracket},
		{"1", token.Int},
		{"]", token.RightBracket},
		{"++", token.Op},
		{"[", token.LeftBracket},
		{"2", token.Int},
		{"]", token.RightBracket},
		{"\n", token.EOF},
	})
}

const testUnclosedString = `
foo = "unclosed
`

func TestLexUnclosedString(t *testing.T) {
	testLex(t, testUnclosedString, []expectedToken{
		{"foo", token.Identifier},
		{"=", token.Assign},
		{"", token.Error},
	})
}

const testUnclosedChar = `
foo = 'a
`

func TestLexUnclosedChar(t *testing.T) {
	testLex(t, testUnclosedChar, []expectedToken{
		{"foo", token.Identifier},
		{"=", token.Assign},
		{"", token.Error},
	})
}

const testBadNumber = `
foo = 12a4
`

func TestLexBadNumber(t *testing.T) {
	testLex(t, testBadNumber, []expectedToken{
		{"foo", token.Identifier},
		{"=", token.Assign},
		{"", token.Error},
	})
}

const testCustomOp = `
foo = 12 -: 13
`

func TestLexCustomOp(t *testing.T) {
	testLex(t, testCustomOp, []expectedToken{
		{"foo", token.Identifier},
		{"=", token.Assign},
		{"12", token.Int},
		{"-:", token.Op},
		{"13", token.Int},
		{"\n", token.EOF},
	})
}

func TestBackup(t *testing.T) {
	l := New("test", strings.NewReader(testSumType))
	l.Run()

	cases := []struct {
		breakpoint int
		advance    int
		expected   string
	}{
		{0, 30, "type"},
		{30, 30, "type"},
		{3, 2, "="},
	}

	for i, c := range cases {
		l.Reset()

		for j := 0; j < c.breakpoint-1; j++ {
			l.Next()
		}
		bp := l.Next()

		for j := 0; j < c.advance; j++ {
			l.Next()
		}

		l.Backup(bp)
		tok := l.Next()
		require.NotNil(t, tok, "expected token not to be nil, i=%d", i)
		require.Equal(t, c.expected, tok.Value, "i=%d", i)
	}
}

type expectedToken struct {
	value string
	typ   token.Type
}

func testLex(t *testing.T, input string, expected []expectedToken) {
	l := New("test", strings.NewReader(input))
	l.Run()
	tokens := l.tokens

	require.Equal(t, len(expected), len(tokens))
	for i := range tokens {
		require.Equal(t, expected[i].typ, tokens[i].Type)
		if tokens[i].Type != token.Error {
			require.Equal(t, expected[i].value, tokens[i].Value)
		}
	}
}

func testLexState(t *testing.T, input string, fn stateFunc, testFn func(*Scanner, []*token.Token)) {
	l := New("test", strings.NewReader(input))

	var err error
	l.state, err = fn(l)
	tokens := l.tokens

	if err != nil {
		t.Fatal(err)
	}
	testFn(l, tokens)
}
