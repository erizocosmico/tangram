package ast

import "github.com/erizocosmico/elmo/token"

// Node is a node in the AST.
type Node interface {
	// Pos is the starting position of the node.
	Pos() token.Pos
	// End is the position of the ending of the node.
	End() token.Pos
}

// Module is the AST representation of a module, that is, a single source code
// file.
type Module struct {
	Name    string
	Module  *ModuleDecl
	Imports []*ImportDecl
	Decls   []Decl
	Scope   *ModuleScope
}

func (f *Module) Pos() token.Pos { return f.Module.Pos() }
func (f *Module) End() token.Pos {
	if len(f.Decls) > 0 {
		return f.Decls[len(f.Decls)-1].End()
	}

	if len(f.Imports) > 0 {
		return f.Imports[len(f.Imports)-1].End()
	}

	return f.Module.End()
}

// Package is the set of modules with a certain order of resolution that
// conform a package.
type Package struct {
	// Order in which modules should be resolved.
	Order []string
	// Modules is a mapping between a module name and its module AST
	// representation.
	Modules map[string]*Module
}
