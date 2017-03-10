package parser

import (
	"fmt"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/operator"
	"github.com/erizocosmico/elmo/token"
)

func (p *parser) parseTerm() ast.Expr {
	switch p.tok.Type {
	case token.Int, token.Float, token.Char, token.String, token.True, token.False:
		return p.parseLiteral()
	case token.LeftParen:
		return p.parseLeftParen()
	case token.LeftBracket:
		return p.parseLeftBracket()
	case token.Dot:
		p.expect(token.Dot)
		return &ast.AccessorExpr{Field: p.parseLowerName()}
	case token.LeftBrace:
		return p.parseLeftBrace()
	case token.Backslash:
		return p.parseLambda()
	case token.Op:
		op := p.parseOp()
		if op.Name == "-" && p.tok.Offset-op.Pos() == 1 {
			return &ast.UnaryExpr{
				Op:   op,
				Expr: p.parseTerm(),
			}
		}

		p.errorMessage(p.tok.Position, fmt.Sprintf("I ran into an unexpected operator %s. I was expecting an expression.", op.Name))
		panic(bailout{})
	case token.Identifier:
		return p.parseIdentTerm()
	}

	return nil
}

func (p *parser) parseLambda() *ast.Lambda {
	l := &ast.Lambda{Backslash: p.expect(token.Backslash)}
	l.Args = p.parseFuncArgs(token.Arrow)
	l.Arrow = p.expect(token.Arrow)
	l.Expr = p.parseExpr()
	return l
}

func (p *parser) parseLeftBrace() ast.Expr {
	lbracePos := p.expect(token.LeftBrace)

	backup := p.tok
	ident := p.parseLowerName()
	if p.is(token.Pipe) {
		pipe := p.expect(token.Pipe)
		fields := p.parseRecordFields()
		if len(fields) == 0 {
			p.errorMessage(p.tok.Position, "I was expecting a list of record fields to update, but I got none.")
			// TODO: irrecoverable?
			panic(bailout{})
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
	fields := p.parseRecordFields()
	return &ast.RecordLit{
		Lbrace: lbracePos,
		Rbrace: p.expect(token.RightBrace),
		Fields: fields,
	}
}

func (p *parser) parseRecordFields() (fields []*ast.FieldAssign) {
	for !p.is(token.RightBrace) && !p.is(token.EOF) {
		name := p.parseLowerName()
		eq := p.expect(token.Assign)
		expr := p.parseExpr()
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

func (p *parser) parseLeftParen() ast.Expr {
	lparenPos := p.expect(token.LeftParen)
	// if next token is an op, we're looking at an operator
	// used as a function, e.g. `(++) a b`
	if p.is(token.Op) {
		op := p.parseOp()
		p.expect(token.RightParen)
		return op
	}

	expr := p.parseExpr()
	if p.is(token.Comma) {
		if expr == nil {
			n := 1
			for p.is(token.Comma) {
				p.expect(token.Comma)
				n++
			}

			// TODO: this should be an operator
			// change to an operator when Identifier and Operator are differnt
			return &ast.TupleCtor{
				Lparen: lparenPos,
				Rparen: p.expect(token.RightParen),
				Elems:  n,
			}
		}

		exprs := p.parseExprList(expr)
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

func (p *parser) parseLeftBracket() ast.Expr {
	lbracketPos := p.expect(token.LeftBracket)
	expr := p.parseExpr()
	if p.is(token.Comma) && !p.is(token.EOF) {
		if expr == nil {
			p.errorMessage(p.tok.Position, "I found ',', but I was expecting ']', whitespace or an expression")
			// TODO: not really recoverable?
			panic(bailout{})
		}

		exprs := p.parseExprList(expr)
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

func (p *parser) parseIdentTerm() ast.Expr {
	var path = []*ast.Ident{p.parseIdentifier()}

	for p.is(token.Dot) {
		// TODO: check is right after prev ident
		p.expect(token.Dot)
		path = append(path, p.parseIdentifier())
	}

	if len(path) == 1 {
		return path[0]
	}

	return ast.NewSelectorExpr(path...)
}

func (p *parser) parseExprList(first ast.Expr) []ast.Expr {
	var exprs = []ast.Expr{first}
	for p.is(token.Comma) {
		p.expect(token.Comma)
		exprs = append(exprs, p.parseExpr())
	}
	return exprs
}

func (p *parser) parseExpr() ast.Expr {
	switch p.tok.Type {
	case token.If:
		p.startRegion()
		defer p.endRegion()
		return p.parseIf()
	case token.Case:
		p.startRegion()
		defer p.endRegion()
		return p.parseCase()
	case token.Let:
		p.startRegion()
		defer p.endRegion()
		return p.parseLet()
	case token.EOF:
		p.errorMessage(p.tok.Position, "Unexpected EOF")
		panic(bailout{})
	}

	term := p.parseTerm()
	if term == nil {
		return nil
	}

	if _, ok := term.(*ast.BasicLit); ok || p.atLineStart() {
		return term
	}

	if p.isApplicable() || p.is(token.Op) {
		return p.parseBinaryExpr(term, 0)
	}

	return term
}

func (p *parser) isApplicable() bool {
	switch p.tok.Type {
	case token.Int, token.Float, token.String, token.Char, token.True, token.False,
		token.Identifier, token.LeftBrace, token.LeftParen, token.LeftBracket:
		return true
	}
	return false
}

const errorMsgMultipleNonAssocOps = `Binary operators %s and %s are non associative and have the same precedence. Consider using parenthesis to disambiguate.`

func (p *parser) parseBinaryExpr(lhs ast.Expr, precedence uint) ast.Expr {
	if p.isApplicable() {
		return p.parseBinaryExpr(&ast.FuncApp{
			Func: lhs,
			Args: []ast.Expr{
				p.parseTerm(),
			},
		}, 0)
	} else if !p.is(token.Op) {
		return tryFlattenApp(lhs)
	}

	opInfo := p.opInfo(p.tok.Value)
	for p.tok.Type == token.Op &&
		opInfo.Precedence >= precedence {
		op := p.parseOp()
		rhs := p.parseTerm()
		prevOp := opInfo
		opInfo = p.opInfo(p.tok.Value)

		for p.tok.Type == token.Op &&
			(opInfo.Precedence > prevOp.Precedence ||
				(opInfo.Associativity == operator.Right &&
					opInfo.Precedence == prevOp.Precedence)) {
			rhs = p.parseBinaryExpr(rhs, opInfo.Precedence)
			opInfo = p.opInfo(p.tok.Value)
		}

		if p.tok.Type != token.Op {
			rhs = p.parseBinaryExpr(rhs, 0)
		}

		lhs = &ast.BinaryExpr{
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

	return tryFlattenApp(lhs)
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

func (p *parser) parseLet() *ast.LetExpr {
	e := &ast.LetExpr{Let: p.expect(token.Let)}

	p.setLineStart(p.tok.Column)
	for p.is(token.Identifier) || p.is(token.LeftParen) || p.is(token.LeftBrace) {
		switch p.tok.Type {
		case token.Identifier:
			if p.tok.Value != "_" {
				e.Decls = append(e.Decls, p.parseDefinition())
			} else {
				e.Decls = append(e.Decls, p.parseDestructuringAssignment())
			}
		case token.LeftParen, token.LeftBrace:
			e.Decls = append(e.Decls, p.parseDestructuringAssignment())
		}
	}

	e.In = p.expect(token.In)
	p.resetLineStart()
	e.Body = p.parseExpr()
	return e
}

func (p *parser) parseIf() *ast.IfExpr {
	var expr = new(ast.IfExpr)
	expr.If = p.expect(token.If)
	expr.Cond = p.parseExpr()
	expr.Then = p.expect(token.Then)
	expr.ThenExpr = p.parseExpr()
	expr.Else = p.expect(token.Else)
	expr.ThenExpr = p.parseExpr()
	return expr
}

func (p *parser) parseCase() *ast.CaseExpr {
	var expr = new(ast.CaseExpr)
	expr.Case = p.expect(token.Case)
	expr.Expr = p.parseExpr()
	expr.Of = p.expect(token.Of)

	for {
		branch := p.parseCaseBranch()
		if branch == nil {
			break
		}
		expr.Branches = append(expr.Branches, branch)
	}

	return expr
}

func (p *parser) parseCaseBranch() *ast.CaseBranch {
	checkPoint := p.tok

	var branch = new(ast.CaseBranch)
	branch.Pattern = p.parsePattern(true)
	if !p.is(token.Arrow) {
		// this is not a case branch, we need to backup
		p.backup(checkPoint)
		return nil
	}

	branch.Arrow = p.expect(token.Arrow)
	branch.Expr = p.parseExpr()
	return branch
}

func (p *parser) parseFuncApp(term ast.Expr) *ast.FuncApp {
	app := &ast.FuncApp{Func: term}
	for !p.is(token.EOF) && !p.atLineStart() {
		term := p.parseTerm()
		if term == nil {
			break
		}

		app.Args = append(app.Args, term)
	}
	return app
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
		Type:     typ,
		Position: t.Position,
		Value:    t.Value,
	}
}
