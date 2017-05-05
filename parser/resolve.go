package parser

import (
	"strings"

	"github.com/elm-tangram/tangram/ast"
	"github.com/elm-tangram/tangram/report"
)

type resolver struct {
	pkg      *ast.Package
	reporter *report.Reporter

	path string
}

func (r *resolver) resolve(pkg *ast.Package) bool {
	r.pkg = pkg
	var resolved = true
	for _, m := range pkg.Order {
		r.path = pkg.Modules[m].Path
		resolved = r.resolveModule(pkg.Modules[m]) && resolved
	}
	return resolved
}

func (r *resolver) resolveModule(mod *ast.Module) bool {
	mod.Scope = ast.NewModuleScope(mod)

	for _, imp := range mod.Imports {
		r.resolveImport(mod.Scope, imp)
	}

	for _, decl := range mod.Decls {
		r.resolveDecl(mod.Scope, decl)
	}

	r.resolveModuleDecl(mod.Scope, mod.Module)
	return r.checkUnresolved(mod.Scope)
}

// TODO(erizocosmico): please, split this into smaller functions
func (r *resolver) resolveImport(scope *ast.ModuleScope, imp *ast.ImportDecl) {
	mod := imp.ModuleName()
	isNative := strings.HasPrefix(mod, "Native.")
	var kind = ast.Mod
	if isNative {
		kind = ast.NativeMod
	}
	obj := ast.NewObject(mod, kind, imp)
	scope.ImportModule(obj)

	if imp.Alias != nil {
		alias := *obj
		alias.Name = imp.Alias.Name
		scope.ImportModule(&alias)
	}

	if isNative {
		if imp.Exposing != nil {
			r.report(report.NewBaseReport(report.SyntaxError, imp.Exposing.Pos(), "Native modules cannot expose anything.", report.RegionFromNode(imp)))
		}
		return
	}

	importScope := r.pkg.Modules[mod].Scope
	obj.Node = r.pkg.Modules[mod]
	switch exp := imp.Exposing.(type) {
	case *ast.ClosedList:
	Outer:
		for _, id := range exp.Exposed {
			switch id := id.(type) {
			case *ast.ExposedVar:
				if obj := importScope.LookupExposed(id.Name, ast.Var); obj != nil {
					switch obj.Kind {
					case ast.Typ, ast.Var:
						scope.Import(obj)
					default:
						r.report(report.NewImportError(imp, imp.ModuleName(), id.Ident))
					}
				} else {
					r.report(report.NewImportError(imp, imp.ModuleName(), id.Ident))
				}
			case *ast.ExposedUnion:
				if obj := importScope.LookupExposed(id.Type.Name, ast.Typ); obj != nil {
					union, ok := obj.Node.(*ast.UnionDecl)
					if !ok {
						r.report(report.NewExpectedUnionError(imp, obj))
						continue Outer
					}

					scope.Import(obj)
					switch exp := id.Ctors.(type) {
					case *ast.ClosedList:
						for _, id := range exp.Exposed {
							switch id := id.(type) {
							case *ast.ExposedVar:
								if obj := importScope.LookupExposed(id.Name, ast.Ctor); obj != nil {
									if obj.Kind == ast.Ctor {
										scope.Import(obj)
									} else {
										r.report(report.NewExpectedCtorError(imp, obj))
									}
								} else {
									r.report(report.NewImportError(imp, imp.ModuleName(), id.Ident))
								}
							default:
								// unreachable
								panic("constructors cannot expose anything")
							}
						}
					case *ast.OpenList:
						for _, t := range union.Ctors {
							if obj := importScope.LookupExposed(t.Name.Name, ast.Ctor); obj != nil {
								scope.Import(obj)
							}
						}
					}
				} else {
					r.report(report.NewImportError(imp, imp.ModuleName(), id.Type))
				}
			}
		}
	case *ast.OpenList:
		for _, obj := range importScope.Exposed {
			switch obj.Kind {
			case ast.Typ, ast.Var:
				scope.Import(obj)
			}
		}
	}
}

// TODO: add again VarTyp resolution to decls, a lookup is enough
// because they must be previously declared
// TODO: check when adding a new type to the top-level that is not already declared.
func (r *resolver) resolveDecl(scope ast.Scope, decl ast.Decl) {
	switch decl := decl.(type) {
	case *ast.DestructuringAssignment:
		r.resolvePattern(scope, decl.Pattern)
		r.resolveExpr(scope, decl.Expr)
	case *ast.InfixDecl:
		r.resolveExpr(scope, decl.Op)
	case *ast.Definition:
		if decl.Annotation != nil {
			r.resolveType(scope, decl.Annotation.Type, false)
		}
		scope.Add(ast.NewObject(decl.Name.Name, ast.Var, decl.Name))

		defScope := ast.NewNodeScope(decl, scope)
		for _, arg := range decl.Args {
			r.resolvePattern(defScope, arg)
		}
		r.resolveExpr(defScope, decl.Body)
	case *ast.AliasDecl:
		scope.Add(ast.NewObject(decl.Name.Name, ast.Typ, decl))
		declScope := ast.NewNodeScope(decl, scope)
		set := make(map[string]struct{})
		for _, arg := range decl.Args {
			if _, ok := set[arg.Name]; ok {
				r.report(report.NewRepeatedVarTypeError(decl, arg))
				return
			}
			set[arg.Name] = struct{}{}
			declScope.Add(ast.NewObject(arg.Name, ast.VarTyp, arg))
		}
		r.resolveType(declScope, decl.Type, true)
	case *ast.UnionDecl:
		scope.Add(ast.NewObject(decl.Name.Name, ast.Typ, decl))
		declScope := ast.NewNodeScope(decl, scope)
		set := make(map[string]struct{})
		for _, arg := range decl.Args {
			if _, ok := set[arg.Name]; ok {
				r.report(report.NewRepeatedVarTypeError(decl, arg))
				return
			}
			set[arg.Name] = struct{}{}
			declScope.Add(ast.NewObject(arg.Name, ast.VarTyp, arg))
		}

		set = make(map[string]struct{})
		for _, ctor := range decl.Ctors {
			if _, ok := set[ctor.Name.Name]; ok {
				r.report(report.NewRepeatedCtorError(decl, ctor.Name))
				return
			}
			set[ctor.Name.Name] = struct{}{}
			r.resolveCtor(scope, declScope, ctor)
		}
	}
}

// TODO(erizocosmico): please, split this into smaller functions
func (r *resolver) resolveModuleDecl(scope *ast.ModuleScope, mod *ast.ModuleDecl) {
	switch list := mod.Exposing.(type) {
	case *ast.OpenList:
		for _, obj := range scope.Objects {
			switch obj.Kind {
			case ast.Typ, ast.Var, ast.Ctor:
				scope.Expose(obj)
			}
		}
	case *ast.ClosedList:
	Outer:
		for _, ident := range list.Exposed {
			switch exposed := ident.(type) {
			case *ast.ExposedVar:
				r.tryExpose(scope, exposed.Ident)
			case *ast.ExposedUnion:
				if obj := r.tryExpose(scope, exposed.Type); obj != nil {
					union, ok := obj.Node.(*ast.UnionDecl)
					if !ok {
						r.report(report.NewExpectedUnionError(mod, obj))
						continue Outer
					}

					switch list := exposed.Ctors.(type) {
					case *ast.OpenList:
						for _, t := range union.Ctors {
							r.tryExposeCtor(scope, t.Name)
						}
					case *ast.ClosedList:
						for _, id := range list.Exposed {
							if v, ok := id.(*ast.ExposedVar); ok {
								if ctor := union.LookupCtor(v.Name); ctor != nil {
									r.tryExposeCtor(scope, ctor.Name)
								} else {
									r.report(report.NewExportError(mod, v.Ident))
								}
							} else {
								// unreachable
								panic("constructors cannot expose anything")
							}
						}
					default:
						// unreachable
						panic("union must expose something")
					}
				}
			}
		}
	}
}

func (r *resolver) tryExpose(scope *ast.ModuleScope, ident *ast.Ident) *ast.Object {
	if obj := scope.LookupSelf(ident.Name, ast.Var); obj != nil {
		scope.Expose(obj)
		return obj
	}

	if obj := scope.LookupSelf(ident.Name, ast.Typ); obj != nil {
		scope.Expose(obj)
		return obj
	}

	r.report(report.NewExportError(scope.Root.(*ast.Module).Module, ident))
	return nil
}

func (r *resolver) tryExposeCtor(scope *ast.ModuleScope, ident *ast.Ident) *ast.Object {
	if obj := scope.LookupSelf(ident.Name, ast.Ctor); obj != nil {
		scope.Expose(obj)
		return obj
	}

	r.report(report.NewExportError(scope.Root.(*ast.Module).Module, ident))
	return nil
}

func (r *resolver) resolveCtor(outerScope, declScope ast.Scope, ctor *ast.Constructor) {
	// TODO: check is not already defined in scope
	outerScope.Add(ast.NewObject(ctor.Name.Name, ast.Ctor, ctor))
	for _, arg := range ctor.Args {
		r.resolveType(declScope, arg, true)
	}
}

func (r *resolver) resolveExpr(scope ast.Scope, expr ast.Expr) {
	switch expr := expr.(type) {
	case *ast.Ident:
		r.resolveQualifiedName(scope, expr, ast.Var)
	case *ast.SelectorExpr:
		r.resolveQualifiedName(scope, expr, ast.Var)
	case *ast.IfExpr:
		r.resolveExpr(scope, expr.Cond)
		r.resolveExpr(scope, expr.ThenExpr)
		r.resolveExpr(scope, expr.ElseExpr)
	case *ast.CaseExpr:
		r.resolveExpr(scope, expr.Expr)
		for _, b := range expr.Branches {
			branchScope := ast.NewNodeScope(b, scope)
			r.resolvePattern(branchScope, b.Pattern)
			r.resolveExpr(branchScope, b.Expr)
		}
	case *ast.LetExpr:
		letScope := ast.NewNodeScope(expr, scope)
		for _, d := range expr.Decls {
			r.resolveDecl(letScope, d)
		}
		r.resolveExpr(letScope, expr.Body)
	case *ast.TupleLit:
		for _, el := range expr.Elems {
			r.resolveExpr(scope, el)
		}
	case *ast.ListLit:
		for _, el := range expr.Elems {
			r.resolveExpr(scope, el)
		}
	case *ast.FuncApp:
		r.resolveExpr(scope, expr.Func)
		for _, arg := range expr.Args {
			r.resolveExpr(scope, arg)
		}
	case *ast.RecordLit:
		var set = make(map[string]struct{})
		for _, f := range expr.Fields {
			if _, ok := set[f.Field.Name]; ok {
				r.report(report.NewRepeatedFieldError(expr, f.Field))
				return
			}
			set[f.Field.Name] = struct{}{}
			r.resolveExpr(scope, f.Expr)
		}
	case *ast.RecordUpdate:
		r.resolveExpr(scope, expr.Record)
		var set = make(map[string]struct{})
		for _, f := range expr.Fields {
			if _, ok := set[f.Field.Name]; ok {
				r.report(report.NewRepeatedFieldError(expr, f.Field))
				return
			}
			set[f.Field.Name] = struct{}{}
			r.resolveExpr(scope, f.Expr)
		}
	case *ast.UnaryOp:
		r.resolveExpr(scope, expr.Op)
		r.resolveExpr(scope, expr.Expr)
	case *ast.BinaryOp:
		r.resolveExpr(scope, expr.Op)
		r.resolveExpr(scope, expr.Lhs)
		r.resolveExpr(scope, expr.Rhs)
	case *ast.Lambda:
		lambdaScope := ast.NewNodeScope(expr, scope)
		for _, arg := range expr.Args {
			r.resolvePattern(lambdaScope, arg)
		}
		r.resolveExpr(lambdaScope, expr.Expr)
	case *ast.ParensExpr:
		r.resolveExpr(scope, expr.Expr)
	case *ast.AccessorExpr, *ast.TupleCtor, *ast.BadExpr:
		// no need to do anything
	}
}

func (r *resolver) resolvePattern(scope ast.Scope, pattern ast.Pattern) {
	switch pattern := pattern.(type) {
	case *ast.AliasPattern:
		scope.Add(ast.NewObject(pattern.Name.Name, ast.Var, pattern.Pattern))
		r.resolvePattern(scope, pattern.Pattern)
	case *ast.CtorPattern:
		r.resolveQualifiedName(scope, pattern.Ctor, ast.Var)
		for _, p := range pattern.Args {
			r.resolvePattern(scope, p)
		}
	case *ast.TuplePattern:
		for _, el := range pattern.Elems {
			r.resolvePattern(scope, el)
		}
	case *ast.ListPattern:
		for _, el := range pattern.Elems {
			r.resolvePattern(scope, el)
		}
	case *ast.RecordPattern:
		for _, el := range pattern.Fields {
			r.resolvePattern(scope, el)
		}
	case *ast.VarPattern:
		scope.Add(ast.NewObject(pattern.Name.Name, ast.Var, pattern))
	case *ast.LiteralPattern, *ast.AnythingPattern:
		// no need to do anything
	}
}

// TODO: ability to pass a node to get better snippets on reports
// pass the Annotation, TypeDecl or UnionDecl.
func (r *resolver) resolveType(scope ast.Scope, typ ast.Type, resolveVars bool) {
	switch typ := typ.(type) {
	case *ast.NamedType:
		r.resolveQualifiedName(scope, typ.Name, ast.Typ)
		for _, arg := range typ.Args {
			r.resolveType(scope, arg, resolveVars)
		}
	case *ast.VarType:
		if resolveVars {
			if obj := scope.Lookup(typ.Name, ast.VarTyp); obj != nil {
				typ.Obj = obj
			} else {
				r.report(report.NewUndefinedTypeVarError(typ, typ))
			}
		}
	case *ast.FuncType:
		for _, arg := range typ.Args {
			r.resolveType(scope, arg, resolveVars)
		}
		r.resolveType(scope, typ.Return, resolveVars)
	case *ast.RecordType:
		var idents = make(map[string]struct{})
		for _, f := range typ.Fields {
			if _, ok := idents[f.Name.Name]; ok {
				r.report(report.NewRepeatedFieldError(typ, f.Name))
				return
			}
			idents[f.Name.Name] = struct{}{}
			r.resolveType(scope, f.Type, resolveVars)
		}
	case *ast.TupleType:
		for _, el := range typ.Elems {
			r.resolveType(scope, el, resolveVars)
		}
	}
}

// resolveQualifiedName resolves either Idents and SelectorExprs.
// The kind used to resolve a qualified name must be Typ, Var or VarTyp.
// If the kind is Var and the name is uppercase, it will automatically
// look for a Ctor instead.
func (r *resolver) resolveQualifiedName(scope ast.Scope, expr ast.Expr, kind ast.ObjKind) {
	var path []*ast.Ident
	var varIdent *ast.Ident

	exp := expr
	for varIdent == nil && exp != nil {
		switch e := exp.(type) {
		case *ast.Ident:
			if isUpper(e.Name) {
				path = append(path, e)
			} else {
				varIdent = e
			}
			exp = nil
		case *ast.SelectorExpr:
			if isUpper(e.Selector.Name) {
				path = append(path, e.Selector)
			} else {
				varIdent = e.Selector
			}
			exp = e.Expr
		}
	}

	if varIdent == nil {
		varIdent = path[len(path)-1]
		path = path[:len(path)-1]
	}

	if isUpper(varIdent.Name) && kind == ast.Var {
		kind = ast.Ctor
	}

	var modName string
	if len(path) > 0 {
		var parts = make([]string, len(path))
		for i, id := range path {
			parts[i] = id.Name
		}
		modName = strings.Join(parts, ".")

		if obj := scope.Lookup(modName, ast.Mod); obj != nil {
			for _, id := range path {
				id.Obj = obj
			}

			if obj.Kind == ast.NativeMod {
				return
			}

			scope = obj.Node.(*ast.Module).Scope
		} else {
			r.report(report.NewModuleNotImportedError(expr, modName))
			return
		}
	}

	if obj := scope.Lookup(varIdent.Name, kind); obj != nil {
		varIdent.Obj = obj
	} else {
		if len(path) > 0 {
			r.report(report.NewImportError(expr, modName, varIdent))
		} else {
			scope.Resolve(varIdent.Name, varIdent, kind)
		}
	}
}

func (r *resolver) checkUnresolved(scope *ast.ModuleScope) bool {
	var resolved = true
	r.resolveBasicTypes(scope.Unresolved)
	if len(scope.Unresolved) > 0 {
		r.reportUnresolved(scope.Unresolved)
		resolved = false
	}

	return r.checkUnresolvedChildren(scope.Children()) && resolved
}

var basicTypes = map[string]*ast.Object{
	"Int":    ast.NewObject("Int", ast.BuiltinTyp, nil),
	"Float":  ast.NewObject("Float", ast.BuiltinTyp, nil),
	"Bool":   ast.NewObject("Bool", ast.BuiltinTyp, nil),
	"String": ast.NewObject("String", ast.BuiltinTyp, nil),
	"Char":   ast.NewObject("Char", ast.BuiltinTyp, nil),
	"List":   ast.NewObject("List", ast.BuiltinTyp, nil),
}

func (r *resolver) checkUnresolvedChildren(scopes []*ast.NodeScope) bool {
	var resolved = true
	for _, scope := range scopes {
		r.resolveBasicTypes(scope.Unresolved)
		if len(scope.Unresolved) > 0 {
			r.reportUnresolved(scope.Unresolved)
			resolved = false
		}
	}

	return resolved
}

func (r *resolver) resolveBasicTypes(unresolved map[string][]*ast.Ident) {
	for k, idents := range unresolved {
		if obj, ok := basicTypes[k]; ok {
			for _, id := range idents {
				id.Obj = obj
			}
			delete(unresolved, k)
		}
	}
}

func (r *resolver) reportUnresolved(unresolved map[string][]*ast.Ident) {
	for name, idents := range unresolved {
		for _, ident := range idents {
			r.report(report.NewUnresolvedNameError(name, ident))
		}
	}
}

func (r *resolver) report(report report.Report) {
	r.reporter.Report(r.path, report)
}

func isNativeImport(module string) bool {
	return strings.HasPrefix(module, "Native.")
}
