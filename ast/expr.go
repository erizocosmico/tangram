package ast

import (
	"bytes"
	"fmt"

	"github.com/erizocosmico/elmo/token"
)

// Expr is an expression node.
type Expr interface {
	Node
	isExpr()
}

// Ident represents an identifier, which is a name for something.
type Ident struct {
	NamePos *token.Position
	Name    string
	Obj     *Object
}

func NewIdent(name string, pos *token.Position) *Ident {
	return &Ident{pos, name, nil}
}

func (i *Ident) Pos() token.Pos { return i.NamePos.Offset }
func (i *Ident) End() token.Pos { return i.Pos() + token.Pos(len(i.Name)) }
func (*Ident) isExpr()          {}
func (i *Ident) String() string { return i.Name }

// SelectorExpr represents an expression preceded by a selector.
type SelectorExpr struct {
	Expr     Expr
	Selector *Ident
}

// NewSelectorExpr creates a new Selector expression from a list of
// identifiers.
func NewSelectorExpr(idents ...*Ident) *SelectorExpr {
	if len(idents) < 2 {
		return nil
	}

	if len(idents) == 2 {
		return &SelectorExpr{
			Expr:     idents[1],
			Selector: idents[0],
		}
	}

	return &SelectorExpr{
		Expr:     NewSelectorExpr(idents[1:]...),
		Selector: idents[0],
	}
}

func (e *SelectorExpr) Pos() token.Pos { return e.Selector.Pos() }
func (e *SelectorExpr) End() token.Pos { return e.Expr.End() }
func (*SelectorExpr) isExpr()          {}
func (e *SelectorExpr) String() string {
	var buf bytes.Buffer
	buf.WriteString(e.Selector.Name)
	buf.WriteRune('.')
	expr := e.Expr
	for expr != nil {
		switch e := expr.(type) {
		case *Ident:
			buf.WriteString(e.String())
			expr = nil
		case *SelectorExpr:
			buf.WriteString(e.Selector.String())
			buf.WriteRune('.')
			expr = e.Expr
		default:
			// unreachable
			panic(fmt.Errorf("invalid expression of type %T in selector", expr))
		}
	}
	return buf.String()
}

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
	Position *token.Position
	Type     BasicLitType
	Value    string
}

func (b *BasicLit) Pos() token.Pos { return b.Position.Offset }
func (b *BasicLit) End() token.Pos { return b.Pos() + token.Pos(len(b.Value)) }
func (*BasicLit) isExpr()          {}

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

func (t BasicLitType) String() string {
	switch t {
	case Int:
		return "int"
	case Float:
		return "float"
	case String:
		return "string"
	case Bool:
		return "bool"
	case Char:
		return "char"
	default:
		return "error"
	}
}
