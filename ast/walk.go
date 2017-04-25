package ast

import "fmt"

// Visitor is used to traverse an AST. Its method Visit will be invoked with
// every single node to visit during the traversal.
type Visitor interface {
	// Visit will be invoked with the node being visited. If it receives a
	// non-nil node and returns nil, the children of that node will not be
	// visited.
	// If it's called with a nil node, it means all children of a node have
	// been visited.
	Visit(Node) Visitor
}

func walkDecls(v Visitor, decls []Decl) {
	for _, d := range decls {
		Walk(v, d)
	}
}

func walkPatterns(v Visitor, patterns []Pattern) {
	for _, p := range patterns {
		Walk(v, p)
	}
}

func walkExprs(v Visitor, exprs []Expr) {
	for _, e := range exprs {
		Walk(v, e)
	}
}

// Walk traveses an AST with the given node as a starting point in depth-first
// order. The first thing it will do is v.Visit(node) and, if the result is not
// nil, it will continue walking every non-nil child.
// After all the children have been visited, v.Visit(nil) will be performed so
// the visitor knows that all children have been visited.
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch node := node.(type) {
	case *Module:
		if node.Module != nil {
			Walk(v, node.Module)
		}

		for _, i := range node.Imports {
			Walk(v, i)
		}

		walkDecls(v, node.Decls)

	// Decls
	case *ModuleDecl:
		Walk(v, node.Name)
		if node.Exposing != nil {
			Walk(v, node.Exposing)
		}

	case *ImportDecl:
		Walk(v, node.Module)
		if node.Alias != nil {
			Walk(v, node.Alias)
		}

		if node.Exposing != nil {
			Walk(v, node.Exposing)
		}

	case *ClosedList:
		for _, ident := range node.Exposed {
			Walk(v, ident)
		}

	case *OpenList:
		// do nothing

	case *ExposedVar:
		Walk(v, node.Ident)

	case *ExposedUnion:
		Walk(v, node.Type)
		Walk(v, node.Ctors)

	case *InfixDecl:
		Walk(v, node.Op)
		Walk(v, node.Precedence)

	case *AliasDecl:
		Walk(v, node.Name)
		for _, a := range node.Args {
			Walk(v, a)
		}

		Walk(v, node.Type)

	case *UnionDecl:
		Walk(v, node.Name)
		for _, a := range node.Args {
			Walk(v, a)
		}

		for _, t := range node.Types {
			Walk(v, t)
		}

	case *Constructor:
		Walk(v, node.Name)
		for _, a := range node.Args {
			Walk(v, a)
		}

	case *DestructuringAssignment:
		Walk(v, node.Pattern)
		Walk(v, node.Expr)

	case *Definition:
		if node.Annotation != nil {
			Walk(v, node.Annotation)
		}

		Walk(v, node.Name)
		for _, a := range node.Args {
			Walk(v, a)
		}

		Walk(v, node.Body)

	case *TypeAnnotation:
		Walk(v, node.Name)
		Walk(v, node.Type)

	// Types
	case *NamedType:
		Walk(v, node.Name)
		for _, a := range node.Args {
			Walk(v, a)
		}

	case *VarType:
		Walk(v, node.Ident)

	case *FuncType:
		for _, a := range node.Args {
			Walk(v, a)
		}

		Walk(v, node.Return)

	case *RecordType:
		for _, f := range node.Fields {
			Walk(v, f)
		}

	case *RecordField:
		Walk(v, node.Name)
		Walk(v, node.Type)

	case *TupleType:
		for _, el := range node.Elems {
			Walk(v, el)
		}

	// Patterns
	case *VarPattern:
		Walk(v, node.Name)

	case *AnythingPattern:
		// nothing to do

	case *LiteralPattern:
		Walk(v, node.Literal)

	case *AliasPattern:
		Walk(v, node.Pattern)
		Walk(v, node.Name)

	case *CtorPattern:
		Walk(v, node.Ctor)
		for _, p := range node.Patterns {
			Walk(v, p)
		}

	case *TuplePattern:
		walkPatterns(v, node.Patterns)

	case *RecordPattern:
		walkPatterns(v, node.Patterns)

	case *ListPattern:
		walkPatterns(v, node.Patterns)

	// Exprs
	case *Ident, *BasicLit:
		// do nothing

	case *SelectorExpr:
		Walk(v, node.Selector)
		Walk(v, node.Expr)

	case *TupleLit:
		walkExprs(v, node.Elems)

	case *FuncApp:
		Walk(v, node.Func)
		walkExprs(v, node.Args)

	case *RecordLit:
		for _, f := range node.Fields {
			Walk(v, f)
		}

	case *FieldAssign:
		Walk(v, node.Field)
		Walk(v, node.Expr)

	case *RecordUpdate:
		Walk(v, node.Record)
		for _, f := range node.Fields {
			Walk(v, f)
		}

	case *LetExpr:
		walkDecls(v, node.Decls)
		Walk(v, node.Body)

	case *IfExpr:
		Walk(v, node.Cond)
		Walk(v, node.ThenExpr)
		Walk(v, node.ElseExpr)

	case *CaseExpr:
		Walk(v, node.Expr)
		for _, b := range node.Branches {
			Walk(v, b)
		}

	case *CaseBranch:
		Walk(v, node.Pattern)
		Walk(v, node.Expr)

	case *Lambda:
		walkPatterns(v, node.Args)
		Walk(v, node.Expr)

	case *ListLit:
		walkExprs(v, node.Elems)

	case *UnaryOp:
		Walk(v, node.Op)
		Walk(v, node.Expr)

	case *BinaryOp:
		Walk(v, node.Op)
		Walk(v, node.Lhs)
		Walk(v, node.Rhs)

	case *AccessorExpr:
		Walk(v, node.Field)

	case *TupleCtor:
		// nothing to do

	case *ParensExpr:
		Walk(v, node.Expr)

	default:
		panic(fmt.Errorf("walk: unable to walk node of type %T", node))
	}

	v.Visit(nil)
}

type inspector func(Node) bool

func (i inspector) Visit(node Node) Visitor {
	if i(node) {
		return i
	}
	return nil
}

// WalkFunc traverses the AST in depth-first order. It works exactly the same
// way Walk does, only it uses a function to walk the AST instead of a Visitor.
func WalkFunc(node Node, fn func(Node) bool) {
	Walk(inspector(fn), node)
}
