package ast

import "github.com/erizocosmico/elmo/token"

// Type is a node representing a type.
type Type interface {
	Node
	isType()
}

// BasicType is a type that has a name and 0 or more arguments.
type BasicType struct {
	Name Expr
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
}

func (t RecordTypeField) Pos() token.Pos { return t.Name.Pos() }
func (t RecordTypeField) End() token.Pos { return t.Type.End() }

// TupleType is a node representing a tuple with two ore more types.
type TupleType struct {
	Lparen token.Pos
	Rparen token.Pos
	Elems  []Type
}

func (TupleType) isType()          {}
func (t TupleType) Pos() token.Pos { return t.Lparen }
func (t TupleType) End() token.Pos { return t.Rparen }
