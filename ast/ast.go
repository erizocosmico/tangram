package ast

import "github.com/erizocosmico/elmo/token"

// File is the AST representation of a source code file.
type File struct {
	Name    string
	Module  *ModuleDecl
	Imports []*ImportDecl
	Decls   []Decl
}

func (f *File) Pos() token.Pos { return f.Module.Pos() }
func (f *File) End() token.Pos {
	if len(f.Decls) > 0 {
		return f.Decls[len(f.Decls)-1].End()
	}

	if len(f.Imports) > 0 {
		return f.Imports[len(f.Imports)-1].End()
	}

	return f.Module.End()
}

// Node is a node in the AST.
type Node interface {
	// Pos is the starting position of the node.
	Pos() token.Pos
	// End is the position of the ending of the node.
	End() token.Pos
}
