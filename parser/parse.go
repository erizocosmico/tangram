package parser

import (
	"strconv"
	"unicode"

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
			decls = append(decls, p.parseTypeDecl())

		case token.Infixl, token.Infixr, token.Infix:
			decls = append(decls, p.parseInfixDecl())

		case token.Identifier, token.LeftParen:
			p.errorMessage(p.tok.Position, "Declarations are not implemented yet.")
			panic(bailout{})

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
		decl.Alias = p.parseUpperName()
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
	name := ast.ModuleName{p.parseUpperName()}

	for {
		if !p.is(token.Dot) {
			break
		}

		p.expect(token.Dot)
		name = append(name, p.parseUpperName())
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
		if !unicode.IsUpper(rune(ident.Name[0])) {
			p.errorMessage(ident.NamePos, "I was expecting an upper case name.")
		}
		var exposingList = new(ast.ExposingList)
		exposingList.Lparen = p.expect(token.LeftParen)
		exposingList.Idents = p.parseConstructorExposedIdents()
		if len(exposingList.Idents) == 0 {
			p.errorExpectedOneOf(p.tok, token.Range, token.Identifier)
		}
		exposingList.Rparen = p.expect(token.RightParen)
		ident.Exposing = exposingList
	}

	return ident
}

func (p *parser) parseConstructorExposedIdents() (idents []*ast.ExposedIdent) {
	if p.is(token.Range) {
		p.expect(token.Range)
		idents = append(
			idents,
			&ast.ExposedIdent{
				Ident: &ast.Ident{Name: token.Range.String(), NamePos: p.tok.Position},
			},
		)
		return
	}

	for {
		idents = append(
			idents,
			&ast.ExposedIdent{Ident: p.parseUpperName()},
		)
		if p.is(token.RightParen) {
			return
		}

		p.expect(token.Comma)
	}
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

func (p *parser) parseUpperName() *ast.Ident {
	ident := p.parseIdentifier()
	if !unicode.IsUpper(rune(ident.Name[0])) {
		p.errorMessage(ident.NamePos, "I was expecting an upper case name.")
	}
	return ident
}

func (p *parser) parseLowerName() *ast.Ident {
	ident := p.parseIdentifier()
	if !unicode.IsLower(rune(ident.Name[0])) {
		p.errorMessage(ident.NamePos, "I was expecting a lower case name.")
	}
	return ident
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
	var assoc ast.Associativity
	if p.is(token.Infixl) {
		assoc = ast.LeftAssoc
	} else if p.is(token.Infixr) {
		assoc = ast.RightAssoc
	}

	pos := p.expectOneOf(token.Infixl, token.Infixr, token.Infix)
	if !p.is(token.Int) {
		p.errorExpected(p.tok, token.Int)
	}

	priority := p.parseLiteral()
	n, _ := strconv.Atoi(priority.Value)
	if n < 0 || n > 9 {
		p.errorMessage(priority.Pos, "Operator priority must be a number between 0 and 9, both included.")
	}

	op := p.parseOp()
	return &ast.InfixDecl{
		InfixPos: pos,
		Assoc:    assoc,
		Priority: priority,
		Op:       op,
	}
}

func (p *parser) parseTypeDecl() ast.Decl {
	typePos := p.expect(token.TypeDef)
	if p.is(token.Alias) {
		return p.parseAliasType(typePos)
	}

	return p.parseUnionType(typePos)
}

func (p *parser) parseAliasType(typePos token.Pos) ast.Decl {
	decl := &ast.AliasDecl{
		TypePos: typePos,
		Alias:   p.expect(token.Alias),
	}
	decl.Name = p.parseUpperName()
	decl.Args = p.parseTypeDeclArgs()
	decl.Eq = p.expect(token.Assign)
	decl.Type = p.parseType()
	return decl
}

func (p *parser) parseUnionType(typePos token.Pos) ast.Decl {
	decl := &ast.UnionDecl{TypePos: typePos}
	decl.Name = p.parseUpperName()
	decl.Args = p.parseTypeDeclArgs()
	decl.Eq = p.expect(token.Assign)
	decl.Types = p.parseConstructors()
	return decl
}

func (p *parser) parseTypeDeclArgs() (idents []*ast.Ident) {
	for p.is(token.Identifier) {
		idents = append(idents, p.parseLowerName())
	}
	return
}

func (p *parser) parseConstructors() (cs []*ast.Constructor) {
	cs = append(cs, p.parseConstructor())
	for p.is(token.Pipe) {
		cs = append(cs, p.parseConstructor())
	}
	return
}

func (p *parser) parseConstructor() *ast.Constructor {
	c := new(ast.Constructor)
	if p.is(token.Pipe) {
		c.Pipe = p.expect(token.Pipe)
	}

	c.Name = p.parseUpperName()
	c.Args = p.parseTypeList()
	return c
}

func (p *parser) parseTypeList() (types []ast.Type) {
	for p.is(token.LeftParen) || p.is(token.Identifier) || p.is(token.LeftBrace) {
		types = append(types, p.parseType())
	}
	return
}

func (p *parser) parseType() ast.Type {
	switch p.tok.Type {
	case token.LeftParen:
		lparenPos := p.expect(token.LeftParen)
		typ := p.parseType()

		// is a tuple
		if p.is(token.Comma) {
			t := &ast.TupleType{
				Lparen: lparenPos,
				Elems:  []ast.Type{typ},
			}

			for !p.is(token.RightParen) {
				p.expect(token.Comma)
				t.Elems = append(t.Elems, p.parseType())
			}

			t.Rparen = p.expect(token.RightParen)
			return t
		}

		p.expect(token.RightParen)
		return typ
	case token.Identifier:
		name := p.parseIdentifier()
		if unicode.IsLower(rune(name.Name[0])) {
			return &ast.BasicType{Name: name}
		}

		return &ast.BasicType{
			Name: name,
			Args: p.parseTypeList(),
		}
	case token.LeftBrace:
		return p.parseRecordType()
	default:
		p.errorExpectedOneOf(p.tok, token.LeftParen, token.LeftBrace, token.Identifier)
		// TODO: think of a better way to recover from this error
		panic(bailout{})
	}
}

func (p *parser) parseRecordType() *ast.RecordType {
	t := &ast.RecordType{
		Lbrace: p.expect(token.LeftBrace),
	}

	for !p.is(token.RightBrace) {
		comma := token.NoPos
		if len(t.Fields) > 0 {
			comma = p.expect(token.Comma)
		}

		f := &ast.RecordTypeField{Comma: comma}
		f.Name = p.parseLowerName()
		f.Colon = p.expect(token.Colon)
		f.Type = p.parseType()
		t.Fields = append(t.Fields, f)
	}

	t.Rbrace = p.expect(token.RightBrace)
	return t
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
