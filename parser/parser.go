package parser

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/diagnostic"
	"github.com/erizocosmico/elmo/operator"
	"github.com/erizocosmico/elmo/scanner"
	"github.com/erizocosmico/elmo/token"
)

type parser struct {
	sess       *Session
	scanner    *scanner.Scanner
	fileName   string
	unresolved []*ast.Ident
	mode       ParseMode

	tok     *token.Token
	errors  []error
	state   []state
	regions []*token.Position
}

func newParser(sess *Session) *parser {
	return &parser{sess: sess}
}

type bailout struct{}

func (p *parser) init(fileName string, s *scanner.Scanner, mode ParseMode) {
	p.scanner = s
	p.state = []state{{parsingFile, 1, 1}}
	p.fileName = fileName
	p.mode = mode

	p.next()
}

func parseFile(p *parser) *ast.File {
	mod := parseModule(p)
	imports := parseImports(p)
	if p.mode.Is(SkipDefinitions) {
		p.skipUntilNextFixity()
	}

	var decls []ast.Decl
	for p.tok.Type != token.EOF {
		decls = append(decls, parseDecl(p))
	}

	return &ast.File{
		Name:    p.fileName,
		Module:  mod,
		Imports: imports,
		Decls:   decls,
	}
}

func (p *parser) skipUntilNextFixity() {
	p.pushState(skipping, 0)
	for {
		switch p.tok.Type {
		case token.Infix, token.Infixr, token.Infixl, token.EOF:
			p.popState()
			return
		}
		p.next()
	}
}

func parseIdentifierOrOp(p *parser) *ast.Ident {
	if !p.is(token.LeftParen) {
		return parseIdentifier(p)
	}

	p.expect(token.LeftParen)
	op := parseOp(p)
	p.expect(token.RightParen)
	return op
}

func parseUpperName(p *parser) *ast.Ident {
	ident := parseIdentifier(p)
	if !isUpper(ident.Name) {
		p.errorMessage(ident.NamePos, "I was expecting an upper case name.")
	}
	return ident
}

func parseLowerName(p *parser) *ast.Ident {
	ident := parseIdentifier(p)
	if !isLower(ident.Name) {
		p.errorMessage(ident.NamePos, "I was expecting a lower case name.")
	}
	return ident
}

func parseOp(p *parser) *ast.Ident {
	name := "_"
	pos := p.tok.Position
	if p.tok.Type == token.Op {
		name = p.tok.Value
		p.next()
	} else {
		p.expect(token.Op)
	}

	return &ast.Ident{NamePos: pos, Name: name}
}

func (p *parser) checkState(state state) {
	switch state.mode {
	case parsingDecl:
		if p.atLineStart() && state.line != p.tok.Line {
			p.errorMessage(p.tok.Position, "I encountered what looks like a new declaration, but the previous one has not been finished yet.")
		}
	case parsingFile:
		if !p.atLineStart() {
			p.errorMessage(p.tok.Position, "I expected a new declaration. All declarations need to start at the beginning of their line.")
		}
	case parsingLet:
		if p.atLineStart() && !p.is(token.Let) && !p.is(token.In) {
			p.errorExpectedOneOf(p.tok, token.In)
		}
	case parsingCaseBranch:
		if p.tok.Column == 1 {
			p.errorMessage(p.tok.Position, "I encountered what looks like a new declaration, but the previous one has not been finished yet.")
		} else if p.tok.Column < state.lineStart {
			p.errorMessage(p.tok.Position, "I was expecting a new case branch, but the indentation does not match the one in the previous branch.")
		}
	}
}

func (p *parser) next() {
	if p.tok != nil {
		p.checkState(p.currentState())
	}

	p.tok = p.scanner.Next()
	if p.is(token.Comment) {
		// ignore comments for now
		p.next()
	}
}

func (p *parser) backup(until *token.Token) {
	p.scanner.Backup(until)
	p.next()
}

func (p *parser) peek() *token.Token {
	t := p.scanner.Peek()
	if t == nil {
		p.errorUnexpectedEOF()
	}
	return t
}

// finishedDecl marks as finished the parsing of the previous declaration. It
// generates an error otherwise.
func (p *parser) finishedDecl() {
	if p.is(token.EOF) || p.currentState().mode == parsingDecl {
		p.popState()
		return
	}

	p.errorMessage(p.tok.Position, "I was expecting a new declaration or the end of file, but I got %s instead.", p.tok.Type)
}

func (p *parser) expect(typ token.Type) token.Pos {
	pos := p.tok.Position
	if p.tok.Type != typ {
		p.errorExpected(p.tok, typ)
	}

	p.next()
	return pos.Offset
}

func (p *parser) expectAfter(typ token.Type, node ast.Node) token.Pos {
	pos := p.tok.Position
	if pos.Offset != node.End() {
		p.errorMessage(pos, "I was expecting %q right after the previous token, but I ran into whitespace.", typ)
	}
	return p.expect(typ)
}

func (p *parser) expectType() ast.Type {
	pos := p.tok.Position
	typ := parseType(p)
	if typ == nil {
		p.errorExpectedType(pos)
	}
	return typ
}

func (p *parser) expectOneOf(types ...token.Type) token.Pos {
	pos := p.tok.Position
	var found bool
	for _, t := range types {
		if p.tok.Type == t {
			found = true
		}
	}

	if !found {
		p.errorExpectedOneOf(p.tok, types...)
	}

	p.next()
	return pos.Offset
}

func (p *parser) is(typ token.Type) bool {
	return p.tok.Type == typ
}

func (p *parser) currentState() state {
	return p.state[len(p.state)-1]
}

func (p *parser) atLineStart() bool {
	return p.tok.Column <= p.currentState().lineStart
}

func (p *parser) opInfo(name string) *operator.OpInfo {
	info := p.sess.Table.Lookup(name, "" /* TODO: path */)
	if info != nil {
		return info
	}

	return &operator.OpInfo{
		Precedence:    0,
		Associativity: operator.Left,
	}
}

func (p *parser) startRegion() {
	p.regions = append(p.regions, p.tok.Position)
}

func (p *parser) endRegion() {
	p.regions = p.regions[:len(p.regions)-1]
}

func (p *parser) regionStart() *token.Position {
	if len(p.regions) == 0 {
		return &token.Position{Offset: token.NoPos, Line: 1}
	}
	return p.regions[len(p.regions)-1]
}

func (p *parser) region(start *token.Position) []string {
	region, err := p.sess.
		Source(p.fileName).
		Region(
			start.Offset,
			p.tok.Offset+token.Pos(len(p.tok.Source)),
		)
	if err != nil {
		panic(fmt.Errorf("unexpected error: %s", err))
	}

	return region
}

func (p *parser) regionError(pos *token.Position, msg diagnostic.Msg) {
	start := p.regionStart()
	p.sess.Diagnose(p.fileName, diagnostic.NewRegionDiagnostic(
		diagnostic.Error,
		msg,
		start,
		pos,
		p.region(start),
	))
}

func (p *parser) errorExpected(t *token.Token, typ token.Type) {
	if t.Type == token.EOF {
		p.regionError(t.Position, diagnostic.UnexpectedEOF(typ))
		panic(bailout{})
	}

	p.errorExpectedOneOf(t, typ)
}

func (p *parser) errorExpectedOneOf(t *token.Token, types ...token.Type) {
	p.regionError(t.Position, diagnostic.Expecting(t.Type, types...))
}

func (p *parser) errorMessage(pos *token.Position, msg string, args ...interface{}) {
	p.regionError(pos, diagnostic.ParseError(fmt.Sprintf(msg, args...)))
}

func (p *parser) errorUnexpectedEOF() {
	p.errorMessage(p.tok.Position, "Unexpected end of file.")
	panic(bailout{})
}

func (p *parser) errorExpectedType(pos *token.Position) {
	p.errorMessage(pos, "I was expecting a type, but I encountered what looks like a declaration instead.")
	panic(bailout{})
}

func isLower(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsLower(r)
}

func isUpper(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

type state struct {
	mode      stateMode
	lineStart int
	line      int
}

type stateMode uint

const (
	parsingFile stateMode = iota
	parsingDecl
	parsingCase
	parsingCaseBranch
	parsingLet
	skipping
)

func (m stateMode) String() string {
	switch m {
	case parsingFile:
		return "parsingFile"
	case parsingDecl:
		return "parsingDecl"
	case parsingCase:
		return "parsingCase"
	case parsingCaseBranch:
		return "parsingCaseBranch"
	case parsingLet:
		return "parsingLet"
	case skipping:
		return "skipping"
	default:
		return "invalid"
	}
}

func (p *parser) pushState(mode stateMode, lineStart int) {
	p.state = append(p.state, state{mode, lineStart, p.tok.Line})
}

func (p *parser) popState() {
	if len(p.state) <= 1 {
		p.state = nil
	} else {
		p.state = p.state[:len(p.state)-1]
	}
}
