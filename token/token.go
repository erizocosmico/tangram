package token

// Token is the smallest part in which the code can be divided and still makes sense on its own.
type Token struct {
	Type  Type
	Value string
	*Position
}

type Pos int

const NoPos Pos = 0

// Position represents the position of the token in a file.
type Position struct {
	Source string
	Offset Pos
	Line   int
	Column int
}

// New creates a new token of type t with start, line, value and position in line.
func New(t Type, source string, start, linePos, line int, val string) *Token {
	return &Token{
		Type:  t,
		Value: val,
		Position: &Position{
			Source: source,
			Offset: Pos(start),
			Line:   line,
			Column: linePos,
		},
	}
}

// Type is the type of an item.
type Type uint

const (
	// Error is an error occurred in the process of lexing, value is the text of the error
	Error Type = iota
	// EOF is the end of the input
	EOF
	// Comment is an user comment
	Comment
	// LeftParen is the left parenthesis "("
	LeftParen
	// RightParen is the right parenthesis ")"
	RightParen
	// LeftBracket is the left bracket "["
	LeftBracket
	// RightBracket is the right bracket "]"
	RightBracket
	// LeftBrace is the left brace "{"
	LeftBrace
	// RightBrace is the right brace "}"
	RightBrace
	// Pipe is the pipe character "|"
	Pipe
	// InfixOp is an identifier between backticks that acts as an infix op
	InfixOp
	// Colon is the colon character ":"
	Colon
	// Assign is the equal character "="
	Assign
	// Comma is the comma character ","
	Comma
	// Arrow is the arrow operator "->"
	Arrow
	// Identifier is an identifier (user defined vars, functions, predefined, ...)
	Identifier
	// Op is an operator
	Op
	// String is a quoted string literal
	String
	// Int is an integer number
	Int
	// Float is a floating point number
	Float
	// Range is a range of integers
	Range
	// Char is a quoted character literal
	Char
	// Dot is the dot character "."
	Dot

	// Keywords

	True
	False
	TypeDef
	As
	Alias
	If
	Then
	Else
	Of
	Case
	Infixl
	Infixr
	Let
	In
	Module
	Exposing
	Import
)
