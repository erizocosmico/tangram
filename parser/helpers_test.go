package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/elm-tangram/tangram/ast"
	"github.com/elm-tangram/tangram/operator"

	"github.com/stretchr/testify/require"
)

type (
	FileAssert         func(*testing.T, *ast.Module)
	TypeAssert         func(*testing.T, ast.Type)
	ConstructorAssert  func(*testing.T, *ast.Constructor)
	DeclAssert         func(*testing.T, ast.Decl)
	AnnotationAssert   func(*testing.T, string, *ast.TypeAnnotation)
	ExprAssert         func(*testing.T, ast.Expr)
	PatternAssert      func(*testing.T, ast.Pattern)
	BranchAssert       func(*testing.T, *ast.CaseBranch)
	ExposedListAssert  func(*testing.T, ast.ExposedList)
	ExposedIdentAssert func(*testing.T, ast.ExposedIdent)
	ImportAssert       DeclAssert
)

func File(module DeclAssert, imports []ImportAssert, decls ...DeclAssert) FileAssert {
	return func(t *testing.T, f *ast.Module) {
		module(t, f.Module)
		require.Len(t, f.Imports, len(imports), "wrong number of imports")
		for i, imp := range f.Imports {
			imports[i](t, imp)
		}

		require.Len(t, f.Decls, len(decls), "wrong number of decls")
		for i, decl := range f.Decls {
			decls[i](t, decl)
		}
	}
}

func Module(module string, exposed ExposedListAssert) DeclAssert {
	return func(t *testing.T, decl ast.Decl) {
		d, ok := decl.(*ast.ModuleDecl)
		require.True(t, ok, "expecting decl to be ModuleDecl, is %T", decl)
		require.Equal(t, module, d.ModuleName())
		if exposed == nil {
			require.Nil(t, d.Exposing)
		} else {
			exposed(t, d.Exposing)
		}
	}
}

func Import(module string, alias ExprAssert, exposed ExposedListAssert) ImportAssert {
	return func(t *testing.T, decl ast.Decl) {
		d, ok := decl.(*ast.ImportDecl)
		require.True(t, ok, "expecting decl to be ImportDecl, is %T", decl)
		require.Equal(t, module, d.ModuleName())
		if alias == nil {
			require.Nil(t, d.Alias)
		} else {
			alias(t, d.Alias)
		}

		if exposed == nil {
			require.Nil(t, d.Exposing)
		} else {
			exposed(t, d.Exposing)
		}
	}
}

func ClosedList(idents ...ExposedIdentAssert) ExposedListAssert {
	return func(t *testing.T, list ast.ExposedList) {
		l, ok := list.(*ast.ClosedList)
		require.True(t, ok, "expecting exposed list to be ClosedList, is %T", list)
		require.Len(t, l.Exposed, len(idents), "invalid number of exposed idents")
		for i, id := range l.Exposed {
			idents[i](t, id)
		}
	}
}

func OpenList(t *testing.T, list ast.ExposedList) {
	_, ok := list.(*ast.OpenList)
	require.True(t, ok, "expecting exposed list to be OpenList")
}

func ExposedVar(name string) ExposedIdentAssert {
	return func(t *testing.T, ident ast.ExposedIdent) {
		v, ok := ident.(*ast.ExposedVar)
		require.True(t, ok, "expected exposed ident to be ExposedVar, is %T", ident)
		require.Equal(t, name, v.Name)
	}
}

func ExposedUnion(name string, exposed ExposedListAssert) ExposedIdentAssert {
	return func(t *testing.T, ident ast.ExposedIdent) {
		union, ok := ident.(*ast.ExposedUnion)
		require.True(t, ok, "expected exposed ident to be ExposedUnion, is %T", ident)
		require.Equal(t, name, union.Type.Name, "wrong exposed union type name")
		exposed(t, union.Ctors)
	}
}

func InfixDecl(op string, assoc operator.Associativity, prec ExprAssert) DeclAssert {
	return func(t *testing.T, decl ast.Decl) {
		d, ok := decl.(*ast.InfixDecl)
		require.True(t, ok, "expecting decl to be InfixDecl, is %T", decl)
		require.Equal(t, op, d.Op.Name)
		require.Equal(t, assoc, d.Assoc, "wrong associativity")
		prec(t, d.Precedence)
	}
}

func Definition(
	name string,
	annAssert AnnotationAssert,
	patterns []PatternAssert,
	exprAssert ExprAssert,
) DeclAssert {
	return func(t *testing.T, decl ast.Decl) {
		def, ok := decl.(*ast.Definition)
		require.True(t, ok, "expected declaration to be a Definition, is %T", decl)
		require.Equal(t, name, def.Name.Name)
		require.Equal(t, len(patterns), len(def.Args), "expected same number of arguments")
		for i := range patterns {
			patterns[i](t, def.Args[i])
		}

		if annAssert != nil {
			annAssert(t, name, def.Annotation)
		} else {
			require.Nil(t, def.Annotation, "expected no type annotation")
		}

		exprAssert(t, def.Body)
	}
}

func Destructuring(pattern PatternAssert, expr ExprAssert) DeclAssert {
	return func(t *testing.T, decl ast.Decl) {
		d, ok := decl.(*ast.DestructuringAssignment)
		require.True(t, ok, "expected definition to be a DestructuredAssignment, is %T", decl)
		pattern(t, d.Pattern)
		expr(t, d.Expr)
	}
}

func Alias(
	name string,
	args []string,
	typeAssert TypeAssert,
) DeclAssert {
	return func(t *testing.T, decl ast.Decl) {
		alias, ok := decl.(*ast.AliasDecl)
		require.True(t, ok, "expected an alias decl")
		assertIdent(t, name, alias.Name)
		assertIdents(t, args, alias.Args)
		typeAssert(t, alias.Type)
	}
}

func Union(
	name string,
	args []string,
	constructors ...ConstructorAssert,
) DeclAssert {
	return func(t *testing.T, decl ast.Decl) {
		union, ok := decl.(*ast.UnionDecl)
		require.True(t, ok, "expected an union decl")
		assertIdent(t, name, union.Name)
		assertIdents(t, args, union.Args)

		require.Equal(t, len(constructors), len(union.Ctors), "invalid number of constructors")
		for i := range constructors {
			constructors[i](t, union.Ctors[i])
		}
	}
}

func Constructor(name string, args ...TypeAssert) ConstructorAssert {
	return func(t *testing.T, c *ast.Constructor) {
		require.Equal(t, name, c.Name.Name, "invalid type name")
		require.Equal(t, len(args), len(c.Args), "invalid number of type arguments")
		for i := range args {
			args[i](t, c.Args[i])
		}
	}
}

func Tuple(types ...TypeAssert) TypeAssert {
	return func(t *testing.T, typ ast.Type) {
		tuple, ok := typ.(*ast.TupleType)
		require.True(t, ok, "type is not tuple")

		require.Equal(t, len(types), len(tuple.Elems), "invalid number of tuple elements")
		for i := range types {
			types[i](t, tuple.Elems[i])
		}
	}
}

func FuncType(elems ...TypeAssert) TypeAssert {
	return func(t *testing.T, typ ast.Type) {
		fn, ok := typ.(*ast.FuncType)
		require.True(t, ok, "type is not a function")

		require.Equal(t, len(elems), len(fn.Args)+1, "invalid number of elements in function signature")
		for i := range fn.Args {
			elems[i](t, fn.Args[i])
		}

		elems[len(elems)-1](t, fn.Return)
	}
}

func Selector(path ...string) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		e, ok := expr.(fmt.Stringer)
		require.True(t, ok, "expected expression in selector to be stringer")
		require.Equal(t, strings.Join(path, "."), e.String(), "expected same selector")
	}
}

func NamedType(name string, args ...TypeAssert) TypeAssert {
	return func(t *testing.T, typ ast.Type) {
		named, ok := typ.(*ast.NamedType)
		require.True(t, ok, "type is not named type, is %T", typ)
		ident, ok := named.Name.(*ast.Ident)
		require.True(t, ok, "expected type name to be an identifier, not %T", named.Name)
		require.Equal(t, name, ident.Name, "invalid type name")
		require.Equal(t, len(args), len(named.Args), "invalid number of type arguments")
		for i := range args {
			args[i](t, named.Args[i])
		}
	}
}

func VarType(name string) TypeAssert {
	return func(t *testing.T, typ ast.Type) {
		v, ok := typ.(*ast.VarType)
		require.True(t, ok, "type is not VarType, is %T", typ)
		require.Equal(t, name, v.Name)
	}
}

func SelectorType(sel ExprAssert, args ...TypeAssert) TypeAssert {
	return func(t *testing.T, typ ast.Type) {
		basic, ok := typ.(*ast.NamedType)
		require.True(t, ok, "type is not basic type, is %T", typ)
		sel(t, basic.Name)
		require.Equal(t, len(args), len(basic.Args), "invalid number of type arguments")
		for i := range args {
			args[i](t, basic.Args[i])
		}
	}
}

type recordFieldAssert func(*testing.T, *ast.RecordField)

func Record(fields ...recordFieldAssert) TypeAssert {
	return func(t *testing.T, typ ast.Type) {
		record, ok := typ.(*ast.RecordType)
		require.True(t, ok, "type is not record type")
		require.Equal(t, len(fields), len(record.Fields), "invalid number of record fields")
		for i := range fields {
			fields[i](t, record.Fields[i])
		}
	}
}

func RecordField(name string, assertType TypeAssert) recordFieldAssert {
	return func(t *testing.T, f *ast.RecordField) {
		require.Equal(t, name, f.Name.Name, "invalid record field name")
		assertType(t, f.Type)
	}
}

func BasicRecordField(name, typ string) recordFieldAssert {
	return func(t *testing.T, f *ast.RecordField) {
		require.Equal(t, name, f.Name.Name, "invalid record field name")
		NamedType(typ)(t, f.Type)
	}
}

func TypeAnnotation(typeAssert TypeAssert) AnnotationAssert {
	return func(t *testing.T, name string, ann *ast.TypeAnnotation) {
		require.Equal(t, name, ann.Name.Name)
		typeAssert(t, ann.Type)
	}
}

func TupleCtor(elems int) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		ctor, ok := expr.(*ast.TupleCtor)
		require.True(t, ok, "expected expr to be TupleCtor, is %T", expr)
		require.Equal(t, elems, ctor.Elems, "expected TupleCtor to have the ame number of elements")
	}
}

func ListLiteral(elems ...ExprAssert) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		lit, ok := expr.(*ast.ListLit)
		require.True(t, ok, "expected expr to be ListLit, is %T", expr)

		require.Len(t, lit.Elems, len(elems), "expected list to have this number of elements")
		for i := range elems {
			elems[i](t, lit.Elems[i])
		}
	}
}

func TupleLiteral(elems ...ExprAssert) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		lit, ok := expr.(*ast.TupleLit)
		require.True(t, ok, "expected expr to be TupleLit, is %T", expr)

		require.Len(t, lit.Elems, len(elems), "expected tuple to have this number of elements")
		for i := range elems {
			elems[i](t, lit.Elems[i])
		}
	}
}

func Identifier(name string) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		ident, ok := expr.(*ast.Ident)
		require.True(t, ok, "expected expr to be Identifier, is %T", expr)
		require.Equal(t, name, ident.Name)
	}
}

func Literal(kind ast.BasicLitType, val string) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		lit, ok := expr.(*ast.BasicLit)
		require.True(t, ok, "expected expr to be BasicLit, is %T", expr)

		require.Equal(t, kind, lit.Type)
		require.Equal(t, val, lit.Value)
	}
}

func Lambda(patterns []PatternAssert, assertExpr ExprAssert) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		lambda, ok := expr.(*ast.Lambda)
		require.True(t, ok, "expected expr to be Lambda, is %T", expr)

		require.Len(t, lambda.Args, len(patterns), "Lambda argument length")
		for i := range patterns {
			patterns[i](t, lambda.Args[i])
		}

		assertExpr(t, lambda.Expr)
	}
}

func BinaryOp(op string, lhs, rhs ExprAssert) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		binaryExpr, ok := expr.(*ast.BinaryOp)
		require.True(t, ok, "expected expr to be BinaryOp, is %T", expr)
		require.Equal(t, op, binaryExpr.Op.Name, "op name")
		lhs(t, binaryExpr.Lhs)
		rhs(t, binaryExpr.Rhs)
	}
}

func UnaryOp(op string, lhs ExprAssert) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		unaryExpr, ok := expr.(*ast.UnaryOp)
		require.True(t, ok, "expected expr to be UnaryOp, is %T", expr)
		require.Equal(t, op, unaryExpr.Op.Name, "op name")
		lhs(t, unaryExpr.Expr)
	}
}

func Parens(assert ExprAssert) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		parens, ok := expr.(*ast.ParensExpr)
		require.True(t, ok, "expected expr to be ParensExpr, is %T", expr)
		assert(t, parens.Expr)
	}
}

func FuncApp(fn ExprAssert, args ...ExprAssert) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		app, ok := expr.(*ast.FuncApp)
		require.True(t, ok, "expected expr to be FuncApp, is %T", expr)
		fn(t, app.Func)
		require.Len(t, args, len(app.Args), "func app arguments")
		for i := range args {
			args[i](t, app.Args[i])
		}
	}
}

type FieldAssignAssert func(*testing.T, *ast.FieldAssign)

func RecordUpdate(v string, fields ...FieldAssignAssert) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		record, ok := expr.(*ast.RecordUpdate)
		require.True(t, ok, "expected expr to be RecordUpdate, is %T", expr)
		require.Equal(t, v, record.Record.Name)
		require.Len(t, record.Fields, len(fields), "invalid number of record fields")
		for i := range fields {
			fields[i](t, record.Fields[i])
		}
	}
}

func RecordLiteral(fields ...FieldAssignAssert) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		record, ok := expr.(*ast.RecordLit)
		require.True(t, ok, "expected expr to be RecordLit, is %T", expr)
		require.Len(t, record.Fields, len(fields), "invalid number of record fields")
		for i := range fields {
			fields[i](t, record.Fields[i])
		}
	}
}

func Let(exprAssert ExprAssert, decls ...DeclAssert) ExprAssert {
	return func(t *testing.T, expr ast.Expr) {
		let, ok := expr.(*ast.LetExpr)
		require.True(t, ok, "expected expr to be LetExpr, is %T", expr)
		require.Len(t, let.Decls, len(decls), "invalid number of declarations")
		for i := range decls {
			decls[i](t, let.Decls[i])
		}
		exprAssert(t, let.Body)
	}
}

func FieldAssign(name string, expr ExprAssert) FieldAssignAssert {
	return func(t *testing.T, f *ast.FieldAssign) {
		require.Equal(t, name, f.Field.Name, "invalid record field name")
		expr(t, f.Expr)
	}
}

func AccessorExpr(field string) ExprAssert {
	return func(t *testing.T, e ast.Expr) {
		acc, ok := e.(*ast.AccessorExpr)
		require.True(t, ok, "expected expr to be AccessorExpr, is %T", e)
		require.Equal(t, field, acc.Field.Name)
	}
}

func IfExpr(cond, then, elseExpr ExprAssert) ExprAssert {
	return func(t *testing.T, e ast.Expr) {
		ifExpr, ok := e.(*ast.IfExpr)
		require.True(t, ok, "expected expr to be IfExpr, is %T", e)
		cond(t, ifExpr.Cond)
		then(t, ifExpr.ThenExpr)
		elseExpr(t, ifExpr.ElseExpr)
	}
}

func CaseExpr(expr ExprAssert, branches ...BranchAssert) ExprAssert {
	return func(t *testing.T, e ast.Expr) {
		caseExpr, ok := e.(*ast.CaseExpr)
		require.True(t, ok, "expected expr to be CaseExpr, is %T", e)
		expr(t, caseExpr.Expr)
		require.Len(t, caseExpr.Branches, len(branches))
		for i := range branches {
			branches[i](t, caseExpr.Branches[i])
		}
	}
}

func CaseBranch(pattern PatternAssert, expr ExprAssert) BranchAssert {
	return func(t *testing.T, branch *ast.CaseBranch) {
		pattern(t, branch.Pattern)
		expr(t, branch.Expr)
	}
}

func AliasPattern(underlying PatternAssert, name string) PatternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		alias, ok := pattern.(*ast.AliasPattern)
		require.True(t, ok, "expected an alias pattern")
		require.Equal(t, name, alias.Name.Name, "alias name")
		underlying(t, alias.Pattern)
	}
}

func AnythingPattern(t *testing.T, pattern ast.Pattern) {
	_, ok := pattern.(*ast.AnythingPattern)
	require.True(t, ok, "expected an anything pattern")
}

func TuplePattern(elems ...PatternAssert) PatternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		tuple, ok := pattern.(*ast.TuplePattern)
		require.True(t, ok, "expected a tuple pattern")
		require.Equal(t, len(elems), len(tuple.Elems), "expecting same number of tuple pattern elements")

		for i := range elems {
			elems[i](t, tuple.Elems[i])
		}
	}
}

func ListPattern(elems ...PatternAssert) PatternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		list, ok := pattern.(*ast.ListPattern)
		require.True(t, ok, "expected a list pattern")
		require.Equal(t, len(elems), len(list.Elems), "expecting same number of list pattern elements")

		for i := range elems {
			elems[i](t, list.Elems[i])
		}
	}
}

func RecordPattern(elems ...PatternAssert) PatternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		r, ok := pattern.(*ast.RecordPattern)
		require.True(t, ok, "expected a record pattern")
		require.Equal(t, len(elems), len(r.Fields), "expecting same number of record pattern elements")

		for i := range elems {
			elems[i](t, r.Fields[i])
		}
	}
}

func VarPattern(name string) PatternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		v, ok := pattern.(*ast.VarPattern)
		require.True(t, ok, "expected a var pattern")
		require.Equal(t, name, v.Name.Name, "expecting same var name")
	}
}

func LiteralPattern(typ ast.BasicLitType, val string) PatternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		l, ok := pattern.(*ast.LiteralPattern)
		require.True(t, ok, "expected a literal pattern")
		require.Equal(t, typ, l.Literal.Type, "expected same kind of literal")
		require.Equal(t, val, l.Literal.Value, "expected same value of literal")
	}
}

func CtorPattern(name string, elems ...PatternAssert) PatternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		ctor, ok := pattern.(*ast.CtorPattern)
		require.True(t, ok, "expected a constructor pattern")
		require.Equal(t, len(elems), len(ctor.Args), "expecting same number of constructor pattern elements")

		for i := range elems {
			elems[i](t, ctor.Args[i])
		}
	}
}

func Patterns(patterns ...PatternAssert) []PatternAssert {
	return patterns
}

func assertIdent(t *testing.T, name string, ident *ast.Ident) {
	require.Equal(t, name, ident.Name)
}

func assertIdents(t *testing.T, names []string, idents []*ast.Ident) {
	require.Equal(t, len(names), len(idents), "expected same number of identifiers")
	for i := range idents {
		assertIdent(t, names[i], idents[i])
	}
}
