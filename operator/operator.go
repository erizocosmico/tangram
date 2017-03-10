package operator

import "github.com/erizocosmico/elmo/ast"

type Operator interface {
	Precedence() uint
	Associativity() Associativity
	Underlying() ast.Expr
	isOp()
}

func NewOperator(ident *ast.Ident, precedence uint, assoc Associativity) Operator {
	return &op{ident, precedence, assoc}
}

func NewFuncOperator(e ast.Expr) Operator {
	return &funcApp{e}
}

type funcApp struct {
	expr ast.Expr
}

func (f *funcApp) Precedence() uint             { return 10 }
func (f *funcApp) Associativity() Associativity { return Left }
func (f *funcApp) Underlying() ast.Expr         { return f.expr }
func (*funcApp) isOp()                          {}

type op struct {
	op    *ast.Ident
	prec  uint
	assoc Associativity
}

func (o *op) Precedence() uint             { return o.prec }
func (o *op) Associativity() Associativity { return o.assoc }
func (o *op) Underlying() ast.Expr         { return o.op }
func (*op) isOp()                          {}
