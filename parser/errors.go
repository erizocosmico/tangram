package parser

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/mvader/elm-compiler/token"
)

type parseError struct {
	subError error
}

func (e *parseError) Error() string {
	return fmt.Sprintf(
		"I ran into something unexpected when parsing your code!\n%s",
		e.subError.Error(),
	)
}

type expectedError struct {
	ctx       []*token.Token
	pos       *token.Position
	expecting []token.Type
}

func (e *expectedError) Error() string {
	return fmt.Sprintf(
		"\nFile: %s%s\n\nI was expecting %s instead",
		e.pos.Source,
		generateErroredChunk(e.ctx, e.pos),
		joinExpectedTypes(e.expecting),
	)
}

type unexpectedEOFError struct {
	ctx       []*token.Token
	pos       *token.Position
	expecting []token.Type
}

func (e *unexpectedEOFError) Error() string {
	return fmt.Sprintf(
		"\nFile: %s%s\n\nUnexpected EOF, I was expecting %s instead",
		e.pos.Source,
		generateErroredChunk(e.ctx, e.pos),
		joinExpectedTypes(e.expecting),
	)
}

func joinExpectedTypes(types []token.Type) string {
	var buf bytes.Buffer
	ln := len(types)
	for i := 0; i < ln; i++ {
		buf.WriteString(color.GreenString(types[i].String()))
		if ln > 1 && i < ln-2 {
			buf.WriteString(", ")
		} else if i == ln-2 {
			buf.WriteString(" or ")
		}
	}
	return buf.String()
}

func generateErroredChunk(tokens []*token.Token, pos *token.Position) string {
	var (
		buf            bytes.Buffer
		prevLine, lpos int
		lastLineDigits = len(fmt.Sprint(tokens[len(tokens)-1].Line))
		lineFormat     = "\n" + `%-` + fmt.Sprint(lastLineDigits) + "d| "
	)

	for _, t := range tokens {
		if prevLine != t.Line {
			prevLine = t.Line
			lpos = 1
			buf.WriteString(fmt.Sprintf(lineFormat, t.Line))
		}

		if t.Column != lpos {
			fillBlanks(&buf, lpos, t.Column)
			lpos = t.Column
		}

		if t.Column == pos.Column && t.Line == pos.Line {
			buf.WriteString(color.RedString(t.Value))
		} else {
			buf.WriteString(t.Value)
		}

		lpos += utf8.RuneCountInString(t.Value)
	}
	return buf.String()
}

func fillBlanks(buf io.Writer, pos, offset int) {
	for pos < offset {
		buf.Write([]byte(" "))
		pos++
	}
}
