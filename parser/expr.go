package parser

import (
	"fmt"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/operator"
	"github.com/erizocosmico/elmo/token"
)

func parseTerm(p *parser) ast.Expr {
	switch p.tok.Type {
	case token.Int, token.Float, token.Char, token.String, token.True, token.False:
		return parseLiteral(p)
	case token.LeftParen:
		return parseLeftParen(p)
	case token.LeftBracket:
		return parseLeftBracket(p)
	case token.Dot:
		p.expect(token.Dot)
		return &ast.AccessorExpr{Field: parseLowerName(p)}
	case token.LeftBrace:
		return parseLeftBrace(p)
	case token.Backslash:
		return parseLambda(p)
	case token.Op:
		op := parseOp(p)
		if op.Name == "-" && p.tok.Offset-op.Pos() == 1 {
			return &ast.UnaryOp{
				Op:   op,
				Expr: parseTerm(p),
			}
		}

		p.errorMessage(p.tok.Position, fmt.Sprintf("I ran into an unexpected operator %s. I was expecting an expression.", op.Name))
		panic(bailout{})
	case token.Identifier:
		return parseIdentTerm(p)
	}

	return nil
}

// parseQualifiedIdentifier parses an identifier and its qualifier, if any.
// This should only be used outside of expressions, as expressions already
// parse selector expressions by themselves.
// This is specially to parse identifiers with an optional qualifier.
func parseQualifiedIdentifier(p *parser) ast.Expr {
	if isUpper(p.tok.Value) {
		var path []*ast.Ident
		path = append(path, parseUpperName(p))

		for p.is(token.Dot) {
			p.expectAfter(token.Dot, path[len(path)-1])

			if p.is(token.Identifier) {
				if isLower(p.tok.Value) {
					path = append(path, parseLowerName(p))
					// a lower identifier means the end of the path
					break
				}
				path = append(path, parseUpperName(p))
			} else {
				p.expect(token.Identifier)
				return nil
			}
		}

		if len(path) == 1 {
			return path[0]
		}

		return ast.NewSelectorExpr(path...)
	}
	return parseLowerName(p)
}

func parseIdentifier(p *parser) *ast.Ident {
	name := "_"
	pos := p.tok.Position
	if p.is(token.Identifier) {
		name = p.tok.Value
		p.next()
	} else {
		p.expect(token.Identifier)
	}

	return ast.NewIdent(name, pos)
}

func parseLambda(p *parser) *ast.Lambda {
	l := &ast.Lambda{Backslash: p.expect(token.Backslash)}
	l.Args = parseFuncArgs(p, token.Arrow)
	l.Arrow = p.expect(token.Arrow)
	l.Expr = parseExpr(p)
	return l
}

func parseLeftBrace(p *parser) ast.Expr {
	lbracePos := p.expect(token.LeftBrace)

	backup := p.tok
	ident := parseLowerName(p)
	if p.is(token.Pipe) {
		pipe := p.expect(token.Pipe)
		fields := parseRecordFields(p)
		if len(fields) == 0 {
			p.errorMessage(p.tok.Position, "I was expecting a list of record fields to update, but I got none.")
			return &ast.BadExpr{
				StartPos: lbracePos,
				EndPos:   p.expect(token.RightBrace),
			}
		}

		return &ast.RecordUpdate{
			Lbrace: lbracePos,
			Pipe:   pipe,
			Record: ident,
			Fields: fields,
			Rbrace: p.expect(token.RightBrace),
		}
	}

	p.backup(backup)
	fields := parseRecordFields(p)
	return &ast.RecordLit{
		Lbrace: lbracePos,
		Rbrace: p.expect(token.RightBrace),
		Fields: fields,
	}
}

func parseRecordFields(p *parser) (fields []*ast.FieldAssign) {
	for !p.is(token.RightBrace) && !p.is(token.EOF) {
		name := parseLowerName(p)
		eq := p.expect(token.Assign)
		expr := parseExpr(p)
		if !p.is(token.RightBrace) {
			p.expect(token.Comma)
		}

		fields = append(fields, &ast.FieldAssign{
			Eq:    eq,
			Field: name,
			Expr:  expr,
		})
	}

	return
}

func parseLeftParen(p *parser) ast.Expr {
	lparenPos := p.expect(token.LeftParen)
	// if next token is an op, we're looking at an operator
	// used as a function, e.g. `(++) a b`
	if p.is(token.Op) {
		op := parseOp(p)
		p.expect(token.RightParen)
		return op
	}

	expr := parseExpr(p)
	if p.is(token.Comma) {
		if expr == nil {
			n := 1
			for p.is(token.Comma) {
				p.expect(token.Comma)
				n++
			}

			return &ast.TupleCtor{
				Lparen: lparenPos,
				Rparen: p.expect(token.RightParen),
				Elems:  n,
			}
		}

		exprs := parseExprList(p, expr)
		return &ast.TupleLit{
			Lparen: lparenPos,
			Rparen: p.expect(token.RightParen),
			Elems:  exprs,
		}
	} else if p.is(token.RightParen) && expr == nil {
		// empty tuple
		return &ast.TupleLit{
			Lparen: lparenPos,
			Rparen: p.expect(token.RightParen),
		}
	}

	return &ast.ParensExpr{
		Lparen: lparenPos,
		Expr:   expr,
		Rparen: p.expect(token.RightParen),
	}
}

func parseLeftBracket(p *parser) ast.Expr {
	lbracketPos := p.expect(token.LeftBracket)
	expr := parseExpr(p)
	if p.is(token.Comma) && !p.is(token.EOF) {
		if expr == nil {
			p.errorMessage(p.tok.Position, "I found ',', but I was expecting ']', whitespace or an expression")
			for !p.is(token.RightBracket) && !p.is(token.EOF) {
				defer p.next()
				return &ast.BadExpr{
					StartPos: lbracketPos,
					EndPos:   p.tok.Offset,
				}
			}
		}

		exprs := parseExprList(p, expr)
		return &ast.ListLit{
			Lbracket: lbracketPos,
			Rbracket: p.expect(token.RightBracket),
			Elems:    exprs,
		}
	}

	lit := &ast.ListLit{
		Lbracket: lbracketPos,
		Rbracket: p.expect(token.RightBracket),
	}

	if expr != nil {
		lit.Elems = append(lit.Elems, expr)
	}

	return lit
}

func parseIdentTerm(p *parser) ast.Expr {
	var path = []*ast.Ident{parseIdentifier(p)}

	for p.is(token.Dot) {
		p.expectAfter(token.Dot, path[len(path)-1])
		path = append(path, parseIdentifier(p))
	}

	if len(path) == 1 {
		return path[0]
	}

	return ast.NewSelectorExpr(path...)
}

func parseExprList(p *parser, first ast.Expr) []ast.Expr {
	var exprs = []ast.Expr{first}
	for p.is(token.Comma) {
		p.expect(token.Comma)
		exprs = append(exprs, parseExpr(p))
	}
	return exprs
}

func parseExpr(p *parser) ast.Expr {
	switch p.tok.Type {
	case token.If:
		defer p.endRegion(p.startRegion())
		return parseIf(p)
	case token.Case:
		defer p.endRegion(p.startRegion())
		return parseCase(p)
	case token.Let:
		defer p.endRegion(p.startRegion())
		return parseLet(p)
	case token.EOF:
		p.errorMessage(p.tok.Position, "Unexpected EOF")
		panic(bailout{})
	}

	term := parseTerm(p)
	if term == nil {
		return nil
	}

	return parseBinaryOp(p, term, 0)
}

func atExprFinalizer(p *parser) bool {
	switch p.tok.Type {
	case token.Comma, token.RightBrace, token.RightParen,
		token.RightBracket, token.Then, token.Else, token.Of,
		token.EOF:
		return true
	}
	return !p.isCorrectlyIndented()
}

const errorMsgMultipleNonAssocOps = `Binary operators %s and %s are non associative and have the same precedence. Consider using parenthesis to disambiguate.`

func parseBinaryOp(p *parser, lhs ast.Expr, precedence uint) ast.Expr {
	lhs = tryFlattenApp(lhs)
	if atExprFinalizer(p) {
		return lhs
	}

	if !p.is(token.Op) {
		return parseBinaryOp(p, &ast.FuncApp{
			Func: lhs,
			Args: []ast.Expr{
				parseTerm(p),
			},
		}, 0)
	}

	opInfo := p.opInfo(p.tok.Value)
	for p.tok.Type == token.Op &&
		opInfo.Precedence >= precedence {
		op := parseOp(p)
		rhs := parseTerm(p)
		prevOp := opInfo
		opInfo = p.opInfo(p.tok.Value)

		for p.tok.Type == token.Op &&
			(opInfo.Precedence > prevOp.Precedence ||
				(opInfo.Associativity == operator.Right &&
					opInfo.Precedence == prevOp.Precedence)) {
			rhs = parseBinaryOp(p, rhs, opInfo.Precedence)
			opInfo = p.opInfo(p.tok.Value)
		}

		if p.tok.Type != token.Op {
			rhs = parseBinaryOp(p, rhs, 0)
		}

		lhs = &ast.BinaryOp{
			Op:  op,
			Lhs: lhs,
			Rhs: rhs,
		}

		if opInfo.Associativity == operator.NonAssoc &&
			opInfo.Precedence == prevOp.Precedence {
			p.errorMessage(p.tok.Position, fmt.Sprintf(
				errorMsgMultipleNonAssocOps,
				p.tok.Value,
				op.Name,
			))
			panic(bailout{})
		}
	}

	return lhs
}

func tryFlattenApp(expr ast.Expr) ast.Expr {
	if app, ok := expr.(*ast.FuncApp); ok {
		return flattenApp(app)
	}
	return expr
}

func flattenApp(app *ast.FuncApp) *ast.FuncApp {
	if app2, ok := app.Func.(*ast.FuncApp); ok {
		flatApp := flattenApp(app2)
		app.Func = flatApp.Func
		app.Args = append(flatApp.Args, app.Args...)
	}

	return app
}

func parseLet(p *parser) *ast.LetExpr {
	indent, line := p.currentPos()
	e := &ast.LetExpr{Let: p.expect(token.Let)}
	stepOut := p.indentedBlockAt(indent, line)
	for p.is(token.Identifier) || p.is(token.LeftParen) || p.is(token.LeftBrace) {
		switch p.tok.Type {
		case token.Identifier:
			if p.tok.Value != "_" {
				e.Decls = append(e.Decls, parseDefinition(p))
			} else {
				e.Decls = append(e.Decls, parseDestructuringAssignment(p))
			}
		case token.LeftParen, token.LeftBrace:
			e.Decls = append(e.Decls, parseDestructuringAssignment(p))
		}
	}

	stepOut()
	indent, line = p.currentPos()
	e.In = p.expect(token.In)
	stepOut = p.indentedBlockAt(indent, line)
	e.Body = parseExpr(p)
	stepOut()
	return e
}

func parseIf(p *parser) *ast.IfExpr {
	var expr = new(ast.IfExpr)

	indent, line := p.currentPos()
	expr.If = p.expect(token.If)
	expr.Cond = parseExpr(p)

	expr.Then = p.expect(token.Then)
	stepOut := p.indentedBlockAt(indent, line)
	expr.ThenExpr = parseExpr(p)
	stepOut()

	indent, line = p.currentPos()
	expr.Else = p.expect(token.Else)
	stepOut = p.indentedBlockAt(indent, line)
	expr.ElseExpr = parseExpr(p)
	stepOut()

	return expr
}

func parseCase(p *parser) *ast.CaseExpr {
	var expr = new(ast.CaseExpr)

	indent, line := p.currentPos()
	expr.Case = p.expect(token.Case)
	expr.Expr = parseExpr(p)
	expr.Of = p.expect(token.Of)

	firstBranchPos := p.tok.Position
	for !p.is(token.EOF) {
		stepOut := p.indentedBlockAt(indent, line)
		branch := parseCaseBranch(p, firstBranchPos.Column)
		stepOut()
		if branch == nil {
			break
		}
		expr.Branches = append(expr.Branches, branch)
	}

	if len(expr.Branches) == 0 {
		p.errorMessage(firstBranchPos, "I was expecting a pattern.")
	}

	return expr
}

func parseCaseBranch(p *parser, alignment int) *ast.CaseBranch {
	checkPoint := p.tok

	var branch = new(ast.CaseBranch)
	if alignment != p.tok.Column {
		return nil
	}

	indent, line := p.currentPos()
	p.silent = true
	branch.Pattern = parsePattern(p, true)
	if !p.is(token.Arrow) {
		// this is not a case branch, we need to backup
		p.backup(checkPoint)
		p.silent = false
		return nil
	}
	p.silent = false

	branch.Arrow = p.expect(token.Arrow)
	stepOut := p.indentedBlockAt(indent, line)
	branch.Expr = parseExpr(p)
	stepOut()
	return branch
}

func parseLiteral(p *parser) *ast.BasicLit {
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
		Type:     typ,
		Position: t.Position,
		Value:    t.Value,
	}
}
