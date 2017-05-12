package types

import "github.com/elm-tangram/tangram/ast"

func convert(mod *ast.Module) (*State, []Expr) {
	s := NewState()
	env := s.env

	var decls []ast.Decl
	for _, d := range mod.Decls {
		switch d := d.(type) {
		case *ast.UnionDecl:
			vars := make([]Type, len(d.Args))
			for i, arg := range d.Args {
				vars[i] = &VarType{arg.Name, false}
			}

			typ := &NamedType{
				Name: &Var{d, d.Name.Name, ""},
				Args: vars,
			}

			for _, ctor := range d.Ctors {
				var t Type
				if len(ctor.Args > 0) {
					args := make([]Type, len(ctor.Args))
					for i, arg := range ctor.Args {
						args[i] = convertType(arg)
					}

					t = &FuncType{
						Args:   args,
						Return: typ,
					}
				} else {
					t = typ
				}

				env.SetVar(ctor.Name.Name, t)
			}

		case *ast.AliasDecl:
			env.SetType(d.Name.Name, convertType(d.Type))

		case *ast.DestructuringAssignment, *ast.Definition:
			decls = append(decls, d)
		}
	}

	var result = make([]Expr, len(decls))
	for i, decl := range decls {
		result[i] = convertDecl(decl)
	}
	return result
}

func generalize(env *TypeEnv, typ Type) *Scheme {
	var vars []string
	for _, v := range typ.freeTypeVars() {
		if !env.freeVars.contains(v) {
			vars = append(vars, v)
		}
	}
	return &Scheme{vars, typ}
}

func instantiate(env *TypeEnv, scheme *Scheme) Type {
	typ := scheme.Type
	for _, v := range scheme.Vars {
		typ = typ.replaceVar(v, env.newVar(v))
	}
	return typ
}

func unify(ta, tb Type) error {
	panic("not implemented")
}

func inferExpr(env *TypeEnv, expr Expr) Type {
	switch expr := expr.(type) {
	case *Var:
	case *FieldVar:
	case *App:
	case *Let:
	case *Abs:
	case *Def:
	case *RecordUpdate:
	case *BinOp:
	case Pattern:
		return inferPattern(env, expr)
	case Lit:
		return inferLit(env, expr)
	}
	panic("unreachable")
}

func inferPattern(env *TypeEnv, expr Type, pat Pattern) Type {
	switch pat := pat.(type) {
	case *Var:
		t := generalize(expr)
		env.elems[pat.Name] = t
		return t

	case *TuplePattern:

	case *RecordPattern:

	case *ListPattern:

	case *CtorPattern:

	case *AnythingPattern:
		return &VarType{env.newVar("a")}

	case *AliasPattern:
		typ := inferPattern(env, expr, pat.Pattern)
		env.elems[pat.Name.Name] = generalize(env, typ)
		return typ
	}
	panic("unreachable")
}

func inferLit(env *TypeEnv, lit Lit) Type {
	switch lit := lit.(type) {
	case *BasicLit:
		return &NamedType{
			Name: &Var{lit.BasicLit, lit.Type.String(), ""},
		}

	case *ListLit:
		elems := make([]Type, len(lit.Elems))
		for i, el := range lit.Elems {
			elems[i] = inferExpr(env, expr)
		}
		return &NamedType{
			&Var{lit.Node, "List", "List"},
			elems,
		}

	case *TupleLit:
		elems := make([]Type, len(lit.Elems))
		for i, el := range lit.Elems {
			elems[i] = inferExpr(env, expr)
		}
		return &TupleType{elems}

	case *RecordLit:
		fields := make([]*FieldType, len(lit.Fields))
		for i, f := range lit.Fields {
			fields[i] = &FieldType{
				Name: f.Name,
				Type: inferExpr(env, f.Value),
			}
		}
		return &RecordType{fields}
	}
	panic("unreachable")
}
