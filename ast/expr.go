package ast

import (
	"bytes"
	"fmt"
	"unicode"
	"unicode/utf8"

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

// NewIdent creates a new identifier with the given name and position.
func NewIdent(name string, pos *token.Position) *Ident {
	return &Ident{pos, name, nil}
}

func (i *Ident) Pos() token.Pos { return i.NamePos.Offset }
func (i *Ident) End() token.Pos { return i.Pos() + token.Pos(len(i.Name)) }
func (*Ident) isExpr()          {}
func (i *Ident) String() string { return i.Name }

// IsOp reports whether the identifiers corresponds to an operator or not.
func (i *Ident) IsOp() bool {
	r, _ := utf8.DecodeRuneInString(i.Name)
	return !unicode.IsLetter(r) && !unicode.IsDigit(r)
}

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
		return "Int"
	case Float:
		return "Float"
	case String:
		return "String"
	case Bool:
		return "Bool"
	case Char:
		return "Char"
	default:
		return "error"
	}
}

type TupleLit struct {
	Lparen token.Pos
	Rparen token.Pos
	Elems  []Expr
}

func (l *TupleLit) Pos() token.Pos { return l.Lparen }
func (l *TupleLit) End() token.Pos { return l.Rparen }
func (*TupleLit) isExpr()          {}

type FuncApp struct {
	Func Expr
	Args []Expr
}

func (e *FuncApp) Pos() token.Pos { return e.Func.Pos() }
func (e *FuncApp) End() token.Pos { return e.Args[len(e.Args)-1].End() }
func (*FuncApp) isExpr()          {}

type RecordLit struct {
	Lbrace token.Pos
	Rbrace token.Pos
	Fields []*FieldAssign
}

func (e *RecordLit) Pos() token.Pos { return e.Lbrace }
func (e *RecordLit) End() token.Pos { return e.Rbrace }
func (*RecordLit) isExpr()          {}

type FieldAssign struct {
	Eq    token.Pos
	Field *Ident
	Expr  Expr
}

func (n *FieldAssign) Pos() token.Pos { return n.Field.Pos() }
func (n *FieldAssign) End() token.Pos { return n.Expr.End() }

type RecordUpdate struct {
	Lbrace token.Pos
	Rbrace token.Pos
	Pipe   token.Pos
	Record *Ident
	Fields []*FieldAssign
}

func (e *RecordUpdate) Pos() token.Pos { return e.Lbrace }
func (e *RecordUpdate) End() token.Pos { return e.Rbrace }
func (*RecordUpdate) isExpr()          {}

type LetExpr struct {
	Let   token.Pos
	Decls []Decl
	In    token.Pos
	Body  Expr
}

func (e *LetExpr) Pos() token.Pos { return e.Let }
func (e *LetExpr) End() token.Pos { return e.Body.End() }
func (*LetExpr) isExpr()          {}

type IfExpr struct {
	If       token.Pos
	Cond     Expr
	Then     token.Pos
	ThenExpr Expr
	Else     token.Pos
	ElseExpr Expr
}

func (e *IfExpr) Pos() token.Pos { return e.If }
func (e *IfExpr) End() token.Pos { return e.ElseExpr.End() }
func (*IfExpr) isExpr()          {}

type CaseExpr struct {
	Case     token.Pos
	Of       token.Pos
	Expr     Expr
	Branches []*CaseBranch
}

func (e *CaseExpr) Pos() token.Pos { return e.Case }
func (e *CaseExpr) End() token.Pos { return e.Branches[len(e.Branches)-1].End() }
func (*CaseExpr) isExpr()          {}

type CaseBranch struct {
	Arrow   token.Pos
	Pattern Pattern
	Expr    Expr
}

func (e *CaseBranch) Pos() token.Pos { return e.Pattern.Pos() }
func (e *CaseBranch) End() token.Pos { return e.Expr.End() }

type ListLit struct {
	Lbracket token.Pos
	Rbracket token.Pos
	Elems    []Expr
}

func (e *ListLit) Pos() token.Pos { return e.Lbracket }
func (e *ListLit) End() token.Pos { return e.Rbracket }
func (*ListLit) isExpr()          {}

type UnaryExpr struct {
	Op   *Ident
	Expr Expr
}

func (e *UnaryExpr) Pos() token.Pos { return e.Op.Pos() }
func (e *UnaryExpr) End() token.Pos { return e.Expr.End() }
func (*UnaryExpr) isExpr()          {}

type BinaryExpr struct {
	Op  *Ident
	Lhs Expr
	Rhs Expr
}

func (e *BinaryExpr) Pos() token.Pos { return e.Lhs.Pos() }
func (e *BinaryExpr) End() token.Pos { return e.Rhs.End() }
func (*BinaryExpr) isExpr()          {}

type AccessorExpr struct {
	Field *Ident
}

func (e *AccessorExpr) Pos() token.Pos { return e.Field.Pos() }
func (e *AccessorExpr) End() token.Pos { return e.Field.End() }
func (*AccessorExpr) isExpr()          {}

type TupleCtor struct {
	Lparen token.Pos
	Rparen token.Pos
	Elems  int
}

func (e *TupleCtor) Pos() token.Pos { return e.Lparen }
func (e *TupleCtor) End() token.Pos { return e.Rparen }
func (e *TupleCtor) isExpr()        {}

type Lambda struct {
	Backslash token.Pos
	Arrow     token.Pos
	Args      []Pattern
	Expr      Expr
}

func (e *Lambda) Pos() token.Pos { return e.Backslash }
func (e *Lambda) End() token.Pos { return e.Expr.End() }
func (*Lambda) isExpr()          {}

type ParensExpr struct {
	Lparen token.Pos
	Rparen token.Pos
	Expr   Expr
}

func (e *ParensExpr) Pos() token.Pos { return e.Lparen }
func (e *ParensExpr) End() token.Pos { return e.Rparen }
func (*ParensExpr) isExpr()          {}
