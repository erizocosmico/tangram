package ast

import "github.com/erizocosmico/elmo/token"

type Pattern interface {
	Node
	isPattern()
}

type ArgPattern interface {
	Pattern
	isArgPattern()
}

type VarPattern struct {
	Name *Ident
}

func (p VarPattern) Pos() token.Pos { return p.Name.Pos() }
func (p VarPattern) End() token.Pos { return p.Name.End() }
func (VarPattern) isPattern()       {}
func (VarPattern) isArgPattern()    {}

type AnythingPattern struct {
	Underscore token.Pos
}

func (p AnythingPattern) Pos() token.Pos { return p.Underscore }
func (p AnythingPattern) End() token.Pos { return p.Pos() + token.Pos(1) }
func (AnythingPattern) isPattern()       {}
func (AnythingPattern) isArgPattern()    {}

type LiteralPattern struct {
	Literal *BasicLit
}

func (p LiteralPattern) Pos() token.Pos { return p.Literal.Pos() }
func (p LiteralPattern) End() token.Pos { return p.Literal.End() }
func (LiteralPattern) isPattern()       {}

type AliasPattern struct {
	Name    *Ident
	Pattern Pattern
}

func (p AliasPattern) Pos() token.Pos { return p.Pattern.Pos() }
func (p AliasPattern) End() token.Pos { return p.Name.End() }
func (AliasPattern) isPattern()       {}
func (AliasPattern) isArgPattern()    {}

type CtorPattern struct {
	Ctor     *Ident
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

type TuplePattern struct {
	Lparen   token.Pos
	Rparen   token.Pos
	Patterns []Pattern
}

func (p TuplePattern) Pos() token.Pos { return p.Lparen }
func (p TuplePattern) End() token.Pos { return p.Rparen }
func (TuplePattern) isPattern()       {}
func (TuplePattern) isArgPattern()    {}

type RecordPattern struct {
	Lbrace   token.Pos
	Rbrace   token.Pos
	Patterns []Pattern
}

func (p RecordPattern) Pos() token.Pos { return p.Lbrace }
func (p RecordPattern) End() token.Pos { return p.Rbrace }
func (RecordPattern) isPattern()       {}
func (RecordPattern) isArgPattern()    {}

type ListPattern struct {
	Lbracket token.Pos
	Rbracket token.Pos
	Patterns []Pattern
}

func (p ListPattern) Pos() token.Pos { return p.Lbracket }
func (p ListPattern) End() token.Pos { return p.Rbracket }
func (ListPattern) isPattern()       {}
