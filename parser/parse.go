package parser

import (
	"strconv"

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

		case token.Infixl, token.Infixr:
			decls = append(decls, p.parseInfixDecl())

		case token.Identifier, token.LeftParen:
			decls = append(decls, p.parseDefinition())

		default:
			p.errorExpectedOneOf(p.tok, token.Import, token.TypeDef, token.Identifier)
			panic(bailout{})
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

func (p *parser) parseLiteral() *ast.BasicLit {
	var typ ast.BasicLitType
	switch p.tok.Type {
	case token.True, token.False:
		typ = ast.Bool
	case token.Int:
		typ = ast.Int
	case token.Float:
		typ = ast.Float
	case token.String:
		typ = ast.String
	case token.Char:
		typ = ast.Char
	}

	t := p.tok
	p.next()
	return &ast.BasicLit{
		Type:  typ,
		Pos:   t.Position,
		Value: t.Value,
	}
}

func (p *parser) parseInfixDecl() ast.Decl {
	dir := ast.Infixr
	if p.is(token.Infixl) {
		dir = ast.Infixl
	}

	pos := p.tok.Position
	p.expectOneOf(token.Infixl, token.Infixr)
	if !p.is(token.Int) {
		p.errorExpected(p.tok, token.Int)
	}

	priority := p.parseLiteral()
	n, _ := strconv.Atoi(priority.Value)
	if n < 1 || n > 9 {
		p.errorMessage(priority.Pos, "Operator priority must be a number between 1 and 9, both included.")
	}

	op := p.parseOp()
	return &ast.InfixDecl{
		InfixPos: pos,
		Dir:      dir,
		Priority: priority,
		Op:       op,
	}
}

func (p *parser) parseDefinition() ast.Decl {
	panic(bailout{})
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

func (p *parser) errorExpected(t *token.Token, typ token.Type) {
	if t.Type == token.EOF {
		p.errors = append(p.errors, &parseError{
			&unexpectedEOFError{
				ctx:       p.ctx,
				pos:       t.Position,
				expecting: []token.Type{typ},
			},
		})
		panic(bailout{})
	}

	p.errorExpectedOneOf(t, typ)
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

func (p *parser) errorMessage(pos *token.Position, msg string) {
	p.errors = append(p.errors, &parseError{
		&msgError{
			ctx: p.ctx,
			pos: pos,
			msg: msg,
		},
	})
}
