package ast

import "github.com/erizocosmico/elmo/token"

// File is the AST representation of a source code file.
type File struct {
	Name    string
	Module  *ModuleDecl
	Imports []*ImportDecl
	Decls   []Decl
}

// Node is a node in the AST.
type Node interface {
	// Pos is the starting position of the node.
	Pos() token.Pos
	// End is the position of the ending of the node.
	End() token.Pos
}
