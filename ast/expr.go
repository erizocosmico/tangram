package ast

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/elm-tangram/tangram/token"
)

// Expr is an expression node.
type Expr interface {
	Node
	isExpr()
}

// Ident represents an identifier, which is a name for something.
type Ident struct {
	// NamePos is the position of the name.
	NamePos token.Pos
	// Name of the identifier.
	Name string
	// Obj is the object this identifier refers to.
	Obj *Object
}

// NewIdent creates a new identifier with the given name and position.
func NewIdent(name string, pos token.Pos) *Ident {
	return &Ident{pos, name, nil}
}

func (i *Ident) Pos() token.Pos { return i.NamePos }
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
	// Expr to perform the selection on.
	Expr Expr
	// Selector identifier.
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
			buf.WriteString(e.Name)
			expr = nil
		case *SelectorExpr:
			buf.WriteString(e.Selector.String())
			buf.WriteRune('.')
			expr = e.Expr
		}
	}
	return buf.String()
}

// BasicLit represents a basic literal.
type BasicLit struct {
	// Position of the literal.
	Position token.Pos
	// Type of the literal.
	Type BasicLitType
	// Value of the literal.
	Value string
}

func (b *BasicLit) Pos() token.Pos { return b.Position }
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

// TupleLit is a tuple literal.
type TupleLit struct {
	// Lparen is the position of the opening parenthesis.
	Lparen token.Pos
	// Rparen is the position of the closing parenthesis.
	Rparen token.Pos
	// Elems are the elements of the tuple. len(l.Elems) will always be
	// greater than 2.
	Elems []Expr
}

func (l *TupleLit) Pos() token.Pos { return l.Lparen }
func (l *TupleLit) End() token.Pos { return l.Rparen }
func (*TupleLit) isExpr()          {}

// FuncApp is a function application, that is, a function and its arguments.
type FuncApp struct {
	// Func is the expression acting as function.
	Func Expr
	// Args of the function.
	Args []Expr
}

func (e *FuncApp) Pos() token.Pos { return e.Func.Pos() }
func (e *FuncApp) End() token.Pos { return e.Args[len(e.Args)-1].End() }
func (*FuncApp) isExpr()          {}

// RecordLit is a record literal.
type RecordLit struct {
	// Lbrace is the position of the opening brace.
	Lbrace token.Pos
	// Rbrace is the position of the closing brace.
	Rbrace token.Pos
	// Fields of the record.
	Fields []*FieldAssign
}

func (e *RecordLit) Pos() token.Pos { return e.Lbrace }
func (e *RecordLit) End() token.Pos { return e.Rbrace }
func (*RecordLit) isExpr()          {}

// FieldAssign is an assignation to a field of a record.
type FieldAssign struct {
	// Eq is the position of the "=" token in the assignation.
	Eq token.Pos
	// Field name.
	Field *Ident
	// Expr being assigned to the field.
	Expr Expr
}

func (n *FieldAssign) Pos() token.Pos { return n.Field.Pos() }
func (n *FieldAssign) End() token.Pos { return n.Expr.End() }

// RecordUpdate is the expression used to create a new record by updating some
// or all fields in another record.
type RecordUpdate struct {
	// Lbrace is the position of the opening brace.
	Lbrace token.Pos
	// Rbrace is the position of the closing brace.
	Rbrace token.Pos
	// Pipe is the position of the "|" token in the update.
	Pipe token.Pos
	// Record is the name of the variable holding the record being updated.
	Record *Ident
	// Fields are the assignation of the fields being modified.
	Fields []*FieldAssign
}

func (e *RecordUpdate) Pos() token.Pos { return e.Lbrace }
func (e *RecordUpdate) End() token.Pos { return e.Rbrace }
func (*RecordUpdate) isExpr()          {}

// LetExpr is an expression that allows declarations to be used inside
// its body.
type LetExpr struct {
	// Let is the position of the "let" keyword in the expression.
	Let token.Pos
	// Decls is the list of declarations.
	Decls []Decl
	// In is the position of the "in" keyword in the expression.
	In token.Pos
	// Body is the expression in which the previous declarations can be used.
	Body Expr
}

func (e *LetExpr) Pos() token.Pos { return e.Let }
func (e *LetExpr) End() token.Pos { return e.Body.End() }
func (*LetExpr) isExpr()          {}

// IfExpr is an if conditional expression.
type IfExpr struct {
	// If is the position of the "if" keyword in the expression.
	If token.Pos
	// Cond is the condition being evaluated.
	Cond Expr
	// Then is the position of the "then" keyword in the expression.
	Then token.Pos
	// ThenExpr is the expression that will be executed if and only if
	// Cond is true.
	ThenExpr Expr
	// Else is the position of the "else" keyword in the expression.
	Else token.Pos
	// ElseExpr is the expression that will be executed if an only if
	// Cond is false.
	ElseExpr Expr
}

func (e *IfExpr) Pos() token.Pos { return e.If }
func (e *IfExpr) End() token.Pos { return e.ElseExpr.End() }
func (*IfExpr) isExpr()          {}

// CaseExpr is an expression that allows conditional behavior based on the
// structure of algebraic data types and literals.
type CaseExpr struct {
	// Case is the position of the "case" keyword in the expression.
	Case token.Pos
	// Of is the position of the "of" keyword in the expression.
	Of token.Pos
	// Expr is the expression being matched.
	Expr Expr
	// Branches are all the possible branches to match Expr.
	Branches []*CaseBranch
}

func (e *CaseExpr) Pos() token.Pos { return e.Case }
func (e *CaseExpr) End() token.Pos { return e.Branches[len(e.Branches)-1].End() }
func (*CaseExpr) isExpr()          {}

// CaseBranch is a single branch of a case expression.
type CaseBranch struct {
	// Arrow is the position of the "->" token in the expression.
	Arrow token.Pos
	// Pattern is the pattern used to match the expression.
	Pattern Pattern
	// Expr is the expression that will be evaluated if and only if
	// the case expression matches the Pattern.
	Expr Expr
}

func (e *CaseBranch) Pos() token.Pos { return e.Pattern.Pos() }
func (e *CaseBranch) End() token.Pos { return e.Expr.End() }

// ListLit represents a list literal.
type ListLit struct {
	// Lbracket is the position of the opening bracket.
	Lbracket token.Pos
	// Rbracket is the position of the closing bracket.
	Rbracket token.Pos
	// Elems are the expressions being used as the list elements.
	Elems []Expr
}

func (e *ListLit) Pos() token.Pos { return e.Lbracket }
func (e *ListLit) End() token.Pos { return e.Rbracket }
func (*ListLit) isExpr()          {}

// UnaryOp is an expression representing an operator being applied to only one
// operand.
type UnaryOp struct {
	// Op is the operator name.
	Op *Ident
	// Expr is the operand.
	Expr Expr
}

func (e *UnaryOp) Pos() token.Pos { return e.Op.Pos() }
func (e *UnaryOp) End() token.Pos { return e.Expr.End() }
func (*UnaryOp) isExpr()          {}

// BinaryOp is an expression representing an operator being applied to two
// operands.
type BinaryOp struct {
	// Op is the operator name.
	Op *Ident
	// Lhs is the left hand side operand.
	Lhs Expr
	// Rhs is the right hand side operand.
	Rhs Expr
}

func (e *BinaryOp) Pos() token.Pos { return e.Lhs.Pos() }
func (e *BinaryOp) End() token.Pos { return e.Rhs.End() }
func (*BinaryOp) isExpr()          {}

// AccessorExpr is an expression for creating a function to access a specific
// field in a record.
type AccessorExpr struct {
	// Field name.
	Field *Ident
}

func (e *AccessorExpr) Pos() token.Pos { return e.Field.Pos() }
func (e *AccessorExpr) End() token.Pos { return e.Field.End() }
func (*AccessorExpr) isExpr()          {}

// TupleCtor is a special function used to create tuples.
type TupleCtor struct {
	// Lparen is the position of the opening parenthesis.
	Lparen token.Pos
	// Rparen is the position of the closing parenthesis.
	Rparen token.Pos
	// Elems is the number of elements in the tuple to create.
	Elems int
}

func (e *TupleCtor) Pos() token.Pos { return e.Lparen }
func (e *TupleCtor) End() token.Pos { return e.Rparen }
func (e *TupleCtor) isExpr()        {}

// Lambda is a lambda function expression.
type Lambda struct {
	// Backslash is the position of the "\" token in the expression.
	Backslash token.Pos
	// Arrow is the position of the "->" token in the expression.
	Arrow token.Pos
	// Args are the arguments of the function.
	Args []Pattern
	// Expr is the body of the function.
	Expr Expr
}

func (e *Lambda) Pos() token.Pos { return e.Backslash }
func (e *Lambda) End() token.Pos { return e.Expr.End() }
func (*Lambda) isExpr()          {}

// ParensExpr is an expression representing another expression wrapped with
// parenthesis.
type ParensExpr struct {
	// Lparen is the position of the opening parenthesis.
	Lparen token.Pos
	// Rparen is the position of the closing parenthesis.
	Rparen token.Pos
	// Expr is the expr being wrapped with parenthesis.
	Expr Expr
}

func (e *ParensExpr) Pos() token.Pos { return e.Lparen }
func (e *ParensExpr) End() token.Pos { return e.Rparen }
func (*ParensExpr) isExpr()          {}

// BadExpr is a malformed expression.
type BadExpr struct {
	StartPos token.Pos
	EndPos   token.Pos
}

func (e *BadExpr) Pos() token.Pos { return e.StartPos }
func (e *BadExpr) End() token.Pos { return e.EndPos }
func (*BadExpr) isExpr()          {}
