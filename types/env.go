package types

import (
	"github.com/elm-tangram/tangram/ast"
	"github.com/elm-tangram/tangram/token"
)

type State struct {
	*nameState
	env *Env
}

func NewState() *State {
	return &State{
		newNameState(),
		AddBuiltins(NewEnv(nil)),
	}
}

func (s *State) NewChild() *State {
	return &State{
		s.nameState,
		NewEnv(s.env),
	}
}

type Env struct {
	Parent   *Env
	Types    map[string]*Scheme
	Vars     map[string]*VarType
	VarTypes map[string]*Scheme
}

func NewEnv(parent *Env) *Env {
	return &Env{
		Parent:   parent,
		Types:    make(map[string]*Scheme),
		VarTypes: make(map[string]*VarType),
		Vars:     make(map[string]*Scheme),
	}
}

func (e *Env) SetType(name string, typ Type) {
	e.Types[name] = generalize(e, typ)
}

func (e *Env) SetVarType(name string, v *VarType) {
	e.VarTypes[name] = v
}

func (e *Env) SetVar(name string, typ Type) {
	e.Vars[name] = generalize(e, typ)
}

func (e *Env) Type(name string) Type {
	if t, ok := e.Types[name]; ok {
		return instantiate(e, t)
	}

	if e.Parent != nil {
		return e.Parent.Type(name)
	}

	panic("unreachable")
}

func (e *Env) Var(name string) Type {
	if v, ok := e.Vars[name]; ok {
		return instantiate(e, v)
	}

	if e.Parent != nil {
		return e.Parent.Var(name)
	}

	panic("unreachable")
}

func (e *Env) VarType(name string) *VarType {
	if v, ok := e.VarTypes[name]; ok {
		return v
	}

	if e.Parent != nil {
		return e.Parent.VarType(name)
	}

	panic("unreachable")
}

func (e *Env) ReplaceVar(name string, typ Type) {
	e.Vars[name] = generalize(e, typ)
}

func (e *Env) ReplaceVarType(name string, v *VarType) {
	if t, ok := e.VarTypes[name]; ok {
		*t = *v
	}
	panic("unreachable")
}

func builtinTypes() map[string]Type {
	return map[string]Type{
		NamedType{nil, nil},
	}
}

func mkBuiltin(name string, args ...Type) *NamedType {
	return &NamedType{
		&Var{ast.NewIdent(name, token.NoPos), name, ""},
		args,
	}
}

func mkArg(name string) *VarType {
	return &VarType{name, false}
}

func AddBuiltins(env *Env) *Env {
	for _, t := range builtinTypes {
		env.SetType(t.Name.Name, t)
	}
	return env
}

var builtinTypes = []*NamedType{
	mkBuiltin("List", mkBuiltin("List", mkArg("a"))),
	mkBuiltin("String"),
	mkBuiltin("Int"),
	mkBuiltin("Float"),
	mkBuiltin("Bool"),
	mkBuiltin("Char"),
}
