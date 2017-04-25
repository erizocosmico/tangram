package ast

type Scope interface {
	Lookup(string) *Object
	Resolve(string, *Ident)
	Add(*Object) bool
}

type ModuleScope struct {
	*NodeScope
	Exported map[string]*Object
	Imported map[string]*Object
}

func NewModuleScope(root Node) *ModuleScope {
	return &ModuleScope{
		NodeScope: NewNodeScope(root, nil),
		Exported:  make(map[string]*Object),
		Imported:  make(map[string]*Object),
	}
}

func (s *ModuleScope) Lookup(name string) *Object {
	if obj := s.Imported[name]; obj != nil {
		return obj
	}

	return s.NodeScope.Lookup(name)
}

func (s *ModuleScope) Resolve(name string, id *Ident) {
	if obj := s.Imported[name]; obj != nil {
		id.Obj = obj
	} else {
		s.NodeScope.Resolve(name, id)
	}
}

type NodeScope struct {
	Parent Scope
	Root   Node
	// Exposed contains all the objects exposed by this scope.
	Exposed map[string]*Object
	// Objects contains all the objects defined in this scope.
	Objects    map[string]*Object
	Unresolved map[string][]*Ident
}

func NewNodeScope(root Node, parent Scope) *NodeScope {
	return &NodeScope{
		Parent:     parent,
		Root:       root,
		Objects:    make(map[string]*Object),
		Unresolved: make(map[string][]*Ident),
	}
}

func (s *NodeScope) Lookup(name string) *Object {
	result := s.Objects[name]
	if result == nil && s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	return result
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

func (s *NodeScope) Resolve(name string, id *Ident) {
	if obj := s.Lookup(name); obj != nil {
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
	Def
	Op
	Ctor
	Var
)

var objKindStrings = [...]string{
	"invalid",
	"module",
	"type",
	"definition",
	"operator",
	"constructor",
	"variable",
}

func (k ObjKind) String() string {
	if k < 0 || int(k) >= len(objKindStrings) {
		return objKindStrings[0]
	}
	return objKindStrings[k]
}
