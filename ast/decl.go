package ast

import (
	"fmt"

	"github.com/elm-tangram/tangram/token"
)

// Decl is a declaration node.
type Decl interface {
	Node
	isDecl()
}

// ExposedList represents the list of exposed objects of an import or
// module declaration.
type ExposedList interface {
	Node
	isExposedList()
}

// ClosedList is a list with a fixed amount of objects exposed.
type ClosedList struct {
	Lparen  token.Pos
	Rparen  token.Pos
	Exposed []ExposedIdent
}

func (ClosedList) isExposedList()    {}
func (l *ClosedList) Pos() token.Pos { return l.Lparen }
func (l *ClosedList) End() token.Pos { return l.Rparen }

// OpenList means all objects in a module are exposed.
type OpenList struct {
	Lparen token.Pos
	Rparen token.Pos
}

func (OpenList) isExposedList()    {}
func (l *OpenList) Pos() token.Pos { return l.Lparen }
func (l *OpenList) End() token.Pos { return l.Rparen }

// ExposedIdent represents an identifier exposed by a module.
type ExposedIdent interface {
	Node
	isExposedIdent()
}

// ExposedVar is a variable, either a definition or the result of a
// destructuring that is exposed by a module.
type ExposedVar struct {
	*Ident
}

func (ExposedVar) isExposedIdent() {}

// ExposedUnion is a type union being exposed by a module along with some or all
// its constructors.
type ExposedUnion struct {
	Type  *Ident
	Ctors ExposedList
}

func (ExposedUnion) isExposedIdent()   {}
func (e *ExposedUnion) Pos() token.Pos { return e.Type.Pos() }
func (e *ExposedUnion) End() token.Pos { return e.Ctors.End() }

// ModuleDecl is a node representing a module declaration and contains the
// name of the module and the identifiers it exposes, if any.
type ModuleDecl struct {
	// Name of the module.
	Name Expr
	// Module is the position of the "module" keyword.
	Module token.Pos
	// Exposing is the list of exposed identifiers, if any.
	Exposing ExposedList
}

func (d *ModuleDecl) Pos() token.Pos { return d.Module }
func (d *ModuleDecl) End() token.Pos { return d.Exposing.End() }
func (d *ModuleDecl) isDecl()        {}

// ModuleName returns the name of the module.
func (d *ModuleDecl) ModuleName() string {
	if stringer, ok := d.Name.(fmt.Stringer); ok {
		return stringer.String()
	}
	return "_"
}

// ImportDecl is a node representing an import declaration. It contains the
// imported module as well as its alias, if any, and the exposed identifiers,
// if any.
type ImportDecl struct {
	// Module is the name of the imported module.
	Module Expr
	// Alias is the name of the alias for the module, if any.
	Alias *Ident
	// Import is the position of the "import" keyword.
	Import token.Pos
	// Exposing is the list of identifiers exposed, if any.
	Exposing ExposedList
}

func (d *ImportDecl) Pos() token.Pos { return d.Import }
func (d *ImportDecl) End() token.Pos {
	if d.Exposing == nil {
		if d.Alias == nil {
			return d.Module.End()
		}
		return d.Alias.End()
	}
	return d.Exposing.End()
}
func (d *ImportDecl) isDecl() {}

// ModuleName returns the name of the imported module.
func (d *ImportDecl) ModuleName() string {
	if stringer, ok := d.Module.(fmt.Stringer); ok {
		return stringer.String()
	}
	return "_"
}

// InfixDecl is a node representing the declaration of an operator's fixity.
// It contains the operator, the priority given and the associativity of the
// operator.
type InfixDecl struct {
	// InfixPos is the position of the "infix", "infixl" or "infixr" keyword.
	InfixPos token.Pos
	// Assoc is the associativity of the infix operator.
	Assoc Associativity
	// Op is the name of the operator.
	Op *Ident
	// Precence of the infix operator.
	Precedence *BasicLit
}

// Associativity is the type of associativity of the operator.
type Associativity byte

const (
	// NonAssoc is a non associative operator.
	NonAssoc Associativity = iota
	// Left associativity.
	Left
	// Right associativity.
	Right
)

func (InfixDecl) isDecl()          {}
func (d InfixDecl) Pos() token.Pos { return d.InfixPos }
func (d InfixDecl) End() token.Pos { return d.Op.End() }

// AliasDecl is a node representing a type alias declaration. It contains
// the name of the alias and its arguments along with the type it is aliasing.
type AliasDecl struct {
	// TypePos is the position of the "type" keyword.
	TypePos token.Pos
	// Alias is the position of the "alias" keyword.
	Alias token.Pos
	// Eq is the position of the "=" token.
	Eq token.Pos
	// Name is the name of the type.
	Name *Ident
	// Args are the optional arguments of the alias type.
	Args []*Ident
	// Type is the type definition of the alias type.
	Type Type
}

func (d AliasDecl) isDecl()        {}
func (d AliasDecl) Pos() token.Pos { return d.TypePos }
func (d AliasDecl) End() token.Pos { return d.Type.Pos() }

// UnionDecl is a node representing an union type declaration. Contains
// the name of the union type, the arguments and all the constructors for
// the type.
type UnionDecl struct {
	// TypePos is the position of the "type" keyword.
	TypePos token.Pos
	// Eq is the position of the "=" token.
	Eq token.Pos
	// Name is the name of the type.
	Name *Ident
	// Args are the optional arguments of the type.
	Args []*Ident
	// Ctors is the list of constructors for the union type.
	Ctors []*Constructor
}

func (d UnionDecl) isDecl()        {}
func (d UnionDecl) Pos() token.Pos { return d.TypePos }
func (d UnionDecl) End() token.Pos {
	if len(d.Ctors) == 0 {
		return token.NoPos
	}
	return d.Ctors[len(d.Ctors)-1].End()
}
func (d UnionDecl) LookupCtor(name string) *Constructor {
	for _, t := range d.Ctors {
		if t.Name.Name == name {
			return t
		}
	}
	return nil
}

// Constructor is a node representing the constructor of an union type.
// It contains the name of the constructor and its type arguments.
type Constructor struct {
	// Name of the constructor.
	Name *Ident
	// Arguments of the constructor.
	Args []Type
}

func (c Constructor) Pos() token.Pos { return c.Name.Pos() }
func (c Constructor) End() token.Pos {
	if len(c.Args) > 0 {
		return c.Args[len(c.Args)-1].End()
	}
	return c.Name.End()
}

// DestructuringAssignment represents a declaration using pattern matching on
// the expression.
type DestructuringAssignment struct {
	// Eq is the position of the "=" token.
	Eq token.Pos
	// Pattern used for the declaration.
	Pattern Pattern
	// Expr being destructured.
	Expr Expr
}

func (a *DestructuringAssignment) Pos() token.Pos { return a.Pattern.Pos() }
func (a *DestructuringAssignment) End() token.Pos { return a.Expr.End() }
func (*DestructuringAssignment) isDecl()          {}

// Definition is a node representing a definition of a value. A definition can
// also be annotated with a type annotation.
type Definition struct {
	// Annotation is the optional type annotation of the definition.
	Annotation *TypeAnnotation
	// Name is the name being defined.
	Name *Ident
	// Eq is the position of the "=" token.
	Eq token.Pos
	// Args are the optional arguments of the definition. A definition with
	// one ore more args is a function definition.
	Args []Pattern
	// Body of the definition.
	Body Expr
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
	// Name of the declaration being annotated.
	Name *Ident
	// Colon is the position of the ":" token.
	Colon token.Pos
	// Type of the declaration.
	Type Type
}

func (ann *TypeAnnotation) Pos() token.Pos { return ann.Name.Pos() }
func (ann *TypeAnnotation) End() token.Pos { return ann.Type.End() }
