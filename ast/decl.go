package ast

import "github.com/erizocosmico/elmo/token"

// Decl is a declaration node.
type Decl interface {
	Node
	isDecl()
}

// ExposedIdent is an identifier exposed in an import or module declaration.
// It can as well expose more identifiers in the case of union types.
type ExposedIdent struct {
	*Ident
	// Exposing will contain all the exposed identifiers of this particular
	// exposed identifier. Only union types will have this.
	Exposing *ExposingList
}

func NewExposedIdent(ident *Ident) *ExposedIdent {
	return &ExposedIdent{Ident: ident}
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
	Name     Expr
	Module   token.Pos
	Exposing *ExposingList
}

func (d *ModuleDecl) Pos() token.Pos { return d.Module }
func (d *ModuleDecl) End() token.Pos {
	if d.Exposing == nil {
		return d.Name.End()
	}
	return d.Exposing.End()
}
func (d *ModuleDecl) isDecl() {}

// ImportDecl is a node representing an import declaration. It contains the
// imported module as well as its alias, if any, and the exposed identifiers,
// if any.
type ImportDecl struct {
	Module   Expr
	Alias    *Ident
	Import   token.Pos
	Exposing *ExposingList
}

func (d *ImportDecl) Pos() token.Pos { return d.Import }
func (d *ImportDecl) End() token.Pos {
	if d.Exposing == nil {
		return d.Module.End()
	}
	return d.Exposing.End()
}
func (d *ImportDecl) isDecl() {}

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

type DestructuringAssignment struct {
	Assign  token.Pos
	Pattern Pattern
	Expr    Expr
}

func (a *DestructuringAssignment) Pos() token.Pos { return a.Pattern.Pos() }
func (a *DestructuringAssignment) End() token.Pos { return a.Expr.End() }
func (*DestructuringAssignment) isDecl()          {}

// Definition is a node representing a definition of a value. A definition can
// also be annotated with a type annotation.
type Definition struct {
	Annotation *TypeAnnotation
	Name       *Ident
	Assign     token.Pos
	Args       []Pattern
	Body       Expr
}

func (*Definition) isDecl() {}
func (d *Definition) Pos() token.Pos {
	if d.Annotation != nil {
		return d.Annotation.Name.Pos()
	}

	return d.Name.Pos()
}
func (d *Definition) End() token.Pos { return d.Body.End() }

// TypeAnnotation is the annotation of a declaration with its type.
type TypeAnnotation struct {
	Name  *Ident
	Colon token.Pos
	Type  Type
}

func (ann *TypeAnnotation) Pos() token.Pos { return ann.Name.Pos() }
func (ann *TypeAnnotation) End() token.Pos { return ann.Type.End() }
