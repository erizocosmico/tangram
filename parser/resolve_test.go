package parser

import (
	"testing"

	"github.com/elm-tangram/tangram/ast"
	"github.com/elm-tangram/tangram/report"
	"github.com/elm-tangram/tangram/source"
	"github.com/elm-tangram/tangram/token"
	"github.com/stretchr/testify/require"
)

func TestResolvePattern(t *testing.T) {
	r := newTestResolver(t)
	newScope := func() *ast.NodeScope {
		return scopeWithObjects(
			ast.NewObject("Just", ast.Ctor, nil),
			ast.NewObject("Int", ast.Typ, nil),
			ast.NewObject("String", ast.Typ, nil),
		)
	}

	t.Run("AliasPattern", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		r.resolvePattern(scope, &ast.AliasPattern{
			Name: ast.NewIdent("foo", token.NoPos),
			Pattern: &ast.LiteralPattern{
				&ast.BasicLit{
					Type:  ast.Int,
					Value: "1",
				},
			},
		})

		require.Len(scope.Objects, 1)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["foo"])
		require.True(r.reporter.IsOK())
	})

	t.Run("CtorPattern", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.CtorPattern{
			Ctor: ast.NewIdent("Just", token.NoPos),
			Args: []ast.Pattern{
				&ast.VarPattern{ast.NewIdent("a", token.NoPos)},
			},
		}
		r.resolvePattern(scope, node)

		require.Len(scope.Objects, 1)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["a"])
		assertObj(t, node.Ctor, "Just")
		require.True(r.reporter.IsOK())
	})

	t.Run("TuplePattern", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.TuplePattern{
			Elems: []ast.Pattern{
				&ast.VarPattern{ast.NewIdent("a", token.NoPos)},
				&ast.VarPattern{ast.NewIdent("b", token.NoPos)},
			},
		}
		r.resolvePattern(scope, node)

		require.Len(scope.Objects, 2)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["a"])
		require.NotNil(scope.Objects["b"])
		require.True(r.reporter.IsOK())
	})

	t.Run("ListPattern", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.ListPattern{
			Elems: []ast.Pattern{
				&ast.VarPattern{ast.NewIdent("a", token.NoPos)},
				&ast.VarPattern{ast.NewIdent("b", token.NoPos)},
			},
		}
		r.resolvePattern(scope, node)

		require.Len(scope.Objects, 2)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["a"])
		require.NotNil(scope.Objects["b"])
		require.True(r.reporter.IsOK())
	})

	t.Run("RecordPattern", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.RecordPattern{
			Fields: []ast.Pattern{
				&ast.VarPattern{ast.NewIdent("a", token.NoPos)},
				&ast.VarPattern{ast.NewIdent("b", token.NoPos)},
				new(ast.AnythingPattern),
			},
		}
		r.resolvePattern(scope, node)

		require.Len(scope.Objects, 2)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["a"])
		require.NotNil(scope.Objects["b"])
		require.True(r.reporter.IsOK())
	})

	t.Run("VarPattern", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		r.resolvePattern(scope, &ast.VarPattern{ast.NewIdent("a", token.NoPos)})

		require.Len(scope.Objects, 1)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["a"])
		require.True(r.reporter.IsOK())
	})
}

func TestResolveType(t *testing.T) {
	r := newTestResolver(t)
	newScope := func() *ast.NodeScope {
		return scopeWithObjects(
			ast.NewObject("Int", ast.Typ, nil),
			ast.NewObject("String", ast.Typ, nil),
			ast.NewObject("Result", ast.Typ, nil),
		)
	}

	t.Run("NamedType", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()

		node := &ast.NamedType{
			Name: ast.NewIdent("Result", token.NoPos),
			Args: []ast.Type{
				&ast.NamedType{Name: ast.NewIdent("String", token.NoPos)},
				&ast.NamedType{Name: ast.NewIdent("Int", token.NoPos)},
			},
		}
		r.resolveType(scope, node, false)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Name, "Result")
		assertObj(t, node.Args[0].(*ast.NamedType).Name, "String")
		assertObj(t, node.Args[1].(*ast.NamedType).Name, "Int")
		require.True(r.reporter.IsOK())
	})

	t.Run("VarType", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()

		node := &ast.VarType{ast.NewIdent("a", token.NoPos)}
		r.resolveType(scope, node, false)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		require.Nil(node.Obj)
		require.True(r.reporter.IsOK())
	})

	t.Run("FuncType", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()

		node := &ast.FuncType{
			Args: []ast.Type{
				&ast.NamedType{Name: ast.NewIdent("String", token.NoPos)},
				&ast.NamedType{Name: ast.NewIdent("Int", token.NoPos)},
			},
			Return: &ast.NamedType{Name: ast.NewIdent("Int", token.NoPos)},
		}
		r.resolveType(scope, node, false)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Return.(*ast.NamedType).Name, "Int")
		assertObj(t, node.Args[0].(*ast.NamedType).Name, "String")
		assertObj(t, node.Args[1].(*ast.NamedType).Name, "Int")
		require.True(r.reporter.IsOK())
	})

	t.Run("TupleType", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()

		node := &ast.TupleType{
			Elems: []ast.Type{
				&ast.NamedType{Name: ast.NewIdent("String", token.NoPos)},
				&ast.NamedType{Name: ast.NewIdent("Int", token.NoPos)},
			},
		}

		r.resolveType(scope, node, false)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Elems[0].(*ast.NamedType).Name, "String")
		assertObj(t, node.Elems[1].(*ast.NamedType).Name, "Int")
		require.True(r.reporter.IsOK())
	})

	t.Run("RecordType", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()

		node := &ast.RecordType{
			Fields: []*ast.RecordField{
				{
					Name: ast.NewIdent("x", token.NoPos),
					Type: &ast.NamedType{Name: ast.NewIdent("Int", token.NoPos)},
				},
				{
					Name: ast.NewIdent("y", token.NoPos),
					Type: &ast.NamedType{Name: ast.NewIdent("Int", token.NoPos)},
				},
			},
		}
		r.resolveType(scope, node, false)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		require.Nil(node.Fields[0].Name.Obj)
		assertObj(t, node.Fields[0].Type.(*ast.NamedType).Name, "Int")
		require.Nil(node.Fields[1].Name.Obj)
		assertObj(t, node.Fields[1].Type.(*ast.NamedType).Name, "Int")
		require.True(r.reporter.IsOK())
	})

	t.Run("RecordType repeated fields", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()

		node := &ast.RecordType{
			Fields: []*ast.RecordField{
				{
					Name: ast.NewIdent("x", token.NoPos),
					Type: &ast.NamedType{Name: ast.NewIdent("Int", token.NoPos)},
				},
				{
					Name: ast.NewIdent("x", token.NoPos),
					Type: &ast.NamedType{Name: ast.NewIdent("Int", token.NoPos)},
				},
			},
		}
		r.resolveType(scope, node, false)

		assertReports(t, r.reporter, new(report.RepeatedFieldError))
		require.False(r.reporter.IsOK())
	})
}

func TestResolveDecl(t *testing.T) {
	r := newTestResolver(t)
	newScope := func() *ast.NodeScope {
		return scopeWithObjects(
			ast.NewObject("Int", ast.Typ, nil),
			ast.NewObject("String", ast.Typ, nil),
			ast.NewObject("Result", ast.Typ, nil),
			ast.NewObject("c", ast.Var, nil),
			ast.NewObject("d", ast.Var, nil),
			ast.NewObject("?", ast.Var, nil),
			ast.NewObject("A", ast.Ctor, nil),
		)
	}

	t.Run("DestructuringAssignment", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.DestructuringAssignment{
			Pattern: &ast.TuplePattern{
				Elems: []ast.Pattern{
					&ast.VarPattern{ast.NewIdent("a", token.NoPos)},
					&ast.VarPattern{ast.NewIdent("b", token.NoPos)},
				},
			},
			Expr: &ast.TupleLit{
				Elems: []ast.Expr{
					ast.NewIdent("c", token.NoPos),
					ast.NewIdent("d", token.NoPos),
				},
			},
		}
		r.resolveDecl(scope, node)

		require.Len(scope.Objects, 2)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["a"])
		require.NotNil(scope.Objects["b"])
		assertObj(t, node.Expr.(*ast.TupleLit).Elems[0], "c")
		assertObj(t, node.Expr.(*ast.TupleLit).Elems[1], "d")
		require.True(r.reporter.IsOK())
	})

	t.Run("InfixDecl", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.InfixDecl{
			Op: ast.NewIdent("?", token.NoPos),
		}
		r.resolveDecl(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Op, "?")
		require.True(r.reporter.IsOK())
	})

	t.Run("Definition", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.Definition{
			Annotation: &ast.TypeAnnotation{
				Type: &ast.FuncType{
					Args: []ast.Type{
						&ast.NamedType{Name: ast.NewIdent("String", token.NoPos)},
						&ast.NamedType{Name: ast.NewIdent("Int", token.NoPos)},
					},
					Return: &ast.NamedType{Name: ast.NewIdent("String", token.NoPos)},
				},
			},
			Name: ast.NewIdent("formatNum", token.NoPos),
			Args: []ast.Pattern{
				&ast.VarPattern{ast.NewIdent("format", token.NoPos)},
				&ast.VarPattern{ast.NewIdent("n", token.NoPos)},
			},
			Body: ast.NewIdent("c", token.NoPos),
		}
		r.resolveDecl(scope, node)

		require.Len(scope.Objects, 1)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["formatNum"])
		assertObj(t, node.Annotation.Type.(*ast.FuncType).Args[0].(*ast.NamedType).Name, "String")
		assertObj(t, node.Annotation.Type.(*ast.FuncType).Args[1].(*ast.NamedType).Name, "Int")
		assertObj(t, node.Annotation.Type.(*ast.FuncType).Return.(*ast.NamedType).Name, "String")
		assertObj(t, node.Body, "c")

		require.Len(scope.Children(), 1)
		scope = scope.Children()[0]
		require.Len(scope.Objects, 2)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["format"])
		require.NotNil(scope.Objects["n"])

		require.True(r.reporter.IsOK())
	})

	t.Run("AliasDecl", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.AliasDecl{
			Name: ast.NewIdent("FancyResult", token.NoPos),
			Args: []*ast.Ident{
				ast.NewIdent("a", token.NoPos),
				ast.NewIdent("b", token.NoPos),
			},
			Type: &ast.NamedType{
				Name: ast.NewIdent("Result", token.NoPos),
				Args: []ast.Type{
					&ast.VarType{ast.NewIdent("a", token.NoPos)},
					&ast.VarType{ast.NewIdent("b", token.NoPos)},
				},
			},
		}
		r.resolveDecl(scope, node)

		require.Len(scope.Objects, 1)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["FancyResult"])
		assertObj(t, node.Type.(*ast.NamedType).Name, "Result")
		require.True(r.reporter.IsOK())
	})

	t.Run("AliasDecl repeated var types", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.AliasDecl{
			Name: ast.NewIdent("FancyResult", token.NoPos),
			Args: []*ast.Ident{
				ast.NewIdent("a", token.NoPos),
				ast.NewIdent("a", token.NoPos),
			},
			Type: &ast.NamedType{
				Name: ast.NewIdent("Result", token.NoPos),
				Args: []ast.Type{
					&ast.VarType{ast.NewIdent("a", token.NoPos)},
					&ast.VarType{ast.NewIdent("a", token.NoPos)},
				},
			},
		}
		r.resolveDecl(scope, node)

		require.False(r.reporter.IsOK())
		assertReports(t, r.reporter, new(report.RepeatedVarTypeError))
	})

	t.Run("UnionDecl", func(t *testing.T) {
		r := newTestResolver(t)
		require := require.New(t)
		scope := newScope()
		node := &ast.UnionDecl{
			Name: ast.NewIdent("Maybe", token.NoPos),
			Args: []*ast.Ident{
				ast.NewIdent("a", token.NoPos),
			},
			Ctors: []*ast.Constructor{
				&ast.Constructor{
					Name: ast.NewIdent("Just", token.NoPos),
					Args: []ast.Type{
						&ast.VarType{ast.NewIdent("a", token.NoPos)},
					},
				},
				&ast.Constructor{
					Name: ast.NewIdent("Nothing", token.NoPos),
				},
			},
		}
		r.resolveDecl(scope, node)

		require.Len(scope.Objects, 3)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["Maybe"])
		require.NotNil(scope.Objects["Just"])
		require.NotNil(scope.Objects["Nothing"])
		require.Len(r.reporter.Reports("test"), 0)
		require.True(r.reporter.IsOK())
	})

	t.Run("UnionDecl repeated vars", func(t *testing.T) {
		r := newTestResolver(t)
		require := require.New(t)
		scope := newScope()
		node := &ast.UnionDecl{
			Name: ast.NewIdent("Result", token.NoPos),
			Args: []*ast.Ident{
				ast.NewIdent("a", token.NoPos),
				ast.NewIdent("a", token.NoPos),
			},
			Ctors: []*ast.Constructor{
				&ast.Constructor{
					Name: ast.NewIdent("Foo", token.NoPos),
				},
			},
		}
		r.resolveDecl(scope, node)

		require.False(r.reporter.IsOK())
		assertReports(t, r.reporter, new(report.RepeatedVarTypeError))
	})

	t.Run("UnionDecl repeated ctor", func(t *testing.T) {
		r := newTestResolver(t)
		require := require.New(t)
		scope := newScope()
		node := &ast.UnionDecl{
			Name: ast.NewIdent("Cmp", token.NoPos),
			Ctors: []*ast.Constructor{
				&ast.Constructor{Name: ast.NewIdent("Gt", token.NoPos)},
				&ast.Constructor{Name: ast.NewIdent("Gt", token.NoPos)},
			},
		}
		r.resolveDecl(scope, node)

		require.False(r.reporter.IsOK())
		assertReports(t, r.reporter, new(report.RepeatedCtorError))
	})
}

func TestResolveQualifiedName(t *testing.T) {
	parent := ast.NewModuleScope(nil)
	fooBarBazMod := &ast.Module{
		Scope: modScopeWithObjects(
			ast.NewObject("Qux", ast.Typ, nil),
		),
	}
	parent.ImportModule(ast.NewObject("Foo.Bar.Baz", ast.Mod, fooBarBazMod))

	fooBarMod := &ast.Module{
		Scope: modScopeWithObjects(
			ast.NewObject("qux", ast.Var, nil),
			ast.NewObject("Gux", ast.Ctor, nil),
		),
	}
	parent.ImportModule(ast.NewObject("Foo.Bar", ast.Mod, fooBarMod))
	scope := ast.NewNodeScope(nil, parent)
	r := newTestResolver(t)

	fooBarBazPath := []*ast.Ident{
		ast.NewIdent("Foo", token.NoPos),
		ast.NewIdent("Bar", token.NoPos),
		ast.NewIdent("Baz", token.NoPos),
	}
	fooBarPath := []*ast.Ident{
		ast.NewIdent("Foo", token.NoPos),
		ast.NewIdent("Bar", token.NoPos),
	}

	t.Run("Type", func(t *testing.T) {
		ident := ast.NewIdent("Qux", token.NoPos)
		node := ast.NewSelectorExpr(append(fooBarBazPath, ident)...)
		r.resolveQualifiedName(scope, node, ast.Typ)

		for _, id := range fooBarBazPath {
			assertObj(t, id, "Foo.Bar.Baz")
		}
		assertObj(t, ident, "Qux")
		require.True(t, r.reporter.IsOK())
	})

	t.Run("Var", func(t *testing.T) {
		ident := ast.NewIdent("qux", token.NoPos)
		node := ast.NewSelectorExpr(append(fooBarPath, ident)...)
		r.resolveQualifiedName(scope, node, ast.Var)

		for _, id := range fooBarPath {
			assertObj(t, id, "Foo.Bar")
		}
		assertObj(t, ident, "qux")
		require.True(t, r.reporter.IsOK())
	})

	t.Run("Ctor", func(t *testing.T) {
		ident := ast.NewIdent("Gux", token.NoPos)
		node := ast.NewSelectorExpr(append(fooBarPath, ident)...)
		r.resolveQualifiedName(scope, node, ast.Var)

		for _, id := range fooBarPath {
			assertObj(t, id, "Foo.Bar")
		}
		assertObj(t, ident, "Gux")
		require.True(t, r.reporter.IsOK())
	})

	t.Run("Var field", func(t *testing.T) {
		ident := ast.NewIdent("qux", token.NoPos)
		field := ast.NewIdent("f", token.NoPos)
		node := ast.NewSelectorExpr(append(fooBarPath, ident, field)...)
		r.resolveQualifiedName(scope, node, ast.Var)

		for _, id := range fooBarPath {
			assertObj(t, id, "Foo.Bar")
		}
		assertObj(t, ident, "qux")
		require.Nil(t, field.Obj)
		require.True(t, r.reporter.IsOK())
	})

	t.Run("Module not imported", func(t *testing.T) {
		r := newTestResolver(t)
		node := ast.NewSelectorExpr(
			ast.NewIdent("Foo", token.NoPos),
			ast.NewIdent("bar", token.NoPos),
		)
		r.resolveQualifiedName(scope, node, ast.Var)

		assertReports(t, r.reporter, new(report.ModuleNotImportedError))
		require.False(t, r.reporter.IsOK())
	})

	t.Run("Import error", func(t *testing.T) {
		r := newTestResolver(t)
		node := ast.NewSelectorExpr(append(fooBarPath, ast.NewIdent("fux", token.NoPos))...)
		r.resolveQualifiedName(scope, node, ast.Var)

		assertReports(t, r.reporter, new(report.ImportError))
		require.False(t, r.reporter.IsOK())
	})
}

func TestResolveExpr(t *testing.T) {
	r := newTestResolver(t)
	newScope := func() *ast.NodeScope {
		return scopeWithObjects(
			ast.NewObject("a", ast.Var, nil),
			ast.NewObject("b", ast.Var, nil),
			ast.NewObject("c", ast.Var, nil),
			ast.NewObject("Just", ast.Ctor, nil),
			ast.NewObject("Nothing", ast.Ctor, nil),
			ast.NewObject("-", ast.Var, nil),
			ast.NewObject("+", ast.Var, nil),
		)
	}

	t.Run("IfExpr", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.IfExpr{
			Cond:     ast.NewIdent("a", token.NoPos),
			ThenExpr: ast.NewIdent("b", token.NoPos),
			ElseExpr: ast.NewIdent("c", token.NoPos),
		}

		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Cond, "a")
		assertObj(t, node.ThenExpr, "b")
		assertObj(t, node.ElseExpr, "c")
		require.True(r.reporter.IsOK())
	})

	t.Run("CaseExpr", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.CaseExpr{
			Expr: ast.NewIdent("a", token.NoPos),
			Branches: []*ast.CaseBranch{
				{
					Pattern: &ast.CtorPattern{
						Ctor: ast.NewIdent("Just", token.NoPos),
						Args: []ast.Pattern{
							&ast.VarPattern{ast.NewIdent("d", token.NoPos)},
						},
					},
					Expr: ast.NewIdent("d", token.NoPos),
				},
				{
					Pattern: &ast.CtorPattern{
						Ctor: ast.NewIdent("Nothing", token.NoPos),
					},
					Expr: ast.NewIdent("c", token.NoPos),
				},
			},
		}

		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Expr, "a")
		assertObj(t, node.Branches[0].Pattern.(*ast.CtorPattern).Ctor, "Just")
		assertObj(t, node.Branches[0].Expr, "d")
		assertObj(t, node.Branches[1].Pattern.(*ast.CtorPattern).Ctor, "Nothing")
		assertObj(t, node.Branches[1].Expr, "c")
		require.True(r.reporter.IsOK())
	})

	t.Run("LetExpr", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()

		xDecl := &ast.Definition{
			Name: ast.NewIdent("x", token.NoPos),
			Body: ast.NewIdent("a", token.NoPos),
		}

		yDecl := &ast.Definition{
			Name: ast.NewIdent("y", token.NoPos),
			Body: ast.NewIdent("b", token.NoPos),
		}

		node := &ast.LetExpr{
			Decls: []ast.Decl{
				xDecl,
				yDecl,
			},
			Body: &ast.BinaryOp{
				Op:  ast.NewIdent("+", token.NoPos),
				Lhs: ast.NewIdent("y", token.NoPos),
				Rhs: ast.NewIdent("x", token.NoPos),
			},
		}
		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)

		require.Len(scope.Children(), 1)
		scope = scope.Children()[0]
		require.Len(scope.Objects, 2)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["x"])
		require.NotNil(scope.Objects["y"])

		expr := node.Body.(*ast.BinaryOp)
		assertObj(t, expr.Op, "+")
		assertObj(t, expr.Rhs, "x")
		assertObj(t, expr.Lhs, "y")
		require.True(r.reporter.IsOK())
	})

	t.Run("TupleLit", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.TupleLit{
			Elems: []ast.Expr{
				ast.NewIdent("a", token.NoPos),
				ast.NewIdent("b", token.NoPos),
				ast.NewIdent("c", token.NoPos),
			},
		}

		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Elems[0], "a")
		assertObj(t, node.Elems[1], "b")
		assertObj(t, node.Elems[2], "c")
		require.True(r.reporter.IsOK())
	})

	t.Run("ListLit", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.ListLit{
			Elems: []ast.Expr{
				ast.NewIdent("a", token.NoPos),
				ast.NewIdent("b", token.NoPos),
				ast.NewIdent("c", token.NoPos),
			},
		}

		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Elems[0], "a")
		assertObj(t, node.Elems[1], "b")
		assertObj(t, node.Elems[2], "c")
		require.True(r.reporter.IsOK())
	})

	t.Run("FuncApp", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.FuncApp{
			Func: ast.NewIdent("a", token.NoPos),
			Args: []ast.Expr{
				ast.NewIdent("b", token.NoPos),
				ast.NewIdent("c", token.NoPos),
			},
		}

		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Func, "a")
		assertObj(t, node.Args[0], "b")
		assertObj(t, node.Args[1], "c")
		require.True(r.reporter.IsOK())
	})

	t.Run("RecordLit", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.RecordLit{
			Fields: []*ast.FieldAssign{
				{Field: ast.NewIdent("x", token.NoPos), Expr: ast.NewIdent("a", token.NoPos)},
				{Field: ast.NewIdent("y", token.NoPos), Expr: ast.NewIdent("b", token.NoPos)},
			},
		}
		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		require.Nil(node.Fields[0].Field.Obj)
		assertObj(t, node.Fields[0].Expr, "a")
		require.Nil(node.Fields[1].Field.Obj)
		assertObj(t, node.Fields[1].Expr, "b")
		require.True(r.reporter.IsOK())
	})

	t.Run("RecordLit repeated field", func(t *testing.T) {
		require := require.New(t)
		r := newTestResolver(t)
		scope := newScope()
		node := &ast.RecordLit{
			Fields: []*ast.FieldAssign{
				{Field: ast.NewIdent("x", token.NoPos), Expr: ast.NewIdent("a", token.NoPos)},
				{Field: ast.NewIdent("x", token.NoPos)},
			},
		}
		r.resolveExpr(scope, node)

		require.False(r.reporter.IsOK())
		assertReports(t, r.reporter, new(report.RepeatedFieldError))
	})

	t.Run("RecordUpdate", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.RecordUpdate{
			Record: ast.NewIdent("c", token.NoPos),
			Fields: []*ast.FieldAssign{
				{Field: ast.NewIdent("x", token.NoPos), Expr: ast.NewIdent("a", token.NoPos)},
				{Field: ast.NewIdent("y", token.NoPos), Expr: ast.NewIdent("b", token.NoPos)},
			},
		}
		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Record, "c")
		require.Nil(node.Fields[0].Field.Obj)
		assertObj(t, node.Fields[0].Expr, "a")
		require.Nil(node.Fields[1].Field.Obj)
		assertObj(t, node.Fields[1].Expr, "b")
		require.True(r.reporter.IsOK())
	})

	t.Run("RecordUpdate repeated field", func(t *testing.T) {
		require := require.New(t)
		r := newTestResolver(t)
		scope := newScope()
		node := &ast.RecordUpdate{
			Record: ast.NewIdent("c", token.NoPos),
			Fields: []*ast.FieldAssign{
				{Field: ast.NewIdent("x", token.NoPos), Expr: ast.NewIdent("a", token.NoPos)},
				{Field: ast.NewIdent("x", token.NoPos)},
			},
		}
		r.resolveExpr(scope, node)

		require.False(r.reporter.IsOK())
		assertReports(t, r.reporter, new(report.RepeatedFieldError))
	})

	t.Run("UnaryOp", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.UnaryOp{
			Op:   ast.NewIdent("-", token.NoPos),
			Expr: ast.NewIdent("a", token.NoPos),
		}

		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Expr, "a")
		assertObj(t, node.Op, "-")
		require.True(r.reporter.IsOK())
	})

	t.Run("BinaryOp", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.BinaryOp{
			Op:  ast.NewIdent("-", token.NoPos),
			Lhs: ast.NewIdent("a", token.NoPos),
			Rhs: ast.NewIdent("b", token.NoPos),
		}
		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Lhs, "a")
		assertObj(t, node.Rhs, "b")
		assertObj(t, node.Op, "-")
		require.True(r.reporter.IsOK())
	})

	t.Run("Lambda", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.Lambda{
			Args: []ast.Pattern{
				&ast.VarPattern{ast.NewIdent("x", token.NoPos)},
				&ast.VarPattern{ast.NewIdent("y", token.NoPos)},
			},
			Expr: &ast.BinaryOp{
				Op:  ast.NewIdent("+", token.NoPos),
				Lhs: ast.NewIdent("x", token.NoPos),
				Rhs: ast.NewIdent("y", token.NoPos),
			},
		}

		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)

		require.Len(scope.Children(), 1)
		scope = scope.Children()[0]
		require.Len(scope.Objects, 2)
		require.Len(scope.Unresolved, 0)
		require.NotNil(scope.Objects["x"])
		require.NotNil(scope.Objects["y"])

		require.Nil(node.Args[0].(*ast.VarPattern).Name.Obj)
		require.Nil(node.Args[1].(*ast.VarPattern).Name.Obj)
		expr := node.Expr.(*ast.BinaryOp)
		assertObj(t, expr.Op, "+")
		assertObj(t, expr.Lhs, "x")
		assertObj(t, expr.Rhs, "y")
		require.True(r.reporter.IsOK())
	})

	t.Run("ParensExpr", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.ParensExpr{
			Expr: ast.NewIdent("a", token.NoPos),
		}

		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		assertObj(t, node.Expr, "a")
		require.True(r.reporter.IsOK())
	})

	t.Run("AccessorExpr", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.AccessorExpr{
			Field: ast.NewIdent("foo", token.NoPos),
		}

		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		require.Nil(node.Field.Obj)
		require.True(r.reporter.IsOK())
	})

	t.Run("TupleCtor", func(t *testing.T) {
		require := require.New(t)
		scope := newScope()
		node := &ast.TupleCtor{Elems: 1}

		r.resolveExpr(scope, node)

		require.Len(scope.Objects, 0)
		require.Len(scope.Unresolved, 0)
		require.True(r.reporter.IsOK())
	})
}

func TestResolveImport(t *testing.T) {
	cases := []struct {
		name     string
		decl     *ast.ImportDecl
		imported []string
		reports  []report.Report
	}{
		{
			"open list",
			&ast.ImportDecl{
				Module:   ast.NewIdent("Foo", token.NoPos),
				Alias:    ast.NewIdent("FooAlias", token.NoPos),
				Exposing: new(ast.OpenList),
			},
			[]string{"foo", "bar", "Cmp"},
			nil,
		},
		{
			"closed list 2 exposed vars",
			&ast.ImportDecl{
				Module: ast.NewIdent("Foo", token.NoPos),
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedVar{ast.NewIdent("foo", token.NoPos)},
						&ast.ExposedVar{ast.NewIdent("bar", token.NoPos)},
					},
				},
			},
			[]string{"foo", "bar"},
			nil,
		},
		{
			"closed list no exposed var",
			&ast.ImportDecl{
				Module: ast.NewIdent("Foo", token.NoPos),
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedVar{ast.NewIdent("foo", token.NoPos)},
						&ast.ExposedVar{ast.NewIdent("baz", token.NoPos)},
					},
				},
			},
			nil,
			[]report.Report{new(report.ImportError)},
		},
		{
			"closed list wrong exposed var",
			&ast.ImportDecl{
				Module: ast.NewIdent("Foo", token.NoPos),
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedVar{ast.NewIdent("foo", token.NoPos)},
						&ast.ExposedVar{ast.NewIdent("Eq", token.NoPos)},
					},
				},
			},
			nil,
			[]report.Report{new(report.ImportError)},
		},
		{
			"exposed union is not union",
			&ast.ImportDecl{
				Module: ast.NewIdent("Foo", token.NoPos),
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type:  ast.NewIdent("Eq", token.NoPos),
							Ctors: new(ast.OpenList),
						},
					},
				},
			},
			nil,
			[]report.Report{new(report.ImportError)},
		},
		{
			"exposed union open list",
			&ast.ImportDecl{
				Module: ast.NewIdent("Foo", token.NoPos),
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type:  ast.NewIdent("Cmp", token.NoPos),
							Ctors: new(ast.OpenList),
						},
					},
				},
			},
			[]string{"Cmp", "Eq", "Lt", "Gt"},
			nil,
		},
		{
			"exposed union closed list",
			&ast.ImportDecl{
				Module: ast.NewIdent("Foo", token.NoPos),
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type: ast.NewIdent("Cmp", token.NoPos),
							Ctors: &ast.ClosedList{
								Exposed: []ast.ExposedIdent{
									&ast.ExposedVar{ast.NewIdent("Eq", token.NoPos)},
									&ast.ExposedVar{ast.NewIdent("Gt", token.NoPos)},
								},
							},
						},
					},
				},
			},
			[]string{"Cmp", "Eq", "Gt"},
			nil,
		},
		{
			"exposed union closed list import error",
			&ast.ImportDecl{
				Module: ast.NewIdent("Foo", token.NoPos),
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type: ast.NewIdent("Cmp", token.NoPos),
							Ctors: &ast.ClosedList{
								Exposed: []ast.ExposedIdent{
									&ast.ExposedVar{ast.NewIdent("Baz", token.NoPos)},
									&ast.ExposedVar{ast.NewIdent("Gt", token.NoPos)},
								},
							},
						},
					},
				},
			},
			nil,
			[]report.Report{new(report.ImportError)},
		},
		{
			"exposed union closed list wrong type",
			&ast.ImportDecl{
				Module: ast.NewIdent("Foo", token.NoPos),
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type: ast.NewIdent("Cmp", token.NoPos),
							Ctors: &ast.ClosedList{
								Exposed: []ast.ExposedIdent{
									&ast.ExposedVar{ast.NewIdent("Cmp", token.NoPos)},
									&ast.ExposedVar{ast.NewIdent("Gt", token.NoPos)},
								},
							},
						},
					},
				},
			},
			nil,
			[]report.Report{new(report.ImportError)},
		},
		{
			"exposed union type does not exist",
			&ast.ImportDecl{
				Module: ast.NewIdent("Foo", token.NoPos),
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type:  ast.NewIdent("Qux", token.NoPos),
							Ctors: new(ast.OpenList),
						},
					},
				},
			},
			nil,
			[]report.Report{new(report.ImportError)},
		},
	}

	fooScope := ast.NewModuleScope(nil)
	fooScope.Expose(ast.NewObject("foo", ast.Var, nil))
	fooScope.Expose(ast.NewObject("bar", ast.Var, nil))
	fooScope.Expose(ast.NewObject("Cmp", ast.Typ, &ast.UnionDecl{
		Ctors: []*ast.Constructor{
			{Name: ast.NewIdent("Eq", token.NoPos)},
			{Name: ast.NewIdent("Lt", token.NoPos)},
			{Name: ast.NewIdent("Gt", token.NoPos)},
		},
	}))
	fooScope.Expose(ast.NewObject("Eq", ast.Ctor, &ast.Constructor{Name: ast.NewIdent("Eq", token.NoPos)}))
	fooScope.Expose(ast.NewObject("Gt", ast.Ctor, &ast.Constructor{Name: ast.NewIdent("Gt", token.NoPos)}))
	fooScope.Expose(ast.NewObject("Lt", ast.Ctor, &ast.Constructor{Name: ast.NewIdent("Lt", token.NoPos)}))
	pkg := &ast.Package{
		Modules: map[string]*ast.Module{
			"Foo": &ast.Module{
				Scope: fooScope,
			},
		},
	}

	for _, c := range cases {
		r := newTestResolver(t)
		r.pkg = pkg

		t.Run(c.name, func(t *testing.T) {
			require := require.New(t)
			scope := ast.NewModuleScope(nil)
			r.resolveImport(scope, c.decl)

			if c.decl.Alias != nil {
				require.NotNil(scope.Modules[c.decl.Alias.Name])
			}

			require.NotNil(scope.Modules[c.decl.ModuleName()])
			if len(c.reports) > 0 {
				assertReports(t, r.reporter, c.reports...)
				require.False(r.reporter.IsOK())
			} else {
				for _, i := range c.imported {
					require.NotNil(scope.Imported[i], "expected %s to be imported", i)
				}
				require.Len(scope.Imported, len(c.imported))
				require.True(r.reporter.IsOK())
			}
		})
	}
}

func TestResolveModuleDecl(t *testing.T) {
	cases := []struct {
		name     string
		decl     *ast.ModuleDecl
		exported []string
		reports  []report.Report
	}{
		{
			"open list",
			&ast.ModuleDecl{
				Exposing: new(ast.OpenList),
			},
			[]string{"foo", "bar", "Cmp", "Eq", "Lt", "Gt"},
			nil,
		},
		{
			"closed list 2 exposed vars",
			&ast.ModuleDecl{
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedVar{ast.NewIdent("foo", token.NoPos)},
						&ast.ExposedVar{ast.NewIdent("bar", token.NoPos)},
					},
				},
			},
			[]string{"foo", "bar"},
			nil,
		},
		{
			"closed list no exposed var",
			&ast.ModuleDecl{
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedVar{ast.NewIdent("foo", token.NoPos)},
						&ast.ExposedVar{ast.NewIdent("baz", token.NoPos)},
					},
				},
			},
			nil,
			[]report.Report{new(report.ExportError)},
		},
		{
			"closed list wrong exposed var",
			&ast.ModuleDecl{
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedVar{ast.NewIdent("foo", token.NoPos)},
						&ast.ExposedVar{ast.NewIdent("Eq", token.NoPos)},
					},
				},
			},
			nil,
			[]report.Report{new(report.ExportError)},
		},
		{
			"exposed union is not union",
			&ast.ModuleDecl{
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type:  ast.NewIdent("Eq", token.NoPos),
							Ctors: new(ast.OpenList),
						},
					},
				},
			},
			nil,
			[]report.Report{new(report.ExportError)},
		},
		{
			"exposed union open list",
			&ast.ModuleDecl{
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type:  ast.NewIdent("Cmp", token.NoPos),
							Ctors: new(ast.OpenList),
						},
					},
				},
			},
			[]string{"Cmp", "Eq", "Lt", "Gt"},
			nil,
		},
		{
			"exposed union closed list",
			&ast.ModuleDecl{
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type: ast.NewIdent("Cmp", token.NoPos),
							Ctors: &ast.ClosedList{
								Exposed: []ast.ExposedIdent{
									&ast.ExposedVar{ast.NewIdent("Eq", token.NoPos)},
									&ast.ExposedVar{ast.NewIdent("Gt", token.NoPos)},
								},
							},
						},
					},
				},
			},
			[]string{"Cmp", "Eq", "Gt"},
			nil,
		},
		{
			"exposed union closed list import error",
			&ast.ModuleDecl{
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type: ast.NewIdent("Cmp", token.NoPos),
							Ctors: &ast.ClosedList{
								Exposed: []ast.ExposedIdent{
									&ast.ExposedVar{ast.NewIdent("Baz", token.NoPos)},
									&ast.ExposedVar{ast.NewIdent("Gt", token.NoPos)},
								},
							},
						},
					},
				},
			},
			nil,
			[]report.Report{new(report.ExportError)},
		},
		{
			"exposed union closed list wrong type",
			&ast.ModuleDecl{
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type: ast.NewIdent("Cmp", token.NoPos),
							Ctors: &ast.ClosedList{
								Exposed: []ast.ExposedIdent{
									&ast.ExposedVar{ast.NewIdent("Cmp", token.NoPos)},
									&ast.ExposedVar{ast.NewIdent("Gt", token.NoPos)},
								},
							},
						},
					},
				},
			},
			nil,
			[]report.Report{new(report.ExportError)},
		},
		{
			"exposed union type does not exist",
			&ast.ModuleDecl{
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type:  ast.NewIdent("Qux", token.NoPos),
							Ctors: new(ast.OpenList),
						},
					},
				},
			},
			nil,
			[]report.Report{new(report.ExportError)},
		},
		{
			"exposed union type is not union",
			&ast.ModuleDecl{
				Exposing: &ast.ClosedList{
					Exposed: []ast.ExposedIdent{
						&ast.ExposedUnion{
							Type:  ast.NewIdent("bar", token.NoPos),
							Ctors: new(ast.OpenList),
						},
					},
				},
			},
			nil,
			[]report.Report{new(report.ExpectedUnionError)},
		},
	}

	for _, c := range cases {
		r := newTestResolver(t)

		scope := ast.NewModuleScope(&ast.Module{
			Module: c.decl,
		})
		scope.Add(ast.NewObject("foo", ast.Var, ast.NewIdent("foo", token.NoPos)))
		scope.Add(ast.NewObject("bar", ast.Var, ast.NewIdent("bar", token.NoPos)))
		scope.Add(ast.NewObject("Cmp", ast.Typ, &ast.UnionDecl{
			Ctors: []*ast.Constructor{
				{Name: ast.NewIdent("Eq", token.NoPos)},
				{Name: ast.NewIdent("Lt", token.NoPos)},
				{Name: ast.NewIdent("Gt", token.NoPos)},
			},
		}))
		scope.Add(ast.NewObject("Eq", ast.Ctor, &ast.Constructor{Name: ast.NewIdent("Eq", token.NoPos)}))
		scope.Add(ast.NewObject("Gt", ast.Ctor, &ast.Constructor{Name: ast.NewIdent("Gt", token.NoPos)}))
		scope.Add(ast.NewObject("Lt", ast.Ctor, &ast.Constructor{Name: ast.NewIdent("Lt", token.NoPos)}))

		t.Run(c.name, func(t *testing.T) {
			require := require.New(t)
			r.resolveModuleDecl(scope, c.decl)

			if len(c.reports) > 0 {
				assertReports(t, r.reporter, c.reports...)
				require.False(r.reporter.IsOK())
			} else {
				for _, i := range c.exported {
					require.NotNil(scope.Exposed[i], "expected %s to be exported", i)
				}
				require.Len(scope.Exposed, len(c.exported))
				require.True(r.reporter.IsOK())
			}
		})
	}
}

func assertReports(t *testing.T, r *report.Reporter, reports ...report.Report) {
	reps := r.Reports("test")
	require.Len(t, reps, len(reports), "incorrect number of reports")
	for i := range reports {
		require.IsType(t, reports[i], reps[i], "incorrect report type for report number %d", i)
	}
}

func assertObj(t *testing.T, node ast.Node, name string) {
	ident, ok := node.(*ast.Ident)
	require.True(t, ok, "expected node to be ident with name %s", name)
	require.NotNil(t, ident.Obj, "expected node to have object with name %s", name)
	require.Equal(t, name, ident.Obj.Name, "expected ident object to be %s", name)
}

func scopeWithObjects(objs ...*ast.Object) *ast.NodeScope {
	parent := ast.NewModuleScope(nil)
	for _, obj := range objs {
		parent.Add(obj)
	}
	return ast.NewNodeScope(nil, parent)
}

func modScopeWithObjects(objs ...*ast.Object) *ast.ModuleScope {
	parent := ast.NewModuleScope(nil)
	for _, obj := range objs {
		parent.Add(obj)
	}
	return parent
}

func newTestResolver(t *testing.T) *resolver {
	path := "test"
	loader := source.NewMemLoader()
	loader.Add(path, "")
	cm := source.NewCodeMap(loader)
	require.NoError(t, cm.Add(path), "adding %s", path)
	reporter := report.NewReporter(cm, report.Stderr(true, true))
	return &resolver{nil, reporter, path}
}
