package ast

import (
	"github.com/elm-tangram/tangram/token"
)

// Type is a node representing a type.
type Type interface {
	Node
	isType()
}

// NamedType is a type that has a name and optional arguments.
type NamedType struct {
	// Name of the type.
	Name Expr
	// Args is the list of optional arguments of the type.
	Args []Type
}

func (NamedType) isType()          {}
func (t NamedType) Pos() token.Pos { return t.Name.Pos() }
func (t NamedType) End() token.Pos {
	if len(t.Args) > 0 {
		return t.Args[len(t.Args)-1].End()
	}

	return t.Name.End()
}

// VarType is a variable type, that is, a generic type.
type VarType struct {
	*Ident
}

func (VarType) isType() {}

// FuncType represents a function type. It has 1 or more arguments and a
// return type.
type FuncType struct {
	// Args is the list of the function argument types.
	Args []Type
	// Return type of the function.
	Return Type
}

func (FuncType) isType()          {}
func (t FuncType) Pos() token.Pos { return t.Args[0].Pos() }
func (t FuncType) End() token.Pos { return t.Return.End() }

// RecordType is a node representing a record type.
type RecordType struct {
	// Lbrace is the position of the opening brace.
	Lbrace token.Pos
	// Rbrace is the position of the closing brace.
	Rbrace token.Pos
	// Fields contains the list of fields and their types in the record.
	Fields []*RecordField
	// Extended is an optional variable type from which the type extends.
	Extended *VarType
	// Pipe is the optional position of the pipe token if Extended is not nil.
	Pipe token.Pos
}

func (RecordType) isType()           {}
func (t *RecordType) Pos() token.Pos { return t.Lbrace }
func (t *RecordType) End() token.Pos { return t.Rbrace }

// RecordField represents a field in a record type node.
type RecordField struct {
	// Name of the field.
	Name *Ident
	// Type of the field.
	Type Type
	// Colon position in the node.
	Colon token.Pos
}

func (t *RecordField) Pos() token.Pos { return t.Name.Pos() }
func (t *RecordField) End() token.Pos { return t.Type.End() }

// TupleType is a node representing a tuple with two ore more types.
type TupleType struct {
	// Lparen is the position of the opening parenthesis.
	Lparen token.Pos
	// Rparen is the position of the closing parenthesis.
	Rparen token.Pos
	// Elems is the list of types of the elements in the tuple.
	Elems []Type
}

func (TupleType) isType()          {}
func (t TupleType) Pos() token.Pos { return t.Lparen }
func (t TupleType) End() token.Pos { return t.Rparen }
