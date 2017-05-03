package ast

type Scope interface {
	Lookup(string, ObjKind) *Object
	Resolve(string, *Ident, ObjKind)
	Add(*Object) bool
	AddChildren(*NodeScope)
	Children() []*NodeScope
}

type ModuleScope struct {
	*NodeScope
	Exposed  map[string]*Object
	Imported map[string]*Object
	Modules  map[string]*Object
}

func NewModuleScope(root Node) *ModuleScope {
	return &ModuleScope{
		NodeScope: NewNodeScope(root, nil),
		Exposed:   make(map[string]*Object),
		Imported:  make(map[string]*Object),
		Modules:   make(map[string]*Object),
	}
}

func (s *ModuleScope) Expose(obj *Object) {
	s.Exposed[obj.Name] = obj
}

func (s *ModuleScope) ImportModule(obj *Object) {
	s.Modules[obj.Name] = obj
}

func (s *ModuleScope) Import(obj *Object) {
	s.Imported[obj.Name] = obj
}

func (s *ModuleScope) Lookup(name string, kind ObjKind) *Object {
	if kind == Mod || kind == NativeMod {
		return s.Modules[name]
	}

	if obj := s.NodeScope.Lookup(name, kind); obj != nil {
		return obj
	}

	if obj := s.Imported[name]; obj != nil && obj.Kind == kind {
		return obj
	}

	return nil
}

func (s *ModuleScope) LookupSelf(name string, kind ObjKind) *Object {
	return s.NodeScope.Lookup(name, kind)
}

func (s *ModuleScope) LookupExposed(name string, kind ObjKind) *Object {
	if obj := s.Exposed[name]; obj != nil && obj.Kind == kind {
		return obj
	}
	return nil
}

func (s *ModuleScope) Resolve(name string, id *Ident, kind ObjKind) {
	if obj := s.Imported[name]; obj != nil && obj.Kind == kind {
		id.Obj = obj
	} else {
		s.NodeScope.Resolve(name, id, kind)
	}
}

type NodeScope struct {
	Parent Scope
	Root   Node
	// Objects contains all the objects defined in this scope.
	Objects    map[string]*Object
	Unresolved map[string][]*Ident
	children   []*NodeScope
}

func NewNodeScope(root Node, parent Scope) *NodeScope {
	s := &NodeScope{
		Parent:     parent,
		Root:       root,
		Objects:    make(map[string]*Object),
		Unresolved: make(map[string][]*Ident),
	}

	if parent != nil {
		parent.AddChildren(s)
	}
	return s
}

func (s *NodeScope) AddChildren(scope *NodeScope) {
	s.children = append(s.children, scope)
}

func (s *NodeScope) Children() []*NodeScope {
	return s.children
}

func (s *NodeScope) Lookup(name string, kind ObjKind) *Object {
	if obj := s.Objects[name]; obj != nil && obj.Kind == kind {
		return obj
	}

	if s.Parent != nil {
		return s.Parent.Lookup(name, kind)
	}
	return nil
}

func (s *NodeScope) Add(obj *Object) bool {
	if obj := s.Objects[obj.Name]; obj != nil {
		return false
	}

	if nodes, ok := s.Unresolved[obj.Name]; ok {
		for _, n := range nodes {
			n.Obj = obj
		}
		delete(s.Unresolved, obj.Name)
	}
	s.Objects[obj.Name] = obj
	return true
}

func (s *NodeScope) Resolve(name string, id *Ident, kind ObjKind) {
	if obj := s.Lookup(name, kind); obj != nil {
		id.Obj = obj
	} else {
		s.Unresolved[name] = append(s.Unresolved[name], id)
	}
}

type Object struct {
	Name string
	Kind ObjKind
	Node Node
	Data interface{}
}

func NewObject(name string, kind ObjKind, node Node) *Object {
	return &Object{
		Name: name,
		Kind: kind,
		Node: node,
	}
}

type ObjKind byte

const (
	Bad ObjKind = iota
	Mod
	Typ
	Ctor
	Var
	VarTyp
	BuiltinTyp
	NativeMod
)

var objKindStrings = [...]string{
	"invalid",
	"module",
	"type",
	"constructor",
	"variable",
	"type variable",
	"builtin type",
	"native module",
}

func (k ObjKind) String() string {
	if k < 0 || int(k) >= len(objKindStrings) {
		return objKindStrings[0]
	}
	return objKindStrings[k]
}
