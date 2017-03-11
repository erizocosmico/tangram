package ast

import "github.com/erizocosmico/elmo/token"

// Pattern represents a pattern node.
type Pattern interface {
	Node
	isPattern()
}

// ArgPattern is a special type of Pattern node that represents pattern nodes
// that can be used as a pattern in a function argument or a destructuring
// assignment. These are basically the following types:
// - Anything pattern
// - Tuple pattern
// - Record pattern
// - Var pattern
// The rest of the patterns can not cover all possible branches and can only be
// used in `case` expressions.
type ArgPattern interface {
	Pattern
	isArgPattern()
}

// VarPattern is a pattern that will bind something to a variable.
// For example, pattern `head :: tail` will bind the head of the list to the
// `head` variable.
type VarPattern struct {
	// Name of the variable.
	Name *Ident
}

func (p VarPattern) Pos() token.Pos { return p.Name.Pos() }
func (p VarPattern) End() token.Pos { return p.Name.End() }
func (VarPattern) isPattern()       {}
func (VarPattern) isArgPattern()    {}

// AnythingPattern is the "_" identifier, which matches everything.
type AnythingPattern struct {
	// Underscore is the position of the "_" identifier.
	Underscore token.Pos
}

func (p AnythingPattern) Pos() token.Pos { return p.Underscore }
func (p AnythingPattern) End() token.Pos { return p.Pos() + token.Pos(1) }
func (AnythingPattern) isPattern()       {}
func (AnythingPattern) isArgPattern()    {}

// LiteralPattern represents a pattern that will match only if a certain
// literal is present.
type LiteralPattern struct {
	Literal *BasicLit
}

func (p LiteralPattern) Pos() token.Pos { return p.Literal.Pos() }
func (p LiteralPattern) End() token.Pos { return p.Literal.End() }
func (LiteralPattern) isPattern()       {}

// AliasPattern is a pattern with an alias.
type AliasPattern struct {
	// Name of the alias.
	Name *Ident
	// Pattern being aliased.
	Pattern Pattern
}

func (p AliasPattern) Pos() token.Pos { return p.Pattern.Pos() }
func (p AliasPattern) End() token.Pos { return p.Name.End() }
func (AliasPattern) isPattern()       {}
func (AliasPattern) isArgPattern()    {}

// CtorPattern is a pattern that will match a given constructor. If the
// constructor has arguments, it can have more patterns for them.
type CtorPattern struct {
	// Ctor is the name of the constructor.
	Ctor *Ident
	// Patterns for the constructor arguments.
	Patterns []Pattern
}

func (p CtorPattern) Pos() token.Pos { return p.Ctor.Pos() }
func (p CtorPattern) End() token.Pos {
	if len(p.Patterns) == 0 {
		return p.Ctor.End()
	}
	return p.Patterns[len(p.Patterns)-1].End()
}
func (CtorPattern) isPattern() {}

// TuplePattern is a pattern that will match a tuple and its elements.
type TuplePattern struct {
	// Lparen is the position of the open parenthesis.
	Lparen token.Pos
	// Rparen is the position of the closing parenthesis.
	Rparen token.Pos
	// Patterns for the elements of the tuple.
	Patterns []Pattern
}

func (p TuplePattern) Pos() token.Pos { return p.Lparen }
func (p TuplePattern) End() token.Pos { return p.Rparen }
func (TuplePattern) isPattern()       {}
func (TuplePattern) isArgPattern()    {}

// RecordPattern is a pattern that will match a record and its fields.
type RecordPattern struct {
	// Lbrace is the position of the opening brace.
	Lbrace token.Pos
	// Rbrace is the position of the closing brace.
	Rbrace token.Pos
	// Patterns for the fields in the record.
	Patterns []Pattern
}

func (p RecordPattern) Pos() token.Pos { return p.Lbrace }
func (p RecordPattern) End() token.Pos { return p.Rbrace }
func (RecordPattern) isPattern()       {}
func (RecordPattern) isArgPattern()    {}

// ListPattern is a pattern that will match a list and its elements.
type ListPattern struct {
	// Lbracket is the position of the opening bracket.
	Lbracket token.Pos
	// Rbracket is the position of the closing bracket.
	Rbracket token.Pos
	// Pattern for the elements in the list.
	Patterns []Pattern
}

func (p ListPattern) Pos() token.Pos { return p.Lbracket }
func (p ListPattern) End() token.Pos { return p.Rbracket }
func (ListPattern) isPattern()       {}
