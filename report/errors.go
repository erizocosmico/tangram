package report

import (
	"fmt"
	"strings"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/token"
)

// Name errors

type UndefinedError struct {
	BaseReport
	Name string
}

func NewUndefinedError(expr ast.Node, name *ast.Ident) *UndefinedError {
	return &UndefinedError{
		NewBaseReport(NameError, name.Pos(), "", RegionFromNode(expr)),
		name.Name,
	}
}

func (e UndefinedError) Message() string {
	return fmt.Sprintf("Name %q is not defined.", e.Name)
}

type ModuleNotImportedError struct {
	BaseReport
	Module string
}

func NewModuleNotImportedError(expr ast.Node, name string) *ModuleNotImportedError {
	return &ModuleNotImportedError{
		NewBaseReport(NameError, expr.Pos(), "", RegionFromNode(expr)),
		name,
	}
}

func (e ModuleNotImportedError) Message() string {
	return fmt.Sprintf("I could not find imported module %q.", e.Module)
}

type ImportError struct {
	BaseReport
	Module string
	Name   string
}

func NewImportError(decl ast.Node, module string, name *ast.Ident) *ImportError {
	return &ImportError{
		NewBaseReport(NameError, name.Pos(), "", RegionFromNode(decl)),
		module,
		name.Name,
	}
}

type ExportError struct {
	BaseReport
	Module string
	Name   string
}

func NewExportError(decl *ast.ModuleDecl, name *ast.Ident) *ExportError {
	return &ExportError{
		NewBaseReport(NameError, name.Pos(), "", RegionFromNode(decl)),
		decl.ModuleName(),
		name.Name,
	}
}

func (e *ExportError) Message() string {
	return fmt.Sprintf("I cannot expose %q in module %q because there is no such thing delcared in this module.", e.Name, e.Module)
}

type ExpectedUnionError struct {
	BaseReport
	Name       string
	ActualKind ast.ObjKind
}

func NewExpectedUnionError(decl ast.Decl, obj *ast.Object) *ExpectedUnionError {
	return &ExpectedUnionError{
		NewBaseReport(NameError, obj.Node.Pos(), "", RegionFromNode(decl)),
		obj.Name,
		obj.Kind,
	}
}

func (e *ExpectedUnionError) Message() string {
	return fmt.Sprintf("I was expecting %q to be an union type, instead it is %q.", e.Name, e.ActualKind)
}

type ExpectedCtorError struct {
	BaseReport
	Name       string
	ActualKind ast.ObjKind
}

func NewExpectedCtorError(decl ast.Decl, obj *ast.Object) *ExpectedCtorError {
	return &ExpectedCtorError{
		NewBaseReport(NameError, obj.Node.Pos(), "", RegionFromNode(decl)),
		obj.Name,
		obj.Kind,
	}
}

func (e *ExpectedCtorError) Message() string {
	return fmt.Sprintf("I was expecting %q to be a constructor, instead it is %q.", e.Name, e.ActualKind)
}

type RepeatedFieldError struct {
	BaseReport
	Field string
}

func NewRepeatedFieldError(record ast.Node, field *ast.Ident) *RepeatedFieldError {
	return &RepeatedFieldError{
		NewBaseReport(NameError, field.Pos(), "", RegionFromNode(record)),
		field.Name,
	}
}

func (e *RepeatedFieldError) Message() string {
	return fmt.Sprintf("Record already has a field named %q.", e.Field)
}

type AlreadyDeclaredError struct {
	BaseReport
	Name string
}

func NewAlreadyDeclaredError(decl ast.Decl, name *ast.Ident) *AlreadyDeclaredError {
	return &AlreadyDeclaredError{
		NewBaseReport(NameError, name.Pos(), "", RegionFromNode(decl)),
		name.Name,
	}
}

func (e *AlreadyDeclaredError) Message() string {
	return fmt.Sprintf("Name %q has already been declared in this module, please make sure your names are unique.", e.Name)
}

type RepeatedVarTypeError struct {
	BaseReport
	Var string
}

func NewRepeatedVarTypeError(decl ast.Decl, name *ast.Ident) *RepeatedVarTypeError {
	return &RepeatedVarTypeError{
		NewBaseReport(NameError, name.Pos(), "", RegionFromNode(decl)),
		name.Name,
	}
}

func (e RepeatedVarTypeError) Message() string {
	return fmt.Sprintf("I found a redeclared variable type %q on this type. Variable types only have to be declared once and they cannot be repeated.", e.Var)
}

type RepeatedCtorError struct {
	BaseReport
	Ctor string
}

func NewRepeatedCtorError(decl ast.Decl, name *ast.Ident) *RepeatedCtorError {
	return &RepeatedCtorError{
		NewBaseReport(NameError, name.Pos(), "", RegionFromNode(decl)),
		name.Name,
	}
}

func (e RepeatedCtorError) Message() string {
	return fmt.Sprintf("I found a repeated constructor in the same type union declaration. Constructor names must be unique.", e.Ctor)
}

type UnresolvedNameError struct {
	BaseReport
	Name string
}

func NewUnresolvedNameError(name string, node *ast.Ident) *UnresolvedNameError {
	return &UnresolvedNameError{
		NewBaseReport(NameError, node.Pos(), "", nil),
		name,
	}
}

func (e *UnresolvedNameError) Message() string {
	return fmt.Sprintf("I could not find any definition for %q.", e.Name)
}

// Parse errors

func NewExpectedTypeError(pos token.Pos, region *Region) Report {
	return NewBaseReport(
		SyntaxError,
		pos,
		"I was expecting a type, but I encountered what looks like a declaration instead.",
		region,
	)
}

func NewUnexpectedEOFError(pos token.Pos, region *Region) Report {
	return NewBaseReport(
		SyntaxError,
		pos,
		"Unexpected end of file.",
		region,
	)
}

type UnexpectedTokenError struct {
	BaseReport
	Token    *token.Token
	Expected []token.Type
}

func NewUnexpectedTokenError(tok *token.Token, region *Region, expected ...token.Type) *UnexpectedTokenError {
	return &UnexpectedTokenError{
		NewBaseReport(SyntaxError, tok.Offset, "", region),
		tok,
		expected,
	}
}

func (e UnexpectedTokenError) Message() string {
	var list = make([]string, len(e.Expected))
	for i, e := range e.Expected {
		list[i] = fmt.Sprintf(" - %s", e)
	}

	return fmt.Sprintf(
		"I encountered an unexpected token %q, but I was expecting one of the following tokens:\n%s",
		e.Token.Type,
		strings.Join(list, "\n"),
	)
}
