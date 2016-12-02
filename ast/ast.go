package ast

import (
	"strings"

	"github.com/mvader/elmo/token"
)

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

// Decl is a declaration node.
type Decl interface {
	Node
	isDecl()
}

// ModuleName is an identifier made of one or more identifiers
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

// ExposedIdent is an identifier exposed in an import or module declaration.
// It can as well expose more identifiers in the case of union types.
type ExposedIdent struct {
	*Ident
	// Exposing will contain all the exposed identifiers of this particular
	// exposed identifier. Only union types will have this.
	Exposing *ExposingList
}

func (i *ExposedIdent) Pos() token.Pos { return i.Pos() }
func (i *ExposedIdent) End() token.Pos { return i.Exposing.End() }

// ExposingList is a list of exposed identifiers delimited by parenthesis.
type ExposingList struct {
	Idents []*ExposedIdent
	Lparen token.Pos
	Rparen token.Pos
}

func (l *ExposingList) Pos() token.Pos { return l.Lparen }
func (l *ExposingList) End() token.Pos { return l.Rparen }

// ModuleDecl is a node representing a module declaration and contains the
// name of the module and the identifiers it exposes, if any.
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

// ImportDecl is a node representing an import declaration. It contains the
// imported module as well as its alias, if any, and the exposed identifiers,
// if any.
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

// Ident represents an identifier, which is a name for something.
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

// BasicLit represents a basic literal.
type BasicLit struct {
	Pos   *token.Position
	Type  BasicLitType
	Value string
}

func (*BasicLit) isExpr() {}

// BasicLitType is the type of a literal.
type BasicLitType byte

const (
	// Error is an invalid literal.
	Error BasicLitType = iota
	// Int is an integer literal.
	Int
	// Float is a floating point number literal.
	Float
	// String is a string literal.
	String
	// Bool is a boolean literal.
	Bool
	// Char is a character literal.
	Char
)

// InfixDecl is a node representing the declaration of an operator's fixity.
// It contains the operator, the priority given and the associativity of the
// operator.
type InfixDecl struct {
	InfixPos token.Pos
	Assoc    Associativity
	Op       *Ident
	Priority *BasicLit
}

// Associativity of the operator.
type Associativity byte

const (
	// NonAssoc is a non associative operator.
	NonAssoc Associativity = iota
	// LeftAssoc is a left associative operator.
	LeftAssoc
	// RightAssoc is a right associative operator.
	RightAssoc
)

func (InfixDecl) isDecl()          {}
func (d InfixDecl) Pos() token.Pos { return d.InfixPos }
func (d InfixDecl) End() token.Pos { return d.Op.End() }

// AliasDecl is a node representing a type alias declaration. It contains
// the name of the alias and its arguments along with the type it is aliasing.
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

// UnionDecl is a node representing an union type declaration. Contains
// the name of the union type, the arguments and all the constructors for
// the type.
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

// Constructor is a node representing the constructor of an union type.
// It contains the name of the constructor and its type arguments.
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

// Type is a node representing a type.
type Type interface {
	Node
	isType()
}

// BasicType is a type that has a name and 0 or more arguments.
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

type FuncType struct {
	Args   []Type
	Return Type
}

func (FuncType) isType()          {}
func (t FuncType) Pos() token.Pos { return t.Args[0].Pos() }
func (t FuncType) End() token.Pos { return t.Return.End() }

// RecordType is a node representing a record type.
type RecordType struct {
	Lbrace token.Pos
	Rbrace token.Pos
	Fields []*RecordTypeField
}

func (RecordType) isType()          {}
func (t RecordType) Pos() token.Pos { return t.Lbrace }
func (t RecordType) End() token.Pos { return t.Rbrace }

// RecordTypeField represents a field in a record type node.
type RecordTypeField struct {
	Name  *Ident
	Type  Type
	Colon token.Pos
	Comma token.Pos
}

// TupleType is a node representing a tuple with two ore more types.
type TupleType struct {
	Lparen token.Pos
	Rparen token.Pos
	Elems  []Type
}

func (TupleType) isType()          {}
func (t TupleType) Pos() token.Pos { return t.Lparen }
func (t TupleType) End() token.Pos { return t.Rparen }

// Definition is a node representing a definition of a value. A definition can
// also be annotated with a type annotation.
type Definition struct {
	Annotation *TypeAnnotation
	Name       *Ident
	Assign     token.Pos
	Args       []*Ident
	Body       Expr
}

// TypeAnnotation is the annotation of a declaration with its type.
type TypeAnnotation struct {
	Name  *Ident
	Colon token.Pos
	Type  Type
}

// Expr is an expression node.
type Expr interface {
	Node
	isExpr()
}
