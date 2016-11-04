package lexer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"

	"github.com/mvader/elm-compiler/token"
)

type stateFunc func(*Lexer) (stateFunc, error)

const (
	eof     = -1
	newLine = '\n'

	leftParen    = '('
	rightParen   = ')'
	leftBracket  = '['
	rightBracket = ']'
	leftBrace    = '{'
	rightBrace   = '}'
	quote        = '"'
	pipe         = '|'
	backslash    = '\\'
	singleQuote  = '\''
	backQuote    = '`'
	underscore   = '_'
	eq           = '='
	colon        = ':'
	comma        = ','
	dash         = '-'
	gt           = '>'
	dot          = '.'

	numDigits = "0123456789"
	hexDigits = "0123456789abcdefABCDEF"
)

// Lexer is in charge of extracting tokens from a source.
type Lexer struct {
	mut     sync.RWMutex
	source  string
	reader  *bufio.Reader
	state   stateFunc
	pos     int
	start   int
	width   int
	line    int
	linePos int
	tokens  chan *token.Token
	word    []rune
}

// New creates a new lexer for the input.
func New(source string, input io.Reader) *Lexer {
	return &Lexer{
		source: source,
		reader: bufio.NewReader(input),
		state:  lexExpr,
		tokens: make(chan *token.Token),
		line:   1,
	}
}

// next returns the next rune in the input or EOF if none left.
func (l *Lexer) next() (r rune, err error) {
	r, l.width, err = l.reader.ReadRune()
	l.pos += l.width
	l.linePos++
	if r != 0x0 {
		l.word = append(l.word, r)
	}
	return
}

// backup steps back to the latest consumed rune.
func (l *Lexer) backup() error {
	l.pos -= l.width
	l.linePos--

	if len(l.word) == 1 {
		l.word = nil
	} else if len(l.word) > 1 {
		l.word = l.word[0 : len(l.word)-1]
	}

	return l.reader.UnreadRune()
}

// peek returns the next rune without actually consuming it.
func (l *Lexer) peek() (r rune, err error) {
	r, err = l.next()
	if err != nil {
		return
	}

	err = l.backup()
	if err != nil {
		return
	}

	return r, nil
}

func (l *Lexer) peekWord() string {
	return string(l.word)
}

// emit sends the token to the consumer.
func (l *Lexer) emit(t token.Type) {
	word := l.peekWord()
	l.word = nil
	l.tokens <- token.New(
		t,
		l.source,
		l.start,
		l.linePos-len(word)+1,
		l.line,
		word,
	)
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	l.start = l.pos
	l.word = nil
}

// accept consumes a rune if it's from the valid set and reports if it was accepted or not.
func (l *Lexer) accept(valid string) (bool, error) {
	r, err := l.next()
	if err != nil {
		return false, err
	}

	if strings.IndexRune(valid, r) >= 0 {
		return true, nil
	}

	l.backup()
	return false, nil
}

// acceptRun consumes a run of runes from the valid set given.
func (l *Lexer) acceptRun(valid string) error {
	for {
		ok, err := l.accept(valid)
		if err != nil {
			return err
		}

		if !ok {
			return nil
		}
	}
}

// Run runs the state machine for the lexer in parallel.
func (l *Lexer) Run() {
	for {
		l.mut.Lock()
		var err error
		l.state, err = l.state(l)
		if err == io.EOF {
			l.emit(token.EOF)
			l.state = nil
		} else if err != nil {
			l.errorf("unexpected error: %s", err.Error())
			break
		}

		if l.state == nil {
			l.mut.Unlock()
			break
		}
		l.mut.Unlock()
	}
	close(l.tokens)
}

// Stop stops the lexer.
func (l *Lexer) Stop() {
	l.mut.Lock()
	defer l.mut.Unlock()
	l.state = nil
}

// newLine increments the line and sets the new line start
func (l *Lexer) newLine() {
	l.line++
	l.linePos = 0
}

// errorf emits an error token.
func (l *Lexer) errorf(format string, args ...interface{}) stateFunc {
	l.backup()
	l.ignore()
	l.next()
	l.tokens <- token.New(
		token.Error,
		l.source,
		l.start,
		l.linePos,
		l.line,
		fmt.Sprintf(format, args...),
	)
	return nil
}

// scanNumber scans a number and returns if the termination is valid
// can detect integers, floats and integer ranges
func (l *Lexer) scanNumber() (bool, token.Type, error) {
	var t = token.Int
	if err := l.acceptRun(numDigits); err != nil {
		return false, t, err
	}

	ok, err := l.accept(".")
	if err != nil {
		return false, t, err
	}

	if ok {
		bs, err := l.reader.Peek(1)
		if err != nil {
			return false, t, err
		}

		if rune(bs[0]) == dot {
			if err := l.backup(); err != nil {
				return false, t, err
			}

			return true, t, nil
		}
		t = token.Float
	}

	if ok {
		if err := l.acceptRun(numDigits); err != nil {
			return false, t, err
		}
	}

	r, err := l.peek()
	if err != nil {
		return false, t, err
	}

	if isAllowedInIdentifier(r) {
		return false, t, nil
	}

	return true, t, nil
}

// Next returns the next Token available in the lexer.
func (l *Lexer) Next() *token.Token {
	return <-l.tokens
}

// lexLeftParen scans the left paren, which is known to be present.
func lexLeftParen(l *Lexer) (stateFunc, error) {
	l.emit(token.LeftParen)
	return lexExpr, nil
}

// lexRightParen scans the right paren, which is known to be present.
func lexRightParen(l *Lexer) (stateFunc, error) {
	l.emit(token.RightParen)
	return lexExpr, nil
}

// lexLeftBracket scans the left bracket, which is known to be present.
func lexLeftBracket(l *Lexer) (stateFunc, error) {
	l.emit(token.LeftBracket)
	return lexExpr, nil
}

// lexRightBracket scans the right bracket, which is known to be present.
func lexRightBracket(l *Lexer) (stateFunc, error) {
	l.emit(token.RightBracket)
	return lexExpr, nil
}

// lexLeftBrace scans the left brace, which is known to be present.
func lexLeftBrace(l *Lexer) (stateFunc, error) {
	l.emit(token.LeftBrace)
	return lexExpr, nil
}

// lexRightBrace scans the right brace, which is known to be present.
func lexRightBrace(l *Lexer) (stateFunc, error) {
	l.emit(token.RightBrace)
	return lexExpr, nil
}

// lexExpr scans the elements inside an expression.
func lexExpr(l *Lexer) (stateFunc, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}

	switch true {
	case r == eof:
		l.emit(token.EOF)
		return nil, nil
	case isEOL(r):
		return lexEOL, nil
	case isSpace(r):
		return lexSpaces, nil
	case r == quote:
		return lexQuote, nil
	case isNumeric(r):
		return lexNumber, nil
	case r == leftParen:
		return lexLeftParen, nil
	case r == rightParen:
		return lexRightParen, nil
	case r == leftBracket:
		return lexLeftBracket, nil
	case r == rightBracket:
		return lexRightBracket, nil
	case r == leftBrace:
		nr, err := l.peek()
		if err != nil {
			return nil, err
		}

		if nr == dash {
			return lexMultiLineComment, nil
		}

		return lexLeftBrace, nil
	case r == rightBrace:
		return lexRightBrace, nil
	case r == singleQuote:
		return lexChar, nil
	case r == backQuote:
		return lexInfixOp, nil
	case r == colon:
		return lexColon, nil
	case r == eq:
		return lexEq, nil
	case r == comma:
		l.emit(token.Comma)
		return lexExpr, nil
	case r == pipe:
		l.emit(token.Pipe)
		return lexExpr, nil
	case r == dot:
		nr, err := l.next()
		if err != nil {
			return nil, err
		}

		if !isAllowedInOp(nr) {
			if err := l.backup(); err != nil {
				return nil, err
			}

			l.emit(token.Dot)
			return lexExpr, nil
		}

		if nr == dot {
			l.emit(token.Range)
			return lexExpr, nil
		}

		return lexOp, nil
	case r == dash:
		nr, err := l.next()
		if err != nil {
			return nil, err
		}

		if nr == dash {
			return lexComment, nil
		}

		if nr == gt {
			l.emit(token.Arrow)
			return lexExpr, nil
		}

		if err := l.backup(); err != nil {
			return nil, err
		}

		return lexOp, nil
	case isAllowedInOp(r):
		return lexOp, nil
	case isAllowedInIdentifier(r) && !isNumeric(r):
		return lexIdentifier, nil
	default:
		return l.errorf("invalid syntax: %q", l.peekWord()), nil
	}
}

func lexInfixOp(l *Lexer) (stateFunc, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}

	if !unicode.IsLetter(r) {
		return l.errorf("invalid start character for identifier: %c", r), nil
	}

	for {
		r, err := l.next()
		if err != nil && err != io.EOF {
			return nil, err
		}

		if !isAllowedInIdentifier(r) {
			if r != backQuote {
				return l.errorf("invalid infix operator: %q", l.peekWord()), nil
			}
			break
		}
	}

	l.emit(token.InfixOp)
	return lexExpr, nil
}

func lexOp(l *Lexer) (stateFunc, error) {
	for {
		r, err := l.next()
		if err != nil {
			return nil, err
		}

		if !isAllowedInOp(r) {
			l.backup()
			l.emit(token.Op)
			return lexExpr, nil
		}
	}
}

func lexColon(l *Lexer) (stateFunc, error) {
	r, err := l.peek()
	if err != nil {
		return nil, err
	}

	if isAllowedInOp(r) {
		return lexOp, nil
	}

	l.emit(token.Colon)
	return lexExpr, nil
}

func lexEq(l *Lexer) (stateFunc, error) {
	r, err := l.peek()
	if err != nil {
		return nil, err
	}

	if isAllowedInOp(r) {
		return lexOp, nil
	}

	l.emit(token.Assign)
	return lexExpr, nil
}

// lexChar scans for a character.
func lexChar(l *Lexer) (stateFunc, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}

	if r == eof {
		return l.errorf("not closed character: %q", l.peekWord()), nil
	} else if r == backslash {
		_, err := l.next()
		if err != nil {
			return nil, err
		}
	}

	ok, err := l.accept("'")
	if err != nil {
		return nil, err
	}

	if !ok {
		return l.errorf("not closed character: %q", l.peekWord()), nil
	}

	l.emit(token.Char)
	return lexExpr, nil
}

// lexEOL scans all end of lines.
func lexEOL(l *Lexer) (stateFunc, error) {
	l.newLine()
	for {
		r, err := l.next()
		if err != nil {
			return nil, err
		}

		if isEOL(r) {
			l.newLine()
		} else {
			l.backup()
			break
		}
	}

	l.ignore()
	return lexExpr, nil
}

// lexSpaces scanns a run of space chars.
func lexSpaces(l *Lexer) (stateFunc, error) {
	for {
		r, err := l.next()
		if err != nil {
			return nil, err
		}

		if !isSpace(r) {
			break
		}
	}

	l.backup()
	l.ignore()
	return lexExpr, nil
}

// lexNumbers scans a number int or float
func lexNumber(l *Lexer) (stateFunc, error) {
	ok, kind, err := l.scanNumber()
	if err == io.EOF {
		l.emit(kind)
		return nil, err
	} else if err != nil {
		return nil, err
	}

	if !ok {
		return l.errorf("bad number syntax: %q", l.peekWord()), nil
	}

	l.emit(kind)
	return lexExpr, nil
}

// lexComment scans a comment. The '--' delimiter has already been scanned.
func lexComment(l *Lexer) (stateFunc, error) {
	for {
		r, err := l.next()
		if err != nil {
			return nil, err
		}

		if isEOL(r) || r == eof {
			l.backup()
			l.emit(token.Comment)
			return lexExpr, nil
		}
	}
}

func lexMultiLineComment(l *Lexer) (stateFunc, error) {
	for {
		r, err := l.next()
		if err != nil {
			return nil, err
		}

		if r == dash {
			nr, err := l.next()
			if err != nil {
				return nil, err
			}

			if nr == rightBrace {
				l.emit(token.Comment)
				return lexExpr, nil
			}
		} else if isEOL(r) {
			l.newLine()
		}
	}
}

// lexInsideQuote scans the next rune and tells if there is an error
// or the scan needs to stop
func lexInsideQuote(l *Lexer) (bool, error) {
	r, err := l.next()
	if err != nil {
		return false, err
	}

	switch true {
	case r == '\\':
		rn, err := l.next()
		if err != nil {
			return false, err
		}

		if rn != eof && !isEOL(rn) {
			return false, nil
		}
		fallthrough
	case r == eof:
		return false, io.EOF
	case r == quote:
		return true, nil
	}
	return false, nil
}

// lexQuote scans a quoted string. The first quote has already been scanned.
func lexQuote(l *Lexer) (stateFunc, error) {
	for {
		stop, err := lexInsideQuote(l)
		if err != nil {
			return l.errorf("quoted string not closed properly: %q", l.peekWord()), nil
		}

		if stop {
			break
		}
	}

	l.emit(token.String)
	return lexExpr, nil
}

// lexIdentifier scans an identifier. First character is already scanned.
func lexIdentifier(l *Lexer) (stateFunc, error) {
	for {
		r, err := l.next()
		if err != nil && err != io.EOF {
			return nil, err
		}

		if !isAllowedInIdentifier(r) {
			if r != 0x0 {
				l.backup()
			}
			word := l.peekWord()

			if typ, ok := isKeyword(word); ok {
				l.emit(typ)
			} else {
				l.emit(token.Identifier)
			}

			return lexExpr, nil
		}
	}
}

// isSpace reports if the rune is a space or a tab.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEOL reports if the rune is an end of line character.
func isEOL(r rune) bool {
	return r == '\r' || r == '\n'
}

// isNumeric reports if the rune is a number.
func isNumeric(r rune) bool {
	return r >= '0' && r <= '9'
}

// isAllowedInIdentifier reports if the rune is allowed in an identifier.
func isAllowedInIdentifier(r rune) bool {
	return isAlphanumeric(r) || r == underscore || r == singleQuote
}

// isAlphanumeric reports if the rune is a letter or a digit
func isAlphanumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isAllowedInOp(r rune) bool {
	return !isAlphanumeric(r) && !unicode.IsSpace(r) && strings.IndexRune("[]{}()`\"|,", r) < 0
}

var keywords = map[string]token.Type{
	"type":     token.TypeDef,
	"as":       token.As,
	"alias":    token.Alias,
	"if":       token.If,
	"then":     token.Then,
	"else":     token.Else,
	"of":       token.Of,
	"case":     token.Case,
	"infixl":   token.Infixl,
	"infixr":   token.Infixr,
	"let":      token.Let,
	"in":       token.In,
	"module":   token.Module,
	"exposing": token.Exposing,
	"import":   token.Import,
	"True":     token.True,
	"False":    token.False,
}

func isKeyword(lit string) (token.Type, bool) {
	typ, ok := keywords[lit]
	return typ, ok
}
