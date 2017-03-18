package ast

import (
	"fmt"
	"testing"

	"github.com/erizocosmico/elmo/token"

	"github.com/stretchr/testify/require"
)

type testVisitor struct {
	visited map[string]int
}

func (v *testVisitor) Visit(node Node) Visitor {
	if node != nil {
		// these methods are not going to be tested for obvious reasons
		// they are used here only for coverage purposes and to spot any
		// possible failure in them.
		node.Pos()
		node.End()
		if stringer, ok := node.(fmt.Stringer); ok {
			stringer.String()
		}

		v.visited[fmt.Sprintf("%T", node)]++
	}
	return v
}

func TestWalk(t *testing.T) {
	v := &testVisitor{make(map[string]int)}
	Walk(v, testFile)

	for typ, num := range expectedVisits {
		require.Equal(t, num, v.visited[typ], "visits for type %s", typ)
	}
}

var expectedVisits = map[string]int{
	"*ast.File": 1,
}

func inc(name string) {
	expectedVisits[name]++
}

var testFile = &File{
	// module Foo.Bar exposing (foo, Foo(A))
	Module: mkModuleDecl(
		mkSelectorExpr(
			mkIdent("Foo"),
			mkIdent("Bar"),
		),
		mkExposingList(
			mkExposedIdent(
				mkIdent("foo"),
				nil,
			),
			mkExposedIdent(
				mkIdent("Foo"),
				mkExposingList(
					mkExposedIdent(
						mkIdent("A"),
						nil,
					),
				),
			),
		),
	),

	Imports: []*ImportDecl{
		// import Foo.Baz as Baz
		mkImportDecl(
			mkSelectorExpr(
				mkIdent("Foo"),
				mkIdent("Baz"),
			),
			mkIdent("Baz"),
			nil,
		),

		// import Qux exposing (qux)
		mkImportDecl(
			mkIdent("Qux"),
			nil,
			mkExposingList(
				mkExposedIdent(
					mkIdent("qux"),
					nil,
				),
			),
		),
	},

	Decls: []Decl{
		// infixl 5 :>
		mkInfixDecl(
			LeftAssoc,
			mkIdent(":>"),
			mkBasicLit(Int, "5"),
		),

		// type alias TupleAlias a b = ( a , b )
		mkAliasDecl(
			mkIdent("TupleAlias"),
			[]*Ident{
				mkIdent("a"),
				mkIdent("b"),
			},
			mkTupleType(
				mkBasicType(mkIdent("a")),
				mkBasicType(mkIdent("b")),
			),
		),

		// type UnionT a b = Foo (List a)
		//                 | Bar { x: b, y : b }
		//                 | Baz
		mkUnionDecl(
			mkIdent("UnionT"),
			[]*Ident{
				mkIdent("a"),
				mkIdent("b"),
			},
			mkConstructor(
				mkIdent("Foo"),
				mkBasicType(
					mkIdent("List"),
					mkBasicType(mkIdent("a")),
				),
			),
			mkConstructor(
				mkIdent("Bar"),
				mkRecordType(
					mkRecordField(
						mkIdent("x"),
						mkBasicType(mkIdent("b")),
					),
					mkRecordField(
						mkIdent("y"),
						mkBasicType(mkIdent("b")),
					),
				),
			),
			mkConstructor(mkIdent("Baz")),
		),

		// ( x, y ) = point
		mkDestructuringAssignment(
			mkTuplePattern(
				mkVarPattern(mkIdent("x")),
				mkVarPattern(mkIdent("y")),
			),
			mkTupleLit(
				mkBasicLit(Int, "1"),
				mkBasicLit(Int, "2"),
			),
		),

		// { x, y } = point
		mkDestructuringAssignment(
			mkRecordPattern(
				mkVarPattern(mkIdent("x")),
				mkVarPattern(mkIdent("y")),
			),
			mkIdent("point"),
		),

		// incrX p = { p | x = p.x + 1 }
		mkDefinition(
			nil,
			mkIdent("incrX"),
			[]Pattern{mkVarPattern(mkIdent("p"))},
			mkRecordUpdate(
				mkIdent("p"),
				mkFieldAssign(
					mkIdent("x"),
					mkBinaryOp(
						mkIdent("+"),
						mkSelectorExpr(
							mkIdent("p"),
							mkIdent("x"),
						),
						mkBasicLit(Int, "1"),
					),
				),
			),
		),

		// sum : Int -> Int -> Int
		// sum a b = a + b
		mkDefinition(
			mkTypeAnnotation(
				mkIdent("sum"),
				mkFuncType(
					[]Type{
						mkBasicType(mkIdent("Int")),
						mkBasicType(mkIdent("Int")),
					},
					mkBasicType(mkIdent("Int")),
				),
			),
			mkIdent("sum"),
			[]Pattern{
				mkVarPattern(mkIdent("a")),
				mkVarPattern(mkIdent("b")),
			},
			mkBinaryOp(
				mkIdent("+"),
				mkIdent("a"),
				mkIdent("b"),
			),
		),

		// typCase a b = case foo a b of
		//   Foo c d -> if c > d then
		//                c
		//              else
		//                d
		//   [1, c] as t -> { bar = c, baz = t }
		//   Baz -> -b
		//   _ -> a
		mkDefinition(
			nil,
			mkIdent("tryCase"),
			[]Pattern{
				mkVarPattern(mkIdent("a")),
				mkVarPattern(mkIdent("b")),
			},
			mkCaseExpr(
				mkFuncApp(
					mkIdent("foo"),
					mkIdent("a"),
					mkIdent("b"),
				),
				mkCaseBranch(
					mkCtorPattern(
						mkIdent("Foo"),
						mkVarPattern(mkIdent("c")),
						mkVarPattern(mkIdent("d")),
					),
					mkIfExpr(
						mkBinaryOp(
							mkIdent(">"),
							mkIdent("c"),
							mkIdent("d"),
						),
						mkIdent("c"),
						mkIdent("d"),
					),
				),
				mkCaseBranch(
					mkAliasPattern(
						mkIdent("t"),
						mkListPattern(
							mkLiteralPattern(
								mkBasicLit(Int, "1"),
							),
							mkVarPattern(mkIdent("c")),
						),
					),
					mkRecordLit(
						mkFieldAssign(
							mkIdent("bar"),
							mkIdent("c"),
						),
						mkFieldAssign(
							mkIdent("baz"),
							mkIdent("t"),
						),
					),
				),
				mkCaseBranch(
					mkCtorPattern(
						mkIdent("Baz"),
					),
					mkUnaryOp(
						mkIdent("-"),
						mkIdent("b"),
					),
				),
				mkCaseBranch(
					mkAnythingPattern(),
					mkIdent("a"),
				),
			),
		),

		// a =
		// 	let
		//		a = \x -> [x, 1]
		//		b = .x
		//		c = (,,)
		//		d = (foo a) b
		// 	in
		//		bar a b c d
		mkDefinition(
			nil,
			mkIdent("a"),
			nil,
			mkLetExpr(
				[]Decl{
					mkDefinition(
						nil,
						mkIdent("a"),
						nil,
						mkLambda(
							[]Pattern{mkVarPattern(mkIdent("x"))},
							mkListLit(mkIdent("x"), mkBasicLit(Int, "1")),
						),
					),

					mkDefinition(
						nil,
						mkIdent("b"),
						nil,
						mkAccessorExpr(mkIdent("x")),
					),

					mkDefinition(
						nil,
						mkIdent("c"),
						nil,
						mkTupleCtor(2),
					),

					mkDefinition(
						nil,
						mkIdent("d"),
						nil,
						mkFuncApp(
							mkParensExpr(
								mkFuncApp(
									mkIdent("foo"),
									mkIdent("a"),
								),
							),
							mkIdent("b"),
						),
					),
				},
				mkFuncApp(
					mkIdent("bar"),
					mkIdent("a"),
					mkIdent("b"),
					mkIdent("c"),
					mkIdent("d"),
				),
			),
		),
	},
}

func mkIdent(name string) *Ident {
	inc("*ast.Ident")
	return NewIdent(name, new(token.Position))
}

func mkModuleDecl(name Expr, exposing *ExposingList) *ModuleDecl {
	inc("*ast.ModuleDecl")
	return &ModuleDecl{Name: name, Exposing: exposing}
}

func mkExposingList(idents ...*ExposedIdent) *ExposingList {
	inc("*ast.ExposingList")
	return &ExposingList{Idents: idents}
}

func mkExposedIdent(name *Ident, exposing *ExposingList) *ExposedIdent {
	inc("*ast.ExposedIdent")
	return &ExposedIdent{Ident: name, Exposing: exposing}
}

func mkImportDecl(module Expr, alias *Ident, exposing *ExposingList) *ImportDecl {
	inc("*ast.ImportDecl")
	return &ImportDecl{Module: module, Alias: alias, Exposing: exposing}
}

func mkSelectorExpr(idents ...*Ident) *SelectorExpr {
	inc("*ast.SelectorExpr")
	return NewSelectorExpr(idents...)
}

func mkBasicLit(kind BasicLitType, val string) *BasicLit {
	inc("*ast.BasicLit")
	return &BasicLit{Type: kind, Value: val, Position: new(token.Position)}
}

func mkInfixDecl(assoc Associativity, op *Ident, prec *BasicLit) *InfixDecl {
	inc("*ast.InfixDecl")
	return &InfixDecl{Assoc: assoc, Op: op, Precedence: prec}
}

func mkAliasDecl(name *Ident, args []*Ident, typ Type) *AliasDecl {
	inc("*ast.AliasDecl")
	return &AliasDecl{Name: name, Args: args, Type: typ}
}

func mkUnionDecl(name *Ident, args []*Ident, types ...*Constructor) *UnionDecl {
	inc("*ast.UnionDecl")
	return &UnionDecl{Name: name, Args: args, Types: types}
}

func mkConstructor(name *Ident, args ...Type) *Constructor {
	inc("*ast.Constructor")
	return &Constructor{Name: name, Args: args}
}

func mkDestructuringAssignment(pat Pattern, expr Expr) *DestructuringAssignment {
	inc("*ast.DestructuringAssignment")
	return &DestructuringAssignment{Pattern: pat, Expr: expr}
}

func mkDefinition(ann *TypeAnnotation, name *Ident, args []Pattern, body Expr) *Definition {
	inc("*ast.Definition")
	return &Definition{ann, name, token.NoPos, args, body}
}

func mkTypeAnnotation(name *Ident, typ Type) *TypeAnnotation {
	inc("*ast.TypeAnnotation")
	return &TypeAnnotation{name, token.NoPos, typ}
}

func mkTupleType(elems ...Type) *TupleType {
	inc("*ast.TupleType")
	return &TupleType{Elems: elems}
}

func mkBasicType(name Expr, args ...Type) *BasicType {
	inc("*ast.BasicType")
	return &BasicType{Name: name, Args: args}
}

func mkFuncType(args []Type, returnType Type) *FuncType {
	inc("*ast.FuncType")
	return &FuncType{Args: args, Return: returnType}
}

func mkRecordType(fields ...*RecordField) *RecordType {
	inc("*ast.RecordType")
	return &RecordType{Fields: fields}
}

func mkRecordField(name *Ident, typ Type) *RecordField {
	inc("*ast.RecordField")
	return &RecordField{Name: name, Type: typ}
}

func mkVarPattern(name *Ident) *VarPattern {
	inc("*ast.VarPattern")
	return &VarPattern{name}
}

func mkAnythingPattern() *AnythingPattern {
	inc("*ast.AnythingPattern")
	return new(AnythingPattern)
}

func mkLiteralPattern(lit *BasicLit) *LiteralPattern {
	inc("*ast.LiteralPattern")
	return &LiteralPattern{lit}
}

func mkAliasPattern(name *Ident, pat Pattern) *AliasPattern {
	inc("*ast.AliasPattern")
	return &AliasPattern{name, pat}
}

func mkCtorPattern(ctor *Ident, patterns ...Pattern) *CtorPattern {
	inc("*ast.CtorPattern")
	return &CtorPattern{ctor, patterns}
}

func mkTuplePattern(patterns ...Pattern) *TuplePattern {
	inc("*ast.TuplePattern")
	return &TuplePattern{Patterns: patterns}
}

func mkRecordPattern(patterns ...Pattern) *RecordPattern {
	inc("*ast.RecordPattern")
	return &RecordPattern{Patterns: patterns}
}

func mkListPattern(patterns ...Pattern) *ListPattern {
	inc("*ast.ListPattern")
	return &ListPattern{Patterns: patterns}
}

func mkTupleLit(elems ...Expr) *TupleLit {
	inc("*ast.TupleLit")
	return &TupleLit{Elems: elems}
}

func mkFuncApp(fn Expr, args ...Expr) *FuncApp {
	inc("*ast.FuncApp")
	return &FuncApp{fn, args}
}

func mkRecordLit(fields ...*FieldAssign) *RecordLit {
	inc("*ast.RecordLit")
	return &RecordLit{Fields: fields}
}

func mkFieldAssign(field *Ident, expr Expr) *FieldAssign {
	inc("*ast.FieldAssign")
	return &FieldAssign{Field: field, Expr: expr}
}

func mkRecordUpdate(record *Ident, fields ...*FieldAssign) *RecordUpdate {
	inc("*ast.RecordUpdate")
	return &RecordUpdate{Record: record, Fields: fields}
}

func mkLetExpr(decls []Decl, body Expr) *LetExpr {
	inc("*ast.LetExpr")
	return &LetExpr{Decls: decls, Body: body}
}

func mkIfExpr(cond, then, elseExpr Expr) *IfExpr {
	inc("*ast.IfExpr")
	return &IfExpr{Cond: cond, ThenExpr: then, ElseExpr: elseExpr}
}

func mkCaseExpr(expr Expr, branches ...*CaseBranch) *CaseExpr {
	inc("*ast.CaseExpr")
	return &CaseExpr{Expr: expr, Branches: branches}
}

func mkCaseBranch(pattern Pattern, expr Expr) *CaseBranch {
	inc("*ast.CaseBranch")
	return &CaseBranch{token.NoPos, pattern, expr}
}

func mkListLit(elems ...Expr) *ListLit {
	inc("*ast.ListLit")
	return &ListLit{Elems: elems}
}

func mkUnaryOp(op *Ident, expr Expr) *UnaryOp {
	inc("*ast.UnaryOp")
	return &UnaryOp{Op: op, Expr: expr}
}

func mkBinaryOp(op *Ident, lhs, rhs Expr) *BinaryOp {
	inc("*ast.BinaryOp")
	return &BinaryOp{op, lhs, rhs}
}

func mkAccessorExpr(field *Ident) *AccessorExpr {
	inc("*ast.AccessorExpr")
	return &AccessorExpr{field}
}

func mkTupleCtor(elems int) *TupleCtor {
	inc("*ast.TupleCtor")
	return &TupleCtor{Elems: elems}
}

func mkLambda(args []Pattern, expr Expr) *Lambda {
	inc("*ast.Lambda")
	return &Lambda{Args: args, Expr: expr}
}

func mkParensExpr(expr Expr) *ParensExpr {
	inc("*ast.ParensExpr")
	return &ParensExpr{Expr: expr}
}
