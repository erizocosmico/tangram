package parser

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/diagnostic"
	"github.com/erizocosmico/elmo/operator"
	"github.com/erizocosmico/elmo/scanner"
	"github.com/erizocosmico/elmo/token"
)

type parser struct {
	sess       *Session
	scanner    *scanner.Scanner
	fileName   string
	unresolved []*ast.Ident
	mode       ParseMode

	// inNewDecl indicates the current token is at the start of the line,
	// that is, we're parsing a new declaration.
	// New declarations must be acknowledge before asking for a new token
	// otherwise, an error will occur.
	inNewDecl bool
	// isStart indicates the current token is the first one of the parsing.
	// What that means is that this token must be at the start of the line, as
	// it begins a new declaration.
	isStart bool
	// ignoreNewDecl indicates if the parser must ignore tokens that are new
	// declarations, and just treat them as if they weren't.
	ignoreNewDecl bool
	lineStart     []int
	tok           *token.Token
	errors        []error
	regions       []*token.Position
}

func newParser(sess *Session) *parser {
	return &parser{sess: sess}
}

type bailout struct{}

func (p *parser) init(fileName string, s *scanner.Scanner, mode ParseMode) {
	p.inNewDecl = false
	p.isStart = true
	p.resetLineStart()
	p.ignoreNewDecl = false
	p.scanner = s
	p.fileName = fileName
	p.mode = mode

	p.next()
}

func (p *parser) parseFile() *ast.File {
	mod := p.parseModule()
	imports := p.parseImports()
	if p.mode == ImportsAndFixity {
		p.skipUntilNextFixity()
	}

	var decls []ast.Decl
	if p.mode == FullParse || p.mode == ImportsAndFixity {
		for p.tok.Type != token.EOF {
			decls = append(decls, p.parseDecl())
		}
	}

	return &ast.File{
		Name:    p.fileName,
		Module:  mod,
		Imports: imports,
		Decls:   decls,
	}
}

func (p *parser) skipUntilNextFixity() {
	p.ignoreNewDecl = true
	for {
		switch p.tok.Type {
		case token.Infix, token.Infixr, token.Infixl, token.EOF:
			p.ignoreNewDecl = false
			return
		}
		p.next()
	}
}

func (p *parser) parseModule() *ast.ModuleDecl {
	var decl = new(ast.ModuleDecl)
	p.startRegion()
	decl.Module = p.expect(token.Module)
	decl.Name = p.parseModuleName()

	if p.is(token.Exposing) {
		exposedList := new(ast.ExposingList)
		p.expect(token.Exposing)
		exposedList.Lparen = p.expect(token.LeftParen)
		exposedList.Idents = p.parseExposedIdents()
		if len(exposedList.Idents) == 0 {
			p.errorExpectedOneOf(p.tok, token.Range, token.Identifier)
		}
		exposedList.Rparen = p.expect(token.RightParen)
		decl.Exposing = exposedList
	}

	p.finishedDecl()
	p.endRegion()
	return decl
}

func (p *parser) parseImports() []*ast.ImportDecl {
	var imports []*ast.ImportDecl
	for p.tok.Type == token.Import {
		imports = append(imports, p.parseImport())
	}
	return imports
}

func (p *parser) parseImport() *ast.ImportDecl {
	var decl = new(ast.ImportDecl)
	p.startRegion()
	decl.Import = p.expect(token.Import)
	decl.Module = p.parseModuleName()

	if p.is(token.As) {
		p.expect(token.As)
		decl.Alias = p.parseUpperName()
	}

	if p.is(token.Exposing) {
		exposedList := new(ast.ExposingList)
		p.expect(token.Exposing)
		exposedList.Lparen = p.expect(token.LeftParen)
		exposedList.Idents = p.parseExposedIdents()
		if len(exposedList.Idents) == 0 {
			p.errorExpectedOneOf(p.tok, token.Range, token.Identifier)
		}
		exposedList.Rparen = p.expect(token.RightParen)
		decl.Exposing = exposedList
	}

	p.finishedDecl()
	p.endRegion()
	return decl
}

func (p *parser) parseModuleName() ast.Expr {
	path := []*ast.Ident{p.parseUpperName()}

	for p.is(token.Dot) {
		// TODO(erizocosmico): check dot is right after the ident
		p.expect(token.Dot)
		path = append(path, p.parseUpperName())
	}

	if len(path) == 1 {
		return path[0]
	}

	return ast.NewSelectorExpr(path...)
}

func (p *parser) parseExposedIdents() []*ast.ExposedIdent {
	if p.is(token.Range) {
		p.expect(token.Range)
		return []*ast.ExposedIdent{
			ast.NewExposedIdent(
				ast.NewIdent(token.Range.String(), p.tok.Position),
			),
		}
	}

	if !p.is(token.LeftParen) && !p.is(token.Identifier) {
		return nil
	}

	exposed := []*ast.ExposedIdent{p.parseExposedIdent()}
	for p.is(token.Comma) {
		p.expect(token.Comma)
		exposed = append(exposed, p.parseExposedIdent())
	}

	return exposed
}

func (p *parser) parseExposedIdent() *ast.ExposedIdent {
	ident := ast.NewExposedIdent(p.parseIdentifierOrOp())

	if p.is(token.LeftParen) {
		if !isUpper(ident.Name) {
			p.errorMessage(ident.NamePos, "I was expecting an upper case name.")
		}
		var exposingList = new(ast.ExposingList)
		exposingList.Lparen = p.expect(token.LeftParen)
		exposingList.Idents = p.parseConstructorExposedIdents()
		if len(exposingList.Idents) == 0 {
			p.errorExpectedOneOf(p.tok, token.Range, token.Identifier)
		}
		exposingList.Rparen = p.expect(token.RightParen)
		ident.Exposing = exposingList
	}

	return ident
}

func (p *parser) parseConstructorExposedIdents() (idents []*ast.ExposedIdent) {
	if p.is(token.Range) {
		p.expect(token.Range)
		idents = append(
			idents,
			ast.NewExposedIdent(
				ast.NewIdent(token.Range.String(), p.tok.Position),
			),
		)
		return
	}

	for {
		idents = append(idents, ast.NewExposedIdent(p.parseUpperName()))
		if p.is(token.RightParen) {
			return
		}

		p.expect(token.Comma)
	}
}

func (p *parser) parseIdentifierOrOp() *ast.Ident {
	if !p.is(token.LeftParen) {
		return p.parseIdentifier()
	}

	p.expect(token.LeftParen)
	op := p.parseOp()
	p.expect(token.RightParen)
	return op
}

// parseQualifiedIdentifier parses an identifier and its qualifier, if any.
// This should only be used outside of expressions, as expressions already
// parse selector expressions by themselves.
// This is specially to parse identifiers with an optional qualifier.
func (p *parser) parseQualifiedIdentifier() ast.Expr {
	if isUpper(p.tok.Value) {
		var path []*ast.Ident
		path = append(path, p.parseUpperName())

		for p.is(token.Dot) {
			// TODO(erizocosmico): check the dot comes immediately after the name
			p.expect(token.Dot)

			if p.is(token.Identifier) {
				if isLower(p.tok.Value) {
					path = append(path, p.parseLowerName())
					// a lower identifier means the end of the path
					break
				}
				path = append(path, p.parseUpperName())
			} else {
				p.expect(token.Identifier)
				return nil
			}
		}

		if len(path) == 1 {
			return path[0]
		}

		return ast.NewSelectorExpr(path...)
	}
	return p.parseLowerName()
}

func (p *parser) parseIdentifier() *ast.Ident {
	name := "_"
	pos := p.tok.Position
	if p.is(token.Identifier) {
		name = p.tok.Value
		p.next()
	} else {
		p.expect(token.Identifier)
	}

	return ast.NewIdent(name, pos)
}

func (p *parser) parseUpperName() *ast.Ident {
	ident := p.parseIdentifier()
	if !isUpper(ident.Name) {
		p.errorMessage(ident.NamePos, "I was expecting an upper case name.")
	}
	return ident
}

func (p *parser) parseLowerName() *ast.Ident {
	ident := p.parseIdentifier()
	if !isLower(ident.Name) {
		p.errorMessage(ident.NamePos, "I was expecting a lower case name.")
	}
	return ident
}

func (p *parser) parseOp() *ast.Ident {
	name := "_"
	pos := p.tok.Position
	if p.tok.Type == token.Op {
		name = p.tok.Value
		p.next()
	} else {
		p.expect(token.Op)
	}

	return &ast.Ident{NamePos: pos, Name: name}
}

func (p *parser) parseDecl() ast.Decl {
	p.startRegion()
	var decl ast.Decl
	switch p.tok.Type {
	case token.TypeDef:
		decl = p.parseTypeDecl()

	case token.Infixl, token.Infixr, token.Infix:
		decl = p.parseInfixDecl()

	case token.Identifier:
		if p.tok.Value == "_" {
			decl = p.parseDestructuringAssignment()
		} else {
			decl = p.parseDefinition()
		}

	case token.LeftParen:
		if p.peek().Type == token.Op {
			decl = p.parseDefinition()
		} else {
			decl = p.parseDestructuringAssignment()
		}

	case token.LeftBrace:
		decl = p.parseDestructuringAssignment()

	default:
		p.errorExpectedOneOf(p.tok, token.TypeDef, token.Identifier)
		panic(bailout{})
	}

	p.finishedDecl()
	p.endRegion()

	if p.mode == ImportsAndFixity {
		p.skipUntilNextFixity()
	}

	return decl
}

const errorMsgInvalidDestructuringPattern = `This is not a valid pattern for a destructuring assignment.
I am looking for one of the following things:

- a lower case name
- an underscore ("_")
- a tuple pattern (e.g. "(first, second)")
- a record pattern (e.g. "{x, y}")`

func (p *parser) parseDestructuringAssignment() *ast.DestructuringAssignment {
	a := new(ast.DestructuringAssignment)
	a.Pattern = p.parsePattern(true)
	_, ok := a.Pattern.(ast.ArgPattern)
	if !ok {
		p.errorMessage(
			p.tok.Position,
			errorMsgInvalidDestructuringPattern,
		)
		panic(bailout{})
	}

	a.Eq = p.expect(token.Assign)
	a.Expr = p.parseExpr()

	return a
}

func (p *parser) parseInfixDecl() ast.Decl {
	var assoc ast.Associativity
	if p.is(token.Infixl) {
		assoc = ast.LeftAssoc
	} else if p.is(token.Infixr) {
		assoc = ast.RightAssoc
	}

	pos := p.expectOneOf(token.Infixl, token.Infixr, token.Infix)
	if !p.is(token.Int) {
		p.errorExpected(p.tok, token.Int)
	}

	precedence := p.parseLiteral()
	n, _ := strconv.Atoi(precedence.Value)
	if n < 0 || n > 9 {
		p.errorMessage(precedence.Position, "Operator precedence must be a number between 0 and 9, both included.")
	}

	op := p.parseOp()
	return &ast.InfixDecl{
		InfixPos:   pos,
		Assoc:      assoc,
		Precedence: precedence,
		Op:         op,
	}
}

func (p *parser) parseTypeDecl() ast.Decl {
	typePos := p.expect(token.TypeDef)
	if p.is(token.Alias) {
		return p.parseAliasType(typePos)
	}

	return p.parseUnionType(typePos)
}

func (p *parser) parseAliasType(typePos token.Pos) ast.Decl {
	decl := &ast.AliasDecl{
		TypePos: typePos,
		Alias:   p.expect(token.Alias),
	}
	decl.Name = p.parseUpperName()
	decl.Args = p.parseTypeDeclArgs()
	decl.Eq = p.expect(token.Assign)
	decl.Type = p.expectType()
	return decl
}

func (p *parser) parseUnionType(typePos token.Pos) ast.Decl {
	decl := &ast.UnionDecl{TypePos: typePos}
	decl.Name = p.parseUpperName()
	decl.Args = p.parseTypeDeclArgs()
	decl.Eq = p.expect(token.Assign)
	decl.Types = p.parseConstructors()
	return decl
}

func (p *parser) parseTypeDeclArgs() (idents []*ast.Ident) {
	for p.is(token.Identifier) {
		idents = append(idents, p.parseLowerName())
	}
	return
}

func (p *parser) parseConstructors() (cs []*ast.Constructor) {
	cs = append(cs, p.parseConstructor())
	for p.is(token.Pipe) {
		p.expect(token.Pipe)
		ctor := p.parseConstructor()
		cs = append(cs, ctor)
	}
	return
}

func (p *parser) parseConstructor() *ast.Constructor {
	c := new(ast.Constructor)
	c.Name = p.parseUpperName()
	c.Args = p.parseTypeList()
	return c
}

func (p *parser) parseTypeList() (types []ast.Type) {
	for (p.is(token.LeftParen) || p.is(token.Identifier) || p.is(token.LeftBrace)) && !p.atLineStart() {
		var typ ast.Type
		switch p.tok.Type {
		case token.LeftParen, token.LeftBrace:
			typ = p.parseType()
		case token.Identifier:
			typ = &ast.BasicType{Name: p.parseQualifiedIdentifier()}
		}

		if typ == nil {
			break
		}
		types = append(types, typ)
	}
	return
}

// parseType parses a complete type, whether it is a function type or an atom
// type.
func (p *parser) parseType() ast.Type {
	t := p.parseAtomType()
	if t == nil {
		return nil
	}

	if !p.is(token.Arrow) {
		return t
	}

	types := []ast.Type{t}
	for p.is(token.Arrow) {
		p.expect(token.Arrow)
		typ := p.parseAtomType()
		if typ == nil {
			break
		}

		types = append(types, typ)
	}

	size := len(types)
	return &ast.FuncType{
		Args:   types[:size-1],
		Return: types[size-1],
	}
}

// parseAtomType parses a type that can make sense on their own, that is,
// a tuple, a record or a basic type.
func (p *parser) parseAtomType() ast.Type {
	if p.atLineStart() {
		return nil
	}

	switch p.tok.Type {
	case token.LeftParen:
		lparenPos := p.expect(token.LeftParen)
		typ := p.expectType()

		// is a tuple
		if p.is(token.Comma) {
			t := &ast.TupleType{
				Lparen: lparenPos,
				Elems:  []ast.Type{typ},
			}

			for !p.is(token.RightParen) {
				p.expect(token.Comma)
				t.Elems = append(t.Elems, p.expectType())
			}

			t.Rparen = p.expect(token.RightParen)
			return t
		}

		p.expect(token.RightParen)
		return typ
	case token.Identifier:
		name := p.parseQualifiedIdentifier()
		if name, ok := name.(*ast.Ident); ok && isLower(name.Name) {
			return &ast.BasicType{Name: name}
		}

		return &ast.BasicType{
			Name: name,
			Args: p.parseTypeList(),
		}
	case token.LeftBrace:
		return p.parseRecordType()
	default:
		p.errorExpectedOneOf(p.tok, token.LeftParen, token.LeftBrace, token.Identifier)
		// TODO: think of a better way to recover from this error
		panic(bailout{})
	}
}

func (p *parser) parseRecordType() *ast.RecordType {
	t := &ast.RecordType{
		Lbrace: p.expect(token.LeftBrace),
	}

	for !p.is(token.RightBrace) && !p.is(token.EOF) {
		if len(t.Fields) > 0 {
			p.expect(token.Comma)
		}

		f := new(ast.RecordField)
		f.Name = p.parseLowerName()
		f.Colon = p.expect(token.Colon)
		f.Type = p.expectType()
		t.Fields = append(t.Fields, f)
	}

	t.Rbrace = p.expect(token.RightBrace)
	return t
}

func (p *parser) parseDefinition() ast.Decl {
	decl := new(ast.Definition)

	var name *ast.Ident
	if p.is(token.Identifier) {
		name = p.parseLowerName()
	} else {
		p.expect(token.LeftParen)
		name = p.parseOp()
		p.expect(token.RightParen)
	}

	if p.is(token.Colon) {
		decl.Annotation = &ast.TypeAnnotation{Name: name}
		decl.Annotation.Colon = p.expect(token.Colon)
		decl.Annotation.Type = p.expectType()
		p.finishedDecl()

		defName := p.parseIdentifierOrOp()
		if defName.Name != name.Name {
			p.errorMessage(
				p.tok.Position,
				fmt.Sprintf(
					"A definition must be right below its type annotation, I found the definition of `%s` after the annotation of `%s` instead.",
					defName.Name,
					name.Name,
				),
			)
		}

		decl.Name = defName
	} else {
		decl.Name = name
	}

	decl.Args = p.parseFuncArgs(token.Assign)
	decl.Eq = p.expect(token.Assign)
	decl.Body = p.parseExpr()
	return decl
}

const errorMsgInvalidArgPattern = `This is not a valid pattern for a function argument.
I am looking for one of the following things:

- a lower case name
- an underscore ("_")
- a tuple pattern (e.g. "(first, second)")
- a record pattern (e.g. "{x, y}")`

func (p *parser) parseFuncArgs(end token.Type) []ast.Pattern {
	var args []ast.Pattern
	for !p.is(end) && !p.is(token.EOF) {
		tok := p.tok
		// in arguments, we parse the patterns as non-greedy so it forces the
		// developer to wrap around parenthesis the alias pattern
		pattern := p.parsePattern(false)
		arg, ok := pattern.(ast.ArgPattern)
		if !ok {
			p.errorMessage(
				tok.Position,
				errorMsgInvalidArgPattern,
			)
		}

		args = append(args, arg)
	}

	return args
}

// parsePattern parses the next pattern. If `greedy` is true, it will try to
// find an alias at the end of the pattern, otherwise it will not.
func (p *parser) parsePattern(greedy bool) (pat ast.Pattern) {
	pat = &ast.VarPattern{Name: &ast.Ident{Name: "_"}}
	switch p.tok.Type {
	case token.Identifier:
		if p.tok.Value == "_" {
			pat = &ast.AnythingPattern{Underscore: p.expect(token.Identifier)}
		} else {
			if isUpper(p.tok.Value) {
				pat = p.parseCtorPattern()
			} else {
				pat = &ast.VarPattern{
					Name: p.parseLowerName(),
				}
			}
		}
	case token.LeftBracket:
		pat = p.parseListPattern()
	case token.LeftParen:
		pat = p.parseTupleOrParenthesizedPattern()
	case token.LeftBrace:
		pat = p.parseRecordPattern()
	case token.Int, token.Char, token.String, token.Float:
		pat = &ast.LiteralPattern{p.parseLiteral()}
	case token.True, token.False:
		p.expectOneOf(token.True, token.False)
		pat = &ast.CtorPattern{Ctor: ast.NewIdent(p.tok.Value, p.tok.Position)}
	default:
		p.errorExpectedOneOf(p.tok, token.Identifier, token.LeftParen, token.LeftBrace, token.LeftBracket)
	}

	if p.is(token.As) && greedy {
		return p.parseAliasPattern(pat)
	}

	if p.is(token.Op) && p.tok.Value == "::" {
		return p.parseCtorListPattern(pat)
	}

	return
}

func (p *parser) parseListPattern() ast.Pattern {
	lbracketPos := p.expect(token.LeftBracket)

	if p.is(token.RightBracket) {
		return &ast.ListPattern{
			Lbracket: lbracketPos,
			Rbracket: p.expect(token.RightBracket),
		}
	}

	pat := &ast.ListPattern{Lbracket: lbracketPos}
	pat.Patterns = []ast.Pattern{p.parsePattern(true)}
	for !p.is(token.RightBracket) {
		p.expect(token.Comma)
		pat.Patterns = append(pat.Patterns, p.parsePattern(true))
	}

	pat.Rbracket = p.expect(token.RightBracket)
	return pat
}

func (p *parser) parseCtorListPattern(pat ast.Pattern) ast.Pattern {
	pos := p.tok.Position
	p.expect(token.Op)
	return &ast.CtorPattern{
		Ctor: &ast.Ident{
			Name:    "::",
			NamePos: pos,
		},
		Patterns: []ast.Pattern{
			pat,
			p.parsePattern(false),
		},
	}
}

func (p *parser) parseTupleOrParenthesizedPattern() ast.Pattern {
	lparenPos := p.expect(token.LeftParen)

	var patterns []ast.Pattern
	for !p.is(token.RightParen) {
		patterns = append(patterns, p.parsePattern(true))

		if !p.is(token.RightParen) {
			p.expect(token.Comma)
		}
	}

	rparenPos := p.expect(token.RightParen)
	if len(patterns) > 1 {
		return &ast.TuplePattern{
			Lparen:   lparenPos,
			Patterns: patterns,
			Rparen:   rparenPos,
		}
	}
	return patterns[0]
}

func (p *parser) parseRecordPattern() ast.Pattern {
	lbracePos := p.expect(token.LeftBrace)

	var patterns []ast.Pattern
	for !p.is(token.RightBrace) {
		patterns = append(patterns, p.parsePattern(true))

		if !p.is(token.RightBrace) {
			p.expect(token.Comma)
		}
	}

	return &ast.RecordPattern{
		Lbrace:   lbracePos,
		Patterns: patterns,
		Rbrace:   p.expect(token.RightBrace),
	}
}

func (p *parser) parseCtorPattern() ast.Pattern {
	pat := &ast.CtorPattern{Ctor: p.parseUpperName()}
	var patterns []ast.Pattern

Outer:
	for {
		switch p.tok.Type {
		case token.Identifier, token.LeftParen, token.LeftBracket, token.LeftBrace, token.True, token.False, token.Int, token.Char, token.Float:
			patterns = append(patterns, p.parsePattern(false))
		default:
			break Outer
		}
	}

	pat.Patterns = patterns
	return pat
}

func (p *parser) parseAliasPattern(pat ast.Pattern) ast.Pattern {
	p.expect(token.As)
	return &ast.AliasPattern{
		Name:    p.parseLowerName(),
		Pattern: pat,
	}
}

func (p *parser) next() {
	// if the current position is a new declaration but it has not been
	// acknowledged we need to report it
	if p.inNewDecl {
		p.errorMessage(p.tok.Position, "I encountered what looks like a new declaration, but the previous one has not been finished yet.")
		// silence the error, or it will be repeated forever until the end of
		// tokens
		p.inNewDecl = false
	}

	p.tok = p.scanner.Next()
	if p.is(token.Comment) {
		// ignore comments for now
		p.next()
	}

	// if this is the first token and it's not at the beginning of the line
	// it must be reported, as it must be a declaration
	if p.isStart {
		p.isStart = false
		if !p.atLineStart() {
			p.errorMessage(p.tok.Position, "I expected a new declaration. All declarations need to start at the beginning of their line.")
		}
	} else if !p.ignoreNewDecl && p.atLineStart() {
		p.inNewDecl = true
	}
}

func (p *parser) backup(until *token.Token) {
	p.scanner.Backup(until)
	p.next()
}

func (p *parser) peek() *token.Token {
	t := p.scanner.Peek()
	if t == nil {
		p.errorUnexpectedEOF()
	}
	return t
}

func (p *parser) setLineStart(col int) {
	p.lineStart = append(p.lineStart, col)
}

func (p *parser) resetLineStart() {
	if len(p.lineStart) <= 1 {
		p.lineStart = []int{1}
	} else {
		p.lineStart = p.lineStart[:len(p.lineStart)-1]
	}
}

// finishedDecl marks as finished the parsing of the previous declaration. It
// generates an error otherwise.
func (p *parser) finishedDecl() {
	if p.is(token.EOF) || p.inNewDecl {
		p.inNewDecl = false
		return
	}

	p.errorMessage(p.tok.Position, "I was expecting a new declaration or the end of file, but I got %s instead.", p.tok.Type)
}

func (p *parser) expect(typ token.Type) token.Pos {
	pos := p.tok.Position
	if p.tok.Type != typ {
		p.errorExpected(p.tok, typ)
	}

	p.next()
	return pos.Offset
}

func (p *parser) expectType() ast.Type {
	pos := p.tok.Position
	typ := p.parseType()
	if typ == nil {
		p.errorExpectedType(pos)
	}
	return typ
}

func (p *parser) expectOneOf(types ...token.Type) token.Pos {
	pos := p.tok.Position
	var found bool
	for _, t := range types {
		if p.tok.Type == t {
			found = true
		}
	}

	if !found {
		p.errorExpectedOneOf(p.tok, types...)
	}

	p.next()
	return pos.Offset
}

func (p *parser) is(typ token.Type) bool {
	return p.tok.Type == typ
}

func (p *parser) atLineStart() bool {
	return p.tok.Column <= p.lineStart[len(p.lineStart)-1]
}

func (p *parser) opInfo(name string) *operator.OpInfo {
	info := p.sess.Table.Lookup(name, "" /* TODO: path */)
	if info != nil {
		return info
	}

	return &operator.OpInfo{
		Precedence:    0,
		Associativity: operator.Left,
	}
}

func (p *parser) startRegion() {
	p.regions = append(p.regions, p.tok.Position)
}

func (p *parser) endRegion() {
	p.regions = p.regions[:len(p.regions)-1]
}

func (p *parser) regionStart() *token.Position {
	if len(p.regions) == 0 {
		return &token.Position{Offset: token.NoPos, Line: 1}
	}
	return p.regions[len(p.regions)-1]
}

func (p *parser) region(start *token.Position) []string {
	region, err := p.sess.Source(p.fileName).Region(start.Offset, p.tok.Offset+token.Pos(len(p.tok.Source)))
	if err != nil {
		// TODO(erizocosmico): should never happen, but handle it properly
		panic(err)
	}
	return region
}

func (p *parser) regionError(pos *token.Position, msg diagnostic.Msg) {
	start := p.regionStart()
	p.sess.Diagnose(p.fileName, diagnostic.NewRegionDiagnostic(
		diagnostic.Error,
		msg,
		start,
		pos,
		p.region(start),
	))
}

func (p *parser) errorExpected(t *token.Token, typ token.Type) {
	if t.Type == token.EOF {
		p.regionError(t.Position, diagnostic.UnexpectedEOF(typ))
		panic(bailout{})
	}

	p.errorExpectedOneOf(t, typ)
}

func (p *parser) errorExpectedOneOf(t *token.Token, types ...token.Type) {
	p.regionError(t.Position, diagnostic.Expecting(t.Type, types...))
}

func (p *parser) errorMessage(pos *token.Position, msg string, args ...interface{}) {
	p.regionError(pos, diagnostic.ParseError(fmt.Sprintf(msg, args...)))
}

func (p *parser) errorUnexpectedEOF() {
	p.errorMessage(p.tok.Position, "Unexpected end of file.")
	panic(bailout{})
}

func (p *parser) errorExpectedType(pos *token.Position) {
	p.errorMessage(pos, "I was expecting a type, but I encountered what looks like a declaration instead.")
	panic(bailout{})
}

func isLower(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsLower(r)
}

func isUpper(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}
