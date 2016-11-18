package parser

import (
	"fmt"

	"github.com/mvader/elm-compiler/ast"
	"github.com/mvader/elm-compiler/scanner"
	"github.com/mvader/elm-compiler/token"
)

type parser struct {
	scanner    *scanner.Scanner
	fileName   string
	unresolved []*ast.Ident

	tok    *token.Token
	ctx    []*token.Token
	errors []error
}

type bailout struct{}

func (p *parser) init(fileName string, s *scanner.Scanner) {
	p.scanner = s
	p.fileName = fileName

	p.next()
}

func (p *parser) parseFile() *ast.File {
	var (
		imports []*ast.ImportDecl
		decls   []ast.Decl
		mod     = p.parseModule()
	)

	for {
		p.resetCtx()
		if p.is(token.EOF) {
			break
		}

		switch p.tok.Type {
		case token.Import:
			imports = append(imports, p.parseImport())

		case token.TypeDef:
			panic("type parsing is not implemented yet")

		case token.Identifier:
			panic("value parsing is not implemented yet")

		default:
			p.errorExpectedOneOf(p.tok, token.Import, token.TypeDef, token.Identifier)
		}
	}

	return &ast.File{
		Name:    p.fileName,
		Module:  mod,
		Imports: imports,
		Decls:   decls,
	}
}

func (p *parser) parseModule() *ast.ModuleDecl {
	var decl = new(ast.ModuleDecl)
	decl.Module = p.expect(token.Module)
	decl.Name = p.parseModuleName()

	if p.is(token.Exposing) {
		exposedList := new(ast.ExposingList)
		p.expect(token.Exposing)
		exposedList.Lparen = p.expect(token.LeftParen)
		exposedList.Idents = p.parseExposedIdents()
		if len(exposedList.Idents) == 0 {
			p.errorExpectedOneOf(p.tok, token.Range, token.Identifier)
		}
		exposedList.Rparen = p.expect(token.RightParen)
		decl.Exposing = exposedList
	}

	return decl
}

func (p *parser) parseImport() *ast.ImportDecl {
	var decl = new(ast.ImportDecl)
	decl.Import = p.expect(token.Import)
	decl.Module = p.parseModuleName()

	if p.is(token.As) {
		p.expect(token.As)
		decl.Alias = p.parseIdentifier()
	}

	if p.is(token.Exposing) {
		exposedList := new(ast.ExposingList)
		p.expect(token.Exposing)
		exposedList.Lparen = p.expect(token.LeftParen)
		exposedList.Idents = p.parseExposedIdents()
		if len(exposedList.Idents) == 0 {
			p.errorExpectedOneOf(p.tok, token.Range, token.Identifier)
		}
		exposedList.Rparen = p.expect(token.RightParen)
		decl.Exposing = exposedList
	}

	return decl
}

func (p *parser) parseModuleName() ast.ModuleName {
	name := ast.ModuleName{p.parseIdentifier()}

	for {
		if !p.is(token.Dot) {
			break
		}

		p.expect(token.Dot)
		name = append(name, p.parseIdentifier())
	}

	return name
}

func (p *parser) parseExposedIdents() []*ast.ExposedIdent {
	if p.is(token.Range) {
		p.expect(token.Range)
		return []*ast.ExposedIdent{
			&ast.ExposedIdent{
				Ident: &ast.Ident{Name: token.Range.String(), NamePos: p.tok.Position},
			},
		}
	}

	if !p.is(token.LeftParen) && !p.is(token.Identifier) {
		return nil
	}

	exposed := []*ast.ExposedIdent{p.parseExposedIdent()}
	for {
		if !p.is(token.Comma) {
			break
		}

		p.expect(token.Comma)
		exposed = append(exposed, p.parseExposedIdent())
	}

	return exposed
}

func (p *parser) parseExposedIdent() *ast.ExposedIdent {
	ident := &ast.ExposedIdent{Ident: p.parseIdentifierOrOp()}

	if p.is(token.LeftParen) {
		var exposingList = new(ast.ExposingList)
		exposingList.Lparen = p.expect(token.LeftParen)
		exposingList.Idents = p.parseExposedIdents()
		if len(exposingList.Idents) == 0 {
			p.errorExpectedOneOf(p.tok, token.Range, token.Identifier)
		}
		exposingList.Rparen = p.expect(token.RightParen)
		ident.Exposing = exposingList
	}

	return ident
}

func (p *parser) parseIdentifierOrOp() *ast.Ident {
	if !p.is(token.LeftParen) {
		return p.parseIdentifier()
	}

	p.expect(token.LeftParen)
	defer p.expect(token.RightParen)
	return p.parseOp()
}

func (p *parser) parseIdentifier() *ast.Ident {
	name := "_"
	pos := p.tok.Position
	if p.tok.Type == token.Identifier {
		name = p.tok.Value
		p.next()
	} else {
		p.expect(token.Identifier)
	}

	return &ast.Ident{NamePos: pos, Name: name}
}

func (p *parser) parseOp() *ast.Ident {
	name := "_"
	pos := p.tok.Position
	var obj *ast.Object
	if p.tok.Type == token.Op {
		name = p.tok.Value
		obj = &ast.Object{Kind: ast.Op}
		p.next()
	} else {
		p.expect(token.Op)
	}

	return &ast.Ident{NamePos: pos, Name: name, Obj: obj}
}

func (p *parser) next() {
	p.tok = p.scanner.Next()
	if p.is(token.Comment) {
		// ignore comments for now
		p.next()
	} else {
		p.ctx = append(p.ctx, p.tok)
	}
}

func (p *parser) resetCtx() {
	if len(p.ctx) > 0 {
		p.ctx = []*token.Token{p.ctx[len(p.ctx)-1]}
	}
}

func (p *parser) expect(typ token.Type) token.Pos {
	pos := p.tok.Position
	if p.tok.Type != typ {
		p.errorExpected(p.tok, typ)
	}

	p.next()
	return pos.Offset
}

func (p *parser) is(typ token.Type) bool {
	return p.tok.Type == typ
}

func (p *parser) errorExpected(t *token.Token, typ token.Type) {
	if t.Type == token.EOF {
		p.errors = append(p.errors, fmt.Errorf(
			"%s:%d:%d unexpected EOF, expecting %s",
			t.Source, t.Line, t.Column, typ,
		))
		panic(bailout{})
	}

	p.errors = append(p.errors, &parseError{
		&expectedError{
			ctx:       p.ctx,
			pos:       t.Position,
			expecting: []token.Type{typ},
		},
	})
}

func (p *parser) errorExpectedOneOf(t *token.Token, types ...token.Type) {
	p.errors = append(p.errors, &parseError{
		&expectedError{
			ctx:       p.ctx,
			pos:       t.Position,
			expecting: types,
		},
	})
}
