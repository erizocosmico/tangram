package parser

import (
	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/token"
)

func parseTypeList(p *parser) (types []ast.Type) {
	for (p.is(token.LeftParen) || p.is(token.Identifier) || p.is(token.LeftBrace)) && p.isCorrectlyIndented() {
		var typ ast.Type
		switch p.tok.Type {
		case token.LeftParen, token.LeftBrace:
			typ = parseType(p)
		case token.Identifier:
			ident := parseQualifiedIdentifier(p)
			if name, ok := ident.(*ast.Ident); ok && isLower(name.Name) {
				typ = &ast.VarType{name}
			} else {
				typ = &ast.NamedType{Name: ident}
			}
		}

		if typ == nil {
			break
		}
		types = append(types, typ)
	}
	return
}

// parseType parses a complete type, whether it is a function type or an atom
// type.
func parseType(p *parser) ast.Type {
	t := parseAtomType(p)
	if t == nil {
		return nil
	}

	if !p.is(token.Arrow) {
		return t
	}

	types := []ast.Type{t}
	for p.is(token.Arrow) {
		p.expect(token.Arrow)
		typ := parseAtomType(p)
		if typ == nil {
			break
		}

		types = append(types, typ)
	}

	size := len(types)
	return &ast.FuncType{
		Args:   types[:size-1],
		Return: types[size-1],
	}
}

// parseAtomType parses a type that can make sense on their own, that is,
// a tuple, a record or a basic type.
func parseAtomType(p *parser) ast.Type {
	if !p.isCorrectlyIndented() {
		return nil
	}

	switch p.tok.Type {
	case token.LeftParen:
		lparenPos := p.expect(token.LeftParen)
		typ := p.expectType()

		// is a tuple
		if p.is(token.Comma) {
			t := &ast.TupleType{
				Lparen: lparenPos,
				Elems:  []ast.Type{typ},
			}

			for !p.is(token.RightParen) {
				p.expect(token.Comma)
				t.Elems = append(t.Elems, p.expectType())
			}

			t.Rparen = p.expect(token.RightParen)
			return t
		}

		p.expect(token.RightParen)
		return typ
	case token.Identifier:
		name := parseQualifiedIdentifier(p)
		if name, ok := name.(*ast.Ident); ok && isLower(name.Name) {
			return &ast.VarType{name}
		}

		return &ast.NamedType{
			Name: name,
			Args: parseTypeList(p),
		}
	case token.LeftBrace:
		return parseRecordType(p)
	default:
		p.errorExpectedOneOf(p.tok, token.LeftParen, token.LeftBrace, token.Identifier)
		panic(bailout{})
	}
}

func parseRecordType(p *parser) *ast.RecordType {
	t := &ast.RecordType{
		Lbrace: p.expect(token.LeftBrace),
	}

	for !p.is(token.RightBrace) && !p.is(token.EOF) {
		if len(t.Fields) > 0 {
			p.expect(token.Comma)
		}

		f := new(ast.RecordField)
		f.Name = parseLowerName(p)
		f.Colon = p.expect(token.Colon)
		f.Type = p.expectType()
		t.Fields = append(t.Fields, f)
	}

	t.Rbrace = p.expect(token.RightBrace)
	return t
}
