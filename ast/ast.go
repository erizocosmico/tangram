package ast

import (
	"strings"

	"github.com/mvader/elm-compiler/token"
)

type File struct {
	Name    string
	Module  *ModuleDecl
	Imports []*ImportDecl
	Decls   []Decl
}

type Node interface {
	Pos() token.Pos
	End() token.Pos
}

type Decl interface {
	Node
	isDecl()
}

type ModuleName []*Ident

func (n ModuleName) Pos() token.Pos {
	if len(n) == 0 {
		return token.NoPos
	}
	return n[0].Pos()
}

func (n ModuleName) End() token.Pos {
	if len(n) == 0 {
		return token.NoPos
	}
	return n[len(n)-1].End()
}

func (n ModuleName) String() string {
	var parts = make([]string, len(n))
	for i, p := range n {
		parts[i] = p.Name
	}
	return strings.Join(parts, ".")
}

type ExposedIdent struct {
	*Ident
	Exposing *ExposingList
}

func (i *ExposedIdent) Pos() token.Pos { return i.Pos() }
func (i *ExposedIdent) End() token.Pos { return i.Exposing.End() }

type ExposingList struct {
	Idents []*ExposedIdent
	Lparen token.Pos
	Rparen token.Pos
}

func (l *ExposingList) Pos() token.Pos { return l.Lparen }
func (l *ExposingList) End() token.Pos { return l.Rparen }

type ModuleDecl struct {
	Name     ModuleName
	Module   token.Pos
	Exposing *ExposingList
}

func (d *ModuleDecl) Pos() token.Pos { return d.Module }
func (d *ModuleDecl) End() token.Pos {
	if d.Exposing == nil {
		return d.Name[len(d.Name)-1].End()
	}
	return d.Exposing.End()
}
func (d *ModuleDecl) isDecl() {}

type ImportDecl struct {
	Module   ModuleName
	Alias    *Ident
	Import   token.Pos
	Exposing *ExposingList
}

func (d *ImportDecl) Pos() token.Pos { return d.Import }
func (d *ImportDecl) End() token.Pos {
	if d.Exposing == nil {
		return d.Module[len(d.Module)-1].End()
	}
	return d.Exposing.End()
}
func (d *ImportDecl) isDecl() {}

type Ident struct {
	NamePos *token.Position
	Name    string
	Obj     *Object
}

func (i *Ident) Pos() token.Pos { return i.NamePos.Offset }
func (i *Ident) End() token.Pos { return i.Pos() + token.Pos(len(i.Name)) }

type Object struct {
	Kind ObjKind
	Name string
	Decl interface{}
}

type ObjKind uint

const (
	Bad ObjKind = iota
	Mod
	Var
	Typ
	Fun
	Op
)
