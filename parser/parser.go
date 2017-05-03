package parser

import (
	"fmt"
	"path/filepath"
	"unicode"
	"unicode/utf8"

	"github.com/elm-tangram/tangram/ast"
	"github.com/elm-tangram/tangram/operator"
	"github.com/elm-tangram/tangram/report"
	"github.com/elm-tangram/tangram/scanner"
	"github.com/elm-tangram/tangram/token"
)

// defaultPos is a placeholder for a position of non-existent nodes in the
// source code.
var defaultPos = &token.Position{
	Source: "builtin",
	Offset: token.NoPos,
}

// defaultImports are the default imports included in every single Elm file.
var defaultImports = []*ast.ImportDecl{
	// import Basics exposing (..)
	&ast.ImportDecl{
		Module:   ast.NewIdent("Basics", defaultPos),
		Exposing: new(ast.OpenList),
	},
	// import List exposing ( (::) )
	&ast.ImportDecl{
		Module: ast.NewIdent("List", defaultPos),
		Exposing: &ast.ClosedList{
			Exposed: []ast.ExposedIdent{
				&ast.ExposedVar{ast.NewIdent("::", defaultPos)},
			},
		},
	},
	// import Maybe exposing ( Maybe( Just, Nothing ) )
	&ast.ImportDecl{
		Module: ast.NewIdent("Maybe", defaultPos),
		Exposing: &ast.ClosedList{
			Exposed: []ast.ExposedIdent{
				&ast.ExposedUnion{
					Type: ast.NewIdent("Maybe", defaultPos),
					Ctors: &ast.ClosedList{
						Exposed: []ast.ExposedIdent{
							&ast.ExposedVar{ast.NewIdent("Just", defaultPos)},
							&ast.ExposedVar{ast.NewIdent("Nothing", defaultPos)},
						},
					},
				},
			},
		},
	},
	// import Result exposing ( Result( Ok, Err ) )
	&ast.ImportDecl{
		Module: ast.NewIdent("Result", defaultPos),
		Exposing: &ast.ClosedList{
			Exposed: []ast.ExposedIdent{
				&ast.ExposedUnion{
					Type: ast.NewIdent("Result", defaultPos),
					Ctors: &ast.ClosedList{
						Exposed: []ast.ExposedIdent{
							&ast.ExposedVar{ast.NewIdent("Ok", defaultPos)},
							&ast.ExposedVar{ast.NewIdent("Err", defaultPos)},
						},
					},
				},
			},
		},
	},
	// import String
	&ast.ImportDecl{
		Module: ast.NewIdent("String", defaultPos),
	},
	// import Tuple
	&ast.ImportDecl{
		Module: ast.NewIdent("Tuple", defaultPos),
	},
	// import Debug
	&ast.ImportDecl{
		Module: ast.NewIdent("Debug", defaultPos),
	},
}

type parser struct {
	sess     *Session
	scanner  *scanner.Scanner
	fileName string
	mode     ParseMode

	// tok is the current token.
	tok *token.Token
	// region is the current region start for error reporting.
	region *token.Position
	// indent is the last indentation point.
	indent int
	// indentLine is the line in which the indentation was set.
	indentLine int
	// currentLine is the current line number.
	currentLine int
	// currentIndent is the indentation of the current line.
	currentIndent int
	// expectIndent is a flag that indicates that next tokens are expected to
	// have an indentation greater than `indent`.
	expectIndented bool
	// silent ignores all errors if true. This is only used to
	// avoid errors when there might be a backup parsing or when we're skipping
	// tokens.
	silent bool
}

func newParser(sess *Session) *parser {
	return &parser{sess: sess}
}

// bailout is the type used to stop parsing. It's the only panic
// that will be caught because it means we stopped parsing deliberately.
type bailout struct{}

func (p *parser) init(fileName string, s *scanner.Scanner, mode ParseMode) {
	p.scanner = s
	p.fileName = fileName
	p.mode = mode
	p.indent = 1
	p.indentLine = 1
	p.currentLine = -1
	p.currentIndent = 1
	p.silent = false
	p.expectIndented = false

	p.next()
}

func parseFile(p *parser) *ast.Module {
	mod := parseModule(p)
	var imports []*ast.ImportDecl
	if p.needsDefaultImports() {
		imports = defaultImports
	}

	imports = append(imports, parseImports(p)...)
	if p.mode.Is(SkipDefinitions) {
		p.skipUntilNextFixity()
	}

	var decls []ast.Decl
	for p.tok.Type != token.EOF {
		decls = append(decls, parseDecl(p))
	}

	return &ast.Module{
		Path:    p.fileName,
		Name:    mod.ModuleName(),
		Module:  mod,
		Imports: imports,
		Decls:   decls,
	}
}

func (p *parser) skipUntilNextFixity() {
	p.silent = true
	for {
		switch p.tok.Type {
		case token.Infix, token.Infixr, token.Infixl, token.EOF:
			p.silent = false
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

func (p *parser) indentedBlock() func() {
	return p.indentedBlockAt(p.currentIndent, p.currentLine)
}

func (p *parser) indentedBlockAt(indent, line int) func() {
	prevIndent := p.indent
	prevLine := p.indentLine
	expectingIndented := p.expectIndented

	p.indent = indent
	p.indentLine = line
	p.expectIndented = true

	return func() {
		p.indent = prevIndent
		p.indentLine = prevLine
		p.expectIndented = expectingIndented
	}
}

func (p *parser) currentPos() (indent, line int) {
	return p.currentIndent, p.currentLine
}

func (p *parser) next() {
	if p.tok != nil && !p.is(token.EOF) {
		if p.expectIndented && p.indentLine != p.currentLine {
			if p.tok.Column == 1 {
				p.errorMessage(p.tok.Position, "I encountered what looks like a new declaration, but the previous one has not been finished yet.")
			} else if p.currentIndent <= p.indent {
				p.errorMessage(p.tok.Position, "I was expecting whitespace.")
			}
		}
	}

	p.tok = p.scanner.Next()
	if p.is(token.Comment) {
		// ignore comments for now
		p.next()
	}

	if p.tok.Line != p.currentLine {
		p.currentIndent = p.tok.Column
		p.currentLine = p.tok.Line
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

func (p *parser) needsDefaultImports() bool {
	pkg, err := pkg.Load(filepath.Dir(p.fileName))
	if err != nil {
		return false
	}

	_, ok := specialPackages[pkg.Repository]
	return !ok
}

var specialPackages = map[string]struct{}{
	"https://github.com/elm-lang/core.git":    struct{}{},
	"http://github.com/elm-lang/core.git":     struct{}{},
	"https://github.com/elm-tangram/core.git": struct{}{},
	"http://github.com/elm-tangram/core.git":  struct{}{},
}

func (p *parser) checkAligned() bool {
	return p.indent == p.currentIndent
}

func (p *parser) isCorrectlyIndented() bool {
	return (p.currentIndent > p.indent &&
		p.currentLine > p.indentLine) ||
		p.currentLine == p.indentLine
}

func (p *parser) startRegion() (prev *token.Position) {
	prev = p.region
	p.region = p.tok.Position
	return prev
}

func (p *parser) endRegion(prev *token.Position) {
	p.region = prev
}

func (p *parser) regionStart() *token.Position {
	if p.region == nil {
		return &token.Position{Offset: token.NoPos, Line: 1}
	}
	return p.region
}

func (p *parser) errorExpected(t *token.Token, typ token.Type) {
	if t.Type == token.EOF {
		p.errorUnexpectedEOF()
		panic(bailout{})
	}

	p.errorExpectedOneOf(t, typ)
}

func (p *parser) errorExpectedOneOf(t *token.Token, types ...token.Type) {
	p.report(report.NewUnexpectedTokenError(t, p.currentRegion(), types...))
}

func (p *parser) errorUnexpectedEOF() {
	p.report(report.NewUnexpectedEOFError(p.tok.Offset, p.currentRegion()))
	panic(bailout{})
}

func (p *parser) errorExpectedType(pos *token.Position) {
	p.report(report.NewExpectedTypeError(pos.Offset, p.currentRegion()))
	panic(bailout{})
}

func (p *parser) errorMessage(pos *token.Position, msg string, args ...interface{}) {
	p.report(report.NewBaseReport(report.SyntaxError, pos.Offset, fmt.Sprintf(msg, args...), p.currentRegion()))
}

func (p *parser) currentRegion() *report.Region {
	start := p.regionStart()
	return &report.Region{start.Offset, p.tok.Offset + token.Pos(len(p.tok.Value))}
}

func (p *parser) report(report report.Report) {
	if p.silent {
		return
	}

	p.sess.Report(p.fileName, report)
}

func isLower(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsLower(r)
}

func isUpper(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}
