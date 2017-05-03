package parser

import (
	"github.com/elm-tangram/tangram/ast"
	"github.com/elm-tangram/tangram/token"
)

const errorMsgInvalidArgPattern = `This is not a valid pattern for a function argument.
I am looking for one of the following things:

- a lower case name
- an underscore ("_")
- a tuple pattern (e.g. "(first, second)")
- a record pattern (e.g. "{x, y}")`

func parseFuncArgs(p *parser, end token.Type) []ast.Pattern {
	var args []ast.Pattern
	for !p.is(end) && !p.is(token.EOF) {
		tok := p.tok
		// in arguments, we parse the patterns as non-greedy so it forces the
		// developer to wrap around parenthesis the alias pattern
		pattern := parsePattern(p, false)
		arg, ok := pattern.(ast.ArgPattern)
		if !ok {
			p.errorMessage(
				tok.Position,
				errorMsgInvalidArgPattern,
			)
		}

		args = append(args, arg)
	}

	return args
}

// parsePattern parses the next pattern. If `greedy` is true, it will try to
// find an alias at the end of the pattern, otherwise it will not.
func parsePattern(p *parser, greedy bool) (pat ast.Pattern) {
	pat = &ast.VarPattern{Name: &ast.Ident{Name: "_"}}
	switch p.tok.Type {
	case token.Identifier:
		if p.tok.Value == "_" {
			pat = &ast.AnythingPattern{Underscore: p.expect(token.Identifier)}
		} else {
			if isUpper(p.tok.Value) {
				pat = parseCtorPattern(p)
			} else {
				pat = &ast.VarPattern{
					Name: parseLowerName(p),
				}
			}
		}
	case token.LeftBracket:
		pat = parseListPattern(p)
	case token.LeftParen:
		pat = parseTupleOrParenthesizedPattern(p)
	case token.LeftBrace:
		pat = parseRecordPattern(p)
	case token.Int, token.Char, token.String, token.Float:
		pat = &ast.LiteralPattern{parseLiteral(p)}
	case token.True, token.False:
		p.expectOneOf(token.True, token.False)
		pat = &ast.CtorPattern{Ctor: ast.NewIdent(p.tok.Value, p.tok.Position)}
	default:
		p.errorExpectedOneOf(p.tok, token.Identifier, token.LeftParen, token.LeftBrace, token.LeftBracket)
	}

	if p.is(token.As) && greedy {
		return parseAliasPattern(p, pat)
	}

	if p.is(token.Op) && p.tok.Value == "::" {
		return parseCtorListPattern(p, pat)
	}

	return
}

func parseListPattern(p *parser) ast.Pattern {
	lbracketPos := p.expect(token.LeftBracket)

	if p.is(token.RightBracket) {
		return &ast.ListPattern{
			Lbracket: lbracketPos,
			Rbracket: p.expect(token.RightBracket),
		}
	}

	pat := &ast.ListPattern{Lbracket: lbracketPos}
	pat.Elems = []ast.Pattern{parsePattern(p, true)}
	for !p.is(token.RightBracket) {
		p.expect(token.Comma)
		pat.Elems = append(pat.Elems, parsePattern(p, true))
	}

	pat.Rbracket = p.expect(token.RightBracket)
	return pat
}

func parseCtorListPattern(p *parser, pat ast.Pattern) ast.Pattern {
	pos := p.tok.Position
	p.expect(token.Op)
	return &ast.CtorPattern{
		Ctor: &ast.Ident{
			Name:    "::",
			NamePos: pos,
		},
		Args: []ast.Pattern{
			pat,
			parsePattern(p, false),
		},
	}
}

func parseTupleOrParenthesizedPattern(p *parser) ast.Pattern {
	lparenPos := p.expect(token.LeftParen)

	var patterns []ast.Pattern
	for !p.is(token.RightParen) {
		patterns = append(patterns, parsePattern(p, true))

		if !p.is(token.RightParen) {
			p.expect(token.Comma)
		}
	}

	rparenPos := p.expect(token.RightParen)
	if len(patterns) > 1 {
		return &ast.TuplePattern{
			Lparen: lparenPos,
			Elems:  patterns,
			Rparen: rparenPos,
		}
	}
	return patterns[0]
}

func parseRecordPattern(p *parser) ast.Pattern {
	lbracePos := p.expect(token.LeftBrace)

	var patterns []ast.Pattern
	for !p.is(token.RightBrace) {
		patterns = append(patterns, parsePattern(p, true))

		if !p.is(token.RightBrace) {
			p.expect(token.Comma)
		}
	}

	return &ast.RecordPattern{
		Lbrace: lbracePos,
		Fields: patterns,
		Rbrace: p.expect(token.RightBrace),
	}
}

func parseCtorPattern(p *parser) ast.Pattern {
	pat := &ast.CtorPattern{Ctor: parseUpperQualifiedIdentifier(p)}
	var patterns []ast.Pattern

Outer:
	for {
		switch p.tok.Type {
		case token.Identifier, token.LeftParen, token.LeftBracket, token.LeftBrace, token.True, token.False, token.Int, token.Char, token.Float:
			patterns = append(patterns, parsePattern(p, false))
		default:
			break Outer
		}
	}

	pat.Args = patterns
	return pat
}

func parseAliasPattern(p *parser, pat ast.Pattern) ast.Pattern {
	p.expect(token.As)
	return &ast.AliasPattern{
		Name:    parseLowerName(p),
		Pattern: pat,
	}
}
