package types

import "github.com/elm-tangram/tangram/ast"

type Expr interface {
	ast.Node
	isExpr()
}

type Var struct {
	ast.Node
	Name     string
	Selector string
}

type FieldVar struct {
	ast.Node
	Name   string
	Record Expr
}

type App struct {
	ast.Node
	Func Expr
	Args []Expr
}

type Let struct {
	ast.Node
	Defs []Expr
	Body Expr
}

type Abs struct {
	ast.Node
	Args []Pattern
	Body Expr
}

type Def struct {
	ast.Node
	Pattern Pattern
	Body    Expr
}

type RecordUpdate struct {
	*ast.RecordUpdate
	Record Expr
	Fields []Field
}

type If struct {
	*ast.IfExpr
	Cond Expr
	Then Expr
	Else Expr
}

type Case struct {
	*ast.CaseExpr
	Expr     Expr
	Branches []*CaseBranch
}

type CaseBranch struct {
	*ast.CaseBranch
	Pattern Pattern
	Expr    Expr
}

type Pattern interface {
	Expr
	isPattern()
}

type ListPattern struct {
	*ast.ListPattern
	Elems []Pattern
}

type CtorPattern struct {
	*ast.CtorPattern
	Ctor Expr
	Args []Pattern
}

type TuplePattern struct {
	*ast.TuplePattern
	Elems []Pattern
}

type RecordPattern struct {
	*ast.RecordPattern
	Fields []Pattern
}

type AliasPattern struct {
	*ast.AliasPattern
	Name    *Var
	Pattern Pattern
}

type AnythingPattern struct {
	*ast.AnythingPattern
}

type Lit interface {
	Expr
	isLit()
}

type BasicLit struct {
	*ast.BasicLit
	Kind  BasicLitType
	Value string
}

type BasicLitType byte

const (
	Int BasicLitType = 1 + iota
	Float
	Char
	String
	Bool
)

func (t BasicLitType) String() string {
	switch t {
	case Int:
		return "Int"
	case Float:
		return "Float"
	case Char:
		return "Char"
	case String:
		return "String"
	case Bool:
		return "Bool"
	}
	panic("unreachable")
}

type TupleLit struct {
	ast.Node
	Elems []Expr
}

type ListLit struct {
	*ast.ListLit
	Elems []Expr
}

type RecordLit struct {
	*ast.RecordLit
	Fields []Field
}

type Field struct {
	Name  string
	Value Expr
}

type Type interface {
	ast.Node
	isType()
	freeTypeVars() []string
	replaceVar(string, *Var) Type
}

type VarType struct {
	Name  string
	Rigid bool
}

const (
	number     = "number"
	comparable = "comparable"
	appendable = "appendable"
	compappend = "compappend"
)

type NamedType struct {
	Name *Var
	Args []Type
}

type FuncType struct {
	Args   []Type
	Return Type
}

type TupleType struct {
	Elems []Type
}

type RecordType struct {
	Fields []FieldType
}

type FieldType struct {
	Name string
	Type Type
}

type Scheme struct {
	Vars []string
	Type Type
}

type Types []Type

func (*Var) isExpr()             {}
func (*FieldVar) isExpr()        {}
func (*App) isExpr()             {}
func (*Abs) isExpr()             {}
func (*Let) isExpr()             {}
func (*Def) isExpr()             {}
func (*If) isExpr()              {}
func (*Case) isExpr()            {}
func (*RecordUpdate) isExpr()    {}
func (*ListLit) isExpr()         {}
func (*RecordLit) isExpr()       {}
func (*TupleLit) isExpr()        {}
func (*BasicLit) isExpr()        {}
func (*ListPattern) isExpr()     {}
func (*CtorPattern) isExpr()     {}
func (*TuplePattern) isExpr()    {}
func (*RecordPattern) isExpr()   {}
func (*AnythingPattern) isExpr() {}
func (*AliasPattern) isExpr()    {}

func (*Var) isPattern()             {}
func (*ListPattern) isPattern()     {}
func (*CtorPattern) isPattern()     {}
func (*TuplePattern) isPattern()    {}
func (*RecordPattern) isPattern()   {}
func (*AnythingPattern) isPattern() {}
func (*AliasPattern) isPattern()    {}

func (*ListLit) isLit()   {}
func (*RecordLit) isLit() {}
func (*TupleLit) isLit()  {}
func (*BasicLit) isLit()  {}

func (*VarType) isType()    {}
func (*NamedType) isType()  {}
func (*FuncType) isType()   {}
func (*RecordType) isType() {}
func (*TupleType) isType()  {}
func (*Scheme) isType()     {}
