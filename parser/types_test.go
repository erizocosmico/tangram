package parser

import (
	"testing"

	"github.com/mvader/elmo/ast"

	"github.com/stretchr/testify/require"
)

type typeAssert func(*testing.T, ast.Type)

type constructorAssert func(*testing.T, *ast.Constructor)

type declAssert func(*testing.T, ast.Decl)

func assertAlias(
	nameAssert func(*testing.T, *ast.Ident),
	argsAssert func(*testing.T, []*ast.Ident),
	typeAssert typeAssert,
) declAssert {
	return func(t *testing.T, decl ast.Decl) {
		alias, ok := decl.(*ast.AliasDecl)
		require.True(t, ok, "expected an alias decl")
		nameAssert(t, alias.Name)
		argsAssert(t, alias.Args)
		typeAssert(t, alias.Type)
	}
}

func assertUnion(
	nameAssert func(*testing.T, *ast.Ident),
	argsAssert func(*testing.T, []*ast.Ident),
	constructors ...constructorAssert,
) declAssert {
	return func(t *testing.T, decl ast.Decl) {
		union, ok := decl.(*ast.UnionDecl)
		require.True(t, ok, "expected an union decl")
		nameAssert(t, union.Name)
		argsAssert(t, union.Args)

		require.Equal(t, len(constructors), len(union.Types), "invalid number of constructors")
		for i := range constructors {
			constructors[i](t, union.Types[i])
		}
	}
}

func assertConstructor(name string, args ...typeAssert) constructorAssert {
	return func(t *testing.T, c *ast.Constructor) {
		require.Equal(t, name, c.Name.Name, "invalid type name")
		require.Equal(t, len(args), len(c.Args), "invalid number of type arguments")
		for i := range args {
			args[i](t, c.Args[i])
		}
	}
}

func assertName(name string) func(*testing.T, *ast.Ident) {
	return func(t *testing.T, ident *ast.Ident) {
		require.Equal(t, name, ident.Name, "invalid identifier")
	}
}

func assertNoArgs(t *testing.T, idents []*ast.Ident) {
	require.Equal(t, 0, len(idents), "expected no arguments")
}

func assertArgs(args ...string) func(*testing.T, []*ast.Ident) {
	return func(t *testing.T, idents []*ast.Ident) {
		require.Equal(t, len(args), len(idents), "idents number doesnt't match")
		for i := range args {
			assertName(args[i])(t, idents[i])
		}
	}
}

func assertTuple(types ...typeAssert) typeAssert {
	return func(t *testing.T, typ ast.Type) {
		tuple, ok := typ.(*ast.TupleType)
		require.True(t, ok, "type is not tuple")

		require.Equal(t, len(types), len(tuple.Elems), "invalid number of tuple elements")
		for i := range types {
			types[i](t, tuple.Elems[i])
		}
	}
}

func assertBasicType(name string, args ...typeAssert) typeAssert {
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

func assertRecord(fields ...recordFieldAssert) typeAssert {
	return func(t *testing.T, typ ast.Type) {
		record, ok := typ.(*ast.RecordType)
		require.True(t, ok, "type is not record type")
		require.Equal(t, len(fields), len(record.Fields), "invalid number of record fields")
		for i := range fields {
			fields[i](t, record.Fields[i])
		}
	}
}

func assertRecordField(name string, assertType typeAssert) recordFieldAssert {
	return func(t *testing.T, f *ast.RecordTypeField) {
		require.Equal(t, name, f.Name.Name, "invalid record field name")
		assertType(t, f.Type)
	}
}

func assertBasicRecordField(name, typ string) recordFieldAssert {
	return func(t *testing.T, f *ast.RecordTypeField) {
		require.Equal(t, name, f.Name.Name, "invalid record field name")
		assertBasicType(typ)(t, f.Type)
	}
}
