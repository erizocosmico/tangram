package parser

import (
	"testing"

	"github.com/erizocosmico/elmo/ast"

	"github.com/stretchr/testify/require"
)

type (
	typeAssert        func(*testing.T, ast.Type)
	constructorAssert func(*testing.T, *ast.Constructor)
	declAssert        func(*testing.T, ast.Decl)
	annotationAssert  func(*testing.T, string, *ast.TypeAnnotation)
	exprAssert        func(*testing.T, ast.Expr)
	patternAssert     func(*testing.T, ast.Pattern)
)

func Definition(
	name string,
	annAssert annotationAssert,
	patterns []patternAssert,
	exprAssert exprAssert,
) declAssert {
	return func(t *testing.T, decl ast.Decl) {
		def, ok := decl.(*ast.Definition)
		require.True(t, ok, "expected a definition decl")
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

func Alias(
	name string,
	args []string,
	typeAssert typeAssert,
) declAssert {
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
	constructors ...constructorAssert,
) declAssert {
	return func(t *testing.T, decl ast.Decl) {
		union, ok := decl.(*ast.UnionDecl)
		require.True(t, ok, "expected an union decl")
		assertIdent(t, name, union.Name)
		assertIdents(t, args, union.Args)

		require.Equal(t, len(constructors), len(union.Types), "invalid number of constructors")
		for i := range constructors {
			constructors[i](t, union.Types[i])
		}
	}
}

func Constructor(name string, args ...typeAssert) constructorAssert {
	return func(t *testing.T, c *ast.Constructor) {
		require.Equal(t, name, c.Name.Name, "invalid type name")
		require.Equal(t, len(args), len(c.Args), "invalid number of type arguments")
		for i := range args {
			args[i](t, c.Args[i])
		}
	}
}

func Tuple(types ...typeAssert) typeAssert {
	return func(t *testing.T, typ ast.Type) {
		tuple, ok := typ.(*ast.TupleType)
		require.True(t, ok, "type is not tuple")

		require.Equal(t, len(types), len(tuple.Elems), "invalid number of tuple elements")
		for i := range types {
			types[i](t, tuple.Elems[i])
		}
	}
}

func FuncType(elems ...typeAssert) typeAssert {
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

func BasicType(name string, args ...typeAssert) typeAssert {
	return func(t *testing.T, typ ast.Type) {
		basic, ok := typ.(*ast.BasicType)
		require.True(t, ok, "type is not basic type")
		require.Equal(t, name, basic.Name.Name, "invalid type name")
		require.Equal(t, len(args), len(basic.Args), "invalid number of type arguments")
		for i := range args {
			args[i](t, basic.Args[i])
		}
	}
}

type recordFieldAssert func(*testing.T, *ast.RecordTypeField)

func Record(fields ...recordFieldAssert) typeAssert {
	return func(t *testing.T, typ ast.Type) {
		record, ok := typ.(*ast.RecordType)
		require.True(t, ok, "type is not record type")
		require.Equal(t, len(fields), len(record.Fields), "invalid number of record fields")
		for i := range fields {
			fields[i](t, record.Fields[i])
		}
	}
}

func RecordField(name string, assertType typeAssert) recordFieldAssert {
	return func(t *testing.T, f *ast.RecordTypeField) {
		require.Equal(t, name, f.Name.Name, "invalid record field name")
		assertType(t, f.Type)
	}
}

func BasicRecordField(name, typ string) recordFieldAssert {
	return func(t *testing.T, f *ast.RecordTypeField) {
		require.Equal(t, name, f.Name.Name, "invalid record field name")
		BasicType(typ)(t, f.Type)
	}
}

func TypeAnnotation(typeAssert typeAssert) annotationAssert {
	return func(t *testing.T, name string, ann *ast.TypeAnnotation) {
		require.Equal(t, name, ann.Name.Name)
		typeAssert(t, ann.Type)
	}
}

func Literal(kind ast.BasicLitType, val string) exprAssert {
	return func(t *testing.T, expr ast.Expr) {
		lit, ok := expr.(*ast.BasicLit)
		require.True(t, ok, "expected expr to be BasicLit")

		require.Equal(t, kind, lit.Type)
		require.Equal(t, val, lit.Value)
	}
}

func AliasPattern(underlying patternAssert, name string) patternAssert {
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

func TuplePattern(elems ...patternAssert) patternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		tuple, ok := pattern.(*ast.TuplePattern)
		require.True(t, ok, "expected a tuple pattern")
		require.Equal(t, len(elems), len(tuple.Patterns), "expecting same number of tuple pattern elements")

		for i := range elems {
			elems[i](t, tuple.Patterns[i])
		}
	}
}

func ListPattern(elems ...patternAssert) patternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		list, ok := pattern.(*ast.ListPattern)
		require.True(t, ok, "expected a list pattern")
		require.Equal(t, len(elems), len(list.Patterns), "expecting same number of list pattern elements")

		for i := range elems {
			elems[i](t, list.Patterns[i])
		}
	}
}

func RecordPattern(elems ...patternAssert) patternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		r, ok := pattern.(*ast.RecordPattern)
		require.True(t, ok, "expected a record pattern")
		require.Equal(t, len(elems), len(r.Patterns), "expecting same number of record pattern elements")

		for i := range elems {
			elems[i](t, r.Patterns[i])
		}
	}
}

func VarPattern(name string) patternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		v, ok := pattern.(*ast.VarPattern)
		require.True(t, ok, "expected a var pattern")
		require.Equal(t, name, v.Name.Name, "expecting same var name")
	}
}

func LiteralPattern(typ ast.BasicLitType, val string) patternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		l, ok := pattern.(*ast.LiteralPattern)
		require.True(t, ok, "expected a literal pattern")
		require.Equal(t, typ, l.Literal.Type, "expected same kind of literal")
		require.Equal(t, val, l.Literal.Value, "expected same value of literal")
	}
}

func CtorPattern(name string, elems ...patternAssert) patternAssert {
	return func(t *testing.T, pattern ast.Pattern) {
		ctor, ok := pattern.(*ast.CtorPattern)
		require.True(t, ok, "expected a constructor pattern")
		require.Equal(t, len(elems), len(ctor.Patterns), "expecting same number of constructor pattern elements")

		for i := range elems {
			elems[i](t, ctor.Patterns[i])
		}
	}
}

func Patterns(patterns ...patternAssert) []patternAssert {
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
