package parser

import (
	"fmt"
	"strconv"

	"github.com/elm-tangram/tangram/ast"
	"github.com/elm-tangram/tangram/operator"
	"github.com/elm-tangram/tangram/token"
)

func parseModule(p *parser) *ast.ModuleDecl {
	var decl = new(ast.ModuleDecl)
	prevRegion := p.startRegion()

	stepOut := p.indentedBlock()
	decl.Module = p.expect(token.Module)
	decl.Name = parseModuleName(p)

	if p.is(token.Exposing) {
		p.expect(token.Exposing)
		decl.Exposing = parseExposedList(p, false)
	}

	p.endRegion(prevRegion)
	stepOut()
	return decl
}

func parseImports(p *parser) []*ast.ImportDecl {
	var imports []*ast.ImportDecl
	for p.tok.Type == token.Import {
		imports = append(imports, parseImport(p))
	}
	return imports
}

func parseImport(p *parser) *ast.ImportDecl {
	var decl = new(ast.ImportDecl)
	prevRegion := p.startRegion()
	stepOut := p.indentedBlock()
	decl.Import = p.expect(token.Import)
	decl.Module = parseModuleName(p)

	if p.is(token.As) {
		p.expect(token.As)
		decl.Alias = parseUpperName(p)
	}

	if p.is(token.Exposing) {
		p.expect(token.Exposing)
		decl.Exposing = parseExposedList(p, false)
	}

	p.endRegion(prevRegion)
	stepOut()
	return decl
}

// parseModuleName parses the name of a module in either module declarations
// or import declarations. The difference between this function and
// parseQualifiedIdentifier is the fact that this function enforces all the
// terms of the qualified identifier to be uppercase names.
func parseModuleName(p *parser) ast.Expr {
	path := []*ast.Ident{parseUpperName(p)}

	for p.is(token.Dot) {
		p.expectAfter(token.Dot, path[len(path)-1])
		path = append(path, parseUpperName(p))
	}

	if len(path) == 1 {
		return path[0]
	}

	return ast.NewSelectorExpr(path...)
}

func parseExposedList(p *parser, parsingUnion bool) ast.ExposedList {
	lparenPos := p.expect(token.LeftParen)
	if p.is(token.Range) {
		p.expect(token.Range)
		return &ast.OpenList{
			Lparen: lparenPos,
			Rparen: p.expect(token.RightParen),
		}
	}

	exposing := &ast.ClosedList{Lparen: lparenPos}
	exposing.Exposed = parseExposedIdents(p, parsingUnion)

	if len(exposing.Exposed) == 0 {
		p.errorExpectedOneOf(p.tok, token.Range, token.Identifier)
	}

	exposing.Rparen = p.expect(token.RightParen)
	return exposing
}

func parseExposedIdents(p *parser, parsingUnion bool) []ast.ExposedIdent {
	exposed := []ast.ExposedIdent{parseExposedIdent(p, parsingUnion)}
	for p.is(token.Comma) {
		p.expect(token.Comma)
		exposed = append(exposed, parseExposedIdent(p, parsingUnion))
	}

	return exposed
}

func parseExposedIdent(p *parser, parsingUnion bool) ast.ExposedIdent {
	var ident *ast.Ident
	if !parsingUnion {
		ident = parseIdentifierOrOp(p)
	} else {
		ident = parseUpperName(p)
	}

	if p.is(token.LeftParen) {
		if parsingUnion {
			p.errorMessage(p.tok.Position, "A constructor cannot expose anything.")
		}

		if !isUpper(ident.Name) {
			p.errorMessage(ident.NamePos, "%q is exposing constructors, but it is not a type.", ident.Name)
		}

		return &ast.ExposedUnion{
			Type:  ident,
			Ctors: parseExposedList(p, true),
		}
	}

	return &ast.ExposedVar{ident}
}

func parseDecl(p *parser) ast.Decl {
	prevRegion := p.startRegion()
	var decl ast.Decl
	switch p.tok.Type {
	case token.TypeDef:
		decl = parseTypeDecl(p)

	case token.Infixl, token.Infixr, token.Infix:
		decl = parseInfixDecl(p)

	case token.Identifier:
		if p.tok.Value == "_" {
			decl = parseDestructuringAssignment(p)
		} else {
			decl = parseDefinition(p)
		}

	case token.LeftParen:
		if p.peek().Type == token.Op {
			decl = parseDefinition(p)
		} else {
			decl = parseDestructuringAssignment(p)
		}

	case token.LeftBrace:
		decl = parseDestructuringAssignment(p)

	default:
		p.errorExpectedOneOf(p.tok, token.TypeDef, token.Identifier)
		panic(bailout{})
	}

	p.endRegion(prevRegion)

	if p.mode.Is(SkipDefinitions) {
		p.skipUntilNextFixity()
	}

	return decl
}

const errorMsgInvalidDestructuringPattern = `This is not a valid pattern for a destructuring assignment.
I am looking for one of the following things:

- a lower case name
- an underscore ("_")
- a tuple pattern (e.g. "(first, second)")
- a record pattern (e.g. "{x, y}")`

func parseDestructuringAssignment(p *parser) *ast.DestructuringAssignment {
	a := new(ast.DestructuringAssignment)
	indent, line := p.currentPos()
	a.Pattern = parsePattern(p, true)
	_, ok := a.Pattern.(ast.ArgPattern)
	if !ok {
		p.errorMessage(
			p.tok.Position,
			errorMsgInvalidDestructuringPattern,
		)
		panic(bailout{})
	}

	a.Eq = p.expect(token.Assign)
	stepOut := p.indentedBlockAt(indent, line)
	a.Expr = parseExpr(p)
	stepOut()

	return a
}

func parseInfixDecl(p *parser) ast.Decl {
	var assoc operator.Associativity
	if p.is(token.Infixl) {
		assoc = operator.Left
	} else if p.is(token.Infixr) {
		assoc = operator.Right
	}

	stepOut := p.indentedBlock()
	pos := p.expectOneOf(token.Infixl, token.Infixr, token.Infix)
	if !p.is(token.Int) {
		p.errorExpected(p.tok, token.Int)
	}

	precedence := parseLiteral(p)
	n, _ := strconv.Atoi(precedence.Value)
	if n < 0 || n > 9 {
		p.errorMessage(precedence.Position, "Operator precedence must be a number between 0 and 9, both included.")
	}

	op := parseOp(p)
	stepOut()
	return &ast.InfixDecl{
		InfixPos:   pos,
		Assoc:      assoc,
		Precedence: precedence,
		Op:         op,
	}
}

func parseTypeDecl(p *parser) ast.Decl {
	stepOut := p.indentedBlock()
	defer stepOut()
	typePos := p.expect(token.TypeDef)
	if p.is(token.Alias) {
		return parseAliasType(p, typePos)
	}

	return parseUnionType(p, typePos)
}

func parseAliasType(p *parser, typePos token.Pos) ast.Decl {
	decl := &ast.AliasDecl{
		TypePos: typePos,
		Alias:   p.expect(token.Alias),
	}
	decl.Name = parseUpperName(p)
	decl.Args = parseTypeDeclArgs(p)
	decl.Eq = p.expect(token.Assign)
	decl.Type = p.expectType()
	return decl
}

func parseUnionType(p *parser, typePos token.Pos) ast.Decl {
	decl := &ast.UnionDecl{TypePos: typePos}
	decl.Name = parseUpperName(p)
	decl.Args = parseTypeDeclArgs(p)
	decl.Eq = p.expect(token.Assign)
	decl.Ctors = parseConstructors(p)
	return decl
}

func parseTypeDeclArgs(p *parser) (idents []*ast.Ident) {
	for p.is(token.Identifier) {
		idents = append(idents, parseLowerName(p))
	}
	return
}

func parseConstructors(p *parser) (cs []*ast.Constructor) {
	cs = append(cs, parseConstructor(p))
	for p.is(token.Pipe) {
		p.expect(token.Pipe)
		ctor := parseConstructor(p)
		cs = append(cs, ctor)
	}
	return
}

func parseConstructor(p *parser) *ast.Constructor {
	c := new(ast.Constructor)
	c.Name = parseUpperName(p)
	c.Args = parseTypeList(p)
	return c
}

func parseDefinition(p *parser) ast.Decl {
	decl := new(ast.Definition)

	indent, line := p.currentPos()
	var name *ast.Ident
	if p.is(token.Identifier) {
		name = parseLowerName(p)
	} else {
		p.expect(token.LeftParen)
		name = parseOp(p)
		p.expect(token.RightParen)
	}

	if p.is(token.Colon) {
		decl.Annotation = &ast.TypeAnnotation{Name: name}
		decl.Annotation.Colon = p.expect(token.Colon)
		stepOut := p.indentedBlockAt(indent, line)
		decl.Annotation.Type = p.expectType()
		stepOut()

		indent, line = p.currentPos()
		defName := parseIdentifierOrOp(p)
		if defName.Name != name.Name {
			p.errorMessage(
				p.tok.Position,
				fmt.Sprintf(
					"A definition must be right below its type annotation, I found the definition of `%s` after the annotation of `%s` instead.",
					defName.Name,
					name.Name,
				),
			)
		}

		decl.Name = defName
	} else {
		decl.Name = name
	}

	decl.Args = parseFuncArgs(p, token.Assign)
	decl.Eq = p.expect(token.Assign)
	stepOut := p.indentedBlockAt(indent, line)
	decl.Body = parseExpr(p)
	stepOut()
	return decl
}
