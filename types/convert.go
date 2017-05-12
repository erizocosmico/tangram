package types

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/elm-tangram/tangram/ast"
)

func convertType(typ ast.Type) Type {
	switch typ := typ.(type) {
	case *ast.VarType:
		return &VarType{typ.Name}

	case *ast.NamedType:
		var v *Var
		switch n := typ.Name.(type) {
		case *ast.Ident:
			v = &Var{typ.Name, typ.Name.Name, ""}
		case *ast.SelectorExpr:
			v = convertSelector(n).(*Var)
		}

		args := make([]Type, len(typ.Args))
		for i, arg := range typ.Args {
			args[i] = convertType(arg)
		}

		return &NamedType{v, args}

	case *ast.FuncType:
		args := make([]Type, len(typ.Args))
		for i, arg := range typ.Args {
			args[i] = convertType(arg)
		}
		return &FuncType{args, convertType(typ.Return)}

	case *ast.RecordType:
		fields := make([]Type, len(typ.Fields))
		for i, f := range typ.Fields {
			fields[i] = FieldType{
				f.Name.Name,
				convertType(f.Type),
			}
		}
		return &RecordType{fields}

	case *ast.TupleType:
		elems := make([]Type, len(typ.Elems))
		for i, el := range elems {
			elems[i] = convertType(el)
		}
		return &TupleType{elems}
	}

	panic("unreachable")
}

func convertDecl(decl ast.Decl) Expr {
	switch decl := decl.(type) {
	case *ast.DestructuringAssignment:
		return &Def{decl, convertPattern(decl.Pattern), convertExpr(decl.Expr)}

	case *ast.Definition:
		var body Expr
		if len(decl.Args) > 0 {
			args := make([]Pattern, len(decl.Args))
			for i, arg := range decl.Args {
				args[i] = convertPattern(arg)
			}
			body = &Abs{decl, args, convertExpr(decl.Body)}
		} else {
			body = convertExpr(decl.Body)
		}

		return &Def{
			decl,
			&Var{decl.Name, decl.Name.Name, ""},
			body,
		}
	}

	panic("unreachable")
}

func convertExpr(expr ast.Expr) Expr {
	switch expr := expr.(type) {
	case *ast.Lambda:
		var args = make([]Pattern, len(expr.Args))
		for i, a := range expr.Args {
			args[i] = convertPattern(a)
		}
		return &Abs{
			Node: expr,
			Args: args,
			Body: convertExpr(expr.Expr),
		}

	case *ast.Ident:
		return &Var{expr, expr.Name, ""}

	case *ast.SelectorExpr:
		return convertSelector(expr)

	case *ast.AccessorExpr:
		record := &Var{expr, "record", ""}
		return &Abs{
			Node: expr,
			Args: []Pattern{record},
			Body: &FieldVar{
				Node:   expr,
				Name:   expr.Field.Name,
				Record: record,
			},
		}

	case *ast.IfExpr:
		return &If{
			Cond: convertExpr(expr.Cond),
			Then: convertExpr(expr.ThenExpr),
			Else: convertExpr(expr.ElseExpr),
		}

	case *ast.CaseExpr:
		branches := make([]*CaseBranch, len(expr.Branches))
		for i, b := range expr.Branches {
			branches[i] = convertBranch(b)
		}
		return &Case{
			Expr:     convertExpr(expr.Expr),
			Branches: branches,
		}

	case *ast.LetExpr:
		var defs = make([]Expr, len(expr.Decls))
		for i, d := range expr.Decls {
			defs[i] = convertDecl(d)
		}
		return &Let{
			Node: expr,
			Defs: defs,
			Body: convertExpr(expr.Body),
		}

	case *ast.ListLit, *ast.TupleLit, *ast.RecordLit, *ast.BasicLit:
		return convertLit(expr)

	case *ast.RecordUpdate:
		fields := make([]Field, len(expr.Fields))
		for i, f := range expr.Fields {
			fields[i] = Field{
				Name:  f.Field.Name,
				Value: convertExpr(f.Expr),
			}
		}
		return &RecordUpdate{expr, convertExpr(expr.Record), fields}

	case *ast.FuncApp:
		args := make([]Expr, len(expr.Args))
		for i, a := range expr.Args {
			args[i] = convertExpr(a)
		}
		return &App{
			Node: expr,
			Func: convertExpr(expr.Func),
			Args: args,
		}

	case *ast.UnaryOp:
		return &App{
			Node: expr,
			Func: convertExpr(expr.Op),
			Args: []Expr{
				convertExpr(expr.Expr),
			},
		}

	case *ast.BinaryOp:
		return &App{
			Node: expr,
			Func: convertExpr(expr.Op),
			Args: []Expr{
				convertExpr(expr.Lhs),
				convertExpr(expr.Rhs),
			},
		}

	case *ast.ParensExpr:
		return convertExpr(expr.Expr)

	case *ast.TupleCtor:
		args := make([]Pattern, expr.Elems)
		elems := make([]Expr, expr.Elems)
		for i := 0; i < expr.Elems; i++ {
			v := &Var{expr, fmt.Sprintf("e%d", i+1), ""}
			args[i] = v
			elems[i] = v
		}

		return &Abs{
			Node: expr,
			Args: args,
			Body: &TupleLit{expr, elems},
		}

	case *ast.BadExpr:
		// TODO: handle
	}

	panic("unrachable")
}

func convertSelector(sel *ast.SelectorExpr) Expr {
	var upperNames []*ast.Ident
	var lowerNames []*ast.Ident
	var expr ast.Expr = sel

	for expr != nil {
		switch e := expr.(type) {
		case *ast.SelectorExpr:
			if isUpper(e.Selector.Name) {
				upperNames = append(upperNames, e.Selector)
			} else {
				lowerNames = append(lowerNames, e.Selector)
			}

			expr = e.Expr
		case *ast.Ident:
			if isUpper(e.Name) {
				upperNames = append(upperNames, e)
			} else {
				lowerNames = append(lowerNames, e)
			}
			expr = nil
		default:
			break
		}
	}

	if len(lowerNames) == 0 {
		name := upperNames[len(upperNames)-1]
		upperNames = upperNames[:len(upperNames)-1]
		selector := joinIdents(upperNames)

		return &Var{name, name.Name, selector}
	}

	selector := joinIdents(upperNames)
	varName := lowerNames[0]
	lowerNames = lowerNames[1:]

	var v Expr = &Var{varName, varName.Name, selector}
	for _, n := range lowerNames {
		v = &FieldVar{n, n.Name, v}
	}

	return v
}

func joinIdents(idents []*ast.Ident) string {
	var strs = make([]string, len(idents))
	for i, id := range idents {
		strs[i] = id.Name
	}
	return strings.Join(strs, ".")
}

func isUpper(str string) bool {
	return unicode.IsUpper(rune(str[0]))
}

func convertPattern(pat ast.Pattern) Pattern {
	switch pat := pat.(type) {
	case *ast.VarPattern:
		return &Var{pat, pat.Name.Name, ""}

	case *ast.ListPattern:
		elems := make([]Pattern, len(pat.Elems))
		for i, el := range pat.Elems {
			elems[i] = convertPattern(el)
		}
		return &ListPattern{pat, elems}

	case *ast.CtorPattern:
		args := make([]Pattern, len(pat.Args))
		for i, arg := range pat.Args {
			args[i] = convertPattern(arg)
		}
		return &CtorPattern{
			pat,
			convertExpr(pat.Ctor),
			args,
		}

	case *ast.TuplePattern:
		elems := make([]Pattern, len(pat.Elems))
		for i, el := range pat.Elems {
			elems[i] = convertPattern(el)
		}
		return &TuplePattern{pat, elems}

	case *ast.RecordPattern:
		fields := make([]Pattern, len(pat.Fields))
		for i, el := range pat.Fields {
			fields[i] = convertPattern(el)
		}
		return &RecordPattern{pat, fields}

	case *ast.AliasPattern:
		return &AliasPattern{
			pat,
			&Var{pat.Name, pat.Name.Name, ""},
			convertPattern(pat.Pattern),
		}

	case *ast.AnythingPattern:
		return &AnythingPattern{pat}
	}

	panic("unreachable")
}

func convertBranch(branch *ast.CaseBranch) *CaseBranch {
	return &CaseBranch{
		branch,
		convertPattern(branch.Pattern),
		convertExpr(branch.Expr),
	}
}

func convertLit(lit ast.Expr) Expr {
	switch lit := lit.(type) {
	case *ast.BasicLit:
		return &BasicLit{lit, basicLitTypes[lit.Type], lit.Value}

	case *ast.RecordLit:
		fields := make([]Field, len(lit.Fields))
		for i, f := range lit.Fields {
			fields[i] = Field{
				Name:  f.Field.Name,
				Value: convertExpr(f.Expr),
			}
		}
		return &RecordLit{lit, fields}

	case *ast.TupleLit:
		elems := make([]Expr, len(lit.Elems))
		for i, el := range lit.Elems {
			elems[i] = convertExpr(el)
		}
		return &TupleLit{lit, elems}

	case *ast.ListLit:
		elems := make([]Expr, len(lit.Elems))
		for i, el := range lit.Elems {
			elems[i] = convertExpr(el)
		}
		return &ListLit{lit, elems}
	}

	panic("unreachable")
}

var basicLitTypes = map[ast.BasicLitType]BasicLitType{
	ast.Int:    Int,
	ast.String: String,
	ast.Char:   Char,
	ast.Bool:   Bool,
	ast.Float:  Float,
}
