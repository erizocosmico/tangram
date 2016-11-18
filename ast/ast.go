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

type BasicLit struct {
	Pos   *token.Position
	Type  BasicLitType
	Value string
}

type BasicLitType byte

const (
	Error BasicLitType = iota
	Int
	Float
	String
	Bool
	Char
)

type InfixDecl struct {
	InfixPos token.Pos
	Assoc    Associativity
	Op       *Ident
	Priority *BasicLit
}

type Associativity byte

const (
	NonAssoc Associativity = iota
	LeftAssoc
	RightAssoc
)

func (InfixDecl) isDecl()          {}
func (d InfixDecl) Pos() token.Pos { return d.InfixPos }
func (d InfixDecl) End() token.Pos { return d.Op.End() }

type AliasDecl struct {
	TypePos token.Pos
	Alias   token.Pos
	Eq      token.Pos
	Name    *Ident
	Args    []*Ident
	Type    Type
}

func (d AliasDecl) isDecl()        {}
func (d AliasDecl) Pos() token.Pos { return d.TypePos }
func (d AliasDecl) End() token.Pos { return d.Type.Pos() }

type UnionDecl struct {
	TypePos token.Pos
	Eq      token.Pos
	Name    *Ident
	Args    []*Ident
	Types   []*Constructor
}

func (d UnionDecl) isDecl()        {}
func (d UnionDecl) Pos() token.Pos { return d.TypePos }
func (d UnionDecl) End() token.Pos {
	if len(d.Types) == 0 {
		return token.NoPos
	}
	return d.Types[len(d.Types)-1].End()
}

type Constructor struct {
	Name *Ident
	Args []Type
	Pipe token.Pos
}

func (c Constructor) Pos() token.Pos {
	if c.Pipe != token.NoPos {
		return c.Pipe
	}
	return c.Name.Pos()
}
func (c Constructor) End() token.Pos {
	if len(c.Args) > 0 {
		return c.Args[len(c.Args)-1].End()
	}
	return c.Name.End()
}

type Type interface {
	Node
	isType()
}

type ParenthesizedType struct {
	Lparen token.Pos
	Rparen token.Pos
	Type   Type
}

func (ParenthesizedType) isType()          {}
func (t ParenthesizedType) Pos() token.Pos { return t.Lparen }
func (t ParenthesizedType) End() token.Pos { return t.Rparen }

type BasicType struct {
	Name *Ident
	Args []Type
}

func (BasicType) isType()          {}
func (t BasicType) Pos() token.Pos { return t.Name.Pos() }
func (t BasicType) End() token.Pos {
	if len(t.Args) > 0 {
		return t.Args[len(t.Args)-1].End()
	}

	return t.Name.End()
}

type RecordType struct {
	Lbrace token.Pos
	Rbrace token.Pos
	Fields []*RecordTypeField
}

func (RecordType) isType()          {}
func (t RecordType) Pos() token.Pos { return t.Lbrace }
func (t RecordType) End() token.Pos { return t.Rbrace }

type RecordTypeField struct {
	Name  *Ident
	Type  Type
	Colon token.Pos
	Comma token.Pos
}

type TupleType struct {
	Lparen token.Pos
	Rparen token.Pos
	Elems  []*TupleElem
}

func (TupleType) isType()          {}
func (t TupleType) Pos() token.Pos { return t.Lparen }
func (t TupleType) End() token.Pos { return t.Rparen }

type TupleElem struct {
	Type  Type
	Comma token.Pos
}
