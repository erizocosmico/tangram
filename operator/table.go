package operator

import "fmt"

// Table is the implementation of an operator table. It contains the operators
// and their info.
type Table struct {
	ops      map[Op]*OpInfo
	builtins map[string]struct{}
}

// NewTable creates a new empty operator table.
func NewTable() *Table {
	return &Table{
		ops:      make(map[Op]*OpInfo),
		builtins: make(map[string]struct{}),
	}
}

// BuiltinTable creates a new operator table with all the builtin operators
// loaded. This is used for parsing single modules. If all the imports are not
// parsed, core will not be parsed and thus all expressions using the builtin
// operators might not be correctly parsed.
func BuiltinTable() *Table {
	t := NewTable()
	for _, op := range builtinOps {
		t.AddBuiltin(op.name, op.assoc, op.prec)
	}
	return t
}

var builtinOps = []struct {
	name  string
	assoc Associativity
	prec  uint
}{
	{">>", Left, 9},
	{"<<", Right, 9},
	{"^", Right, 8},
	{"*", Left, 7},
	{"%", Left, 7},
	{"/", Left, 7},
	{"//", Left, 7},
	{"+", Left, 6},
	{"-", Left, 6},
	{"++", Right, 5},
	{"::", Right, 5},
	{"==", NonAssoc, 4},
	{"/=", NonAssoc, 4},
	{"<", NonAssoc, 4},
	{">", NonAssoc, 4},
	{"<=", NonAssoc, 4},
	{">=", NonAssoc, 4},
	{"&&", Right, 3},
	{"||", Right, 2},
	{"<|", Right, 0},
	{"|>", Left, 0},
}

// Add inserts the given operator and its data in the operator table. It
// returns an error if the operator is a builtin or has already been defined.
func (t *Table) Add(name, path string, assoc Associativity, precedence uint) error {
	if t.IsBuiltin(name) {
		return fmt.Errorf("operator %s is a builtin operator and can not be overriden", name)
	}

	if _, ok := t.ops[Op{name, path}]; ok {
		return fmt.Errorf("operator %s is already defined somewhere else", name)
	}

	t.ops[Op{name, path}] = &OpInfo{assoc, precedence, false}
	return nil
}

// LookupByName returns the list of possible operators with the given name.
func (t *Table) LookupByName(name string) []Op {
	var result []Op
	for op := range t.ops {
		if op.Name == name {
			result = append(result, op)
		}
	}
	return result
}

// Lookup finds a specific operator and returns its info. Will return nil if
// the operator does not exist.
func (t *Table) Lookup(name, path string) *OpInfo {
	return t.ops[Op{name, path}]
}

// IsBuiltin reports whether the operator with the given name is a builtin.
func (t *Table) IsBuiltin(name string) bool {
	_, ok := t.builtins[name]
	return ok
}

// AddBuiltin adds a new builtin operator to the operator table.
func (t *Table) AddBuiltin(name string, assoc Associativity, precedence uint) error {
	if len(t.LookupByName(name)) > 0 {
		return fmt.Errorf("cannot add builtin operator %s, is already defined", name)
	}
	t.builtins[name] = struct{}{}
	t.ops[Op{name, ""}] = &OpInfo{assoc, precedence, true}
	return nil
}

// Op represents a qualified operator with its path.
type Op struct {
	// Name of the operator.
	Name string
	// Path of the operator.
	Path string
}

// OpInfo contains the info about an operator.
type OpInfo struct {
	// Associativity of the operator.
	Associativity Associativity
	// Precedence of the operator.
	Precedence uint
	// Builtin will be true if the op is a builtin.
	Builtin bool
}

// Associativity is the type of associativity of the operator.
type Associativity byte

const (
	// Left associativity.
	Left Associativity = iota
	// Right associativity.
	Right
	// NonAssoc is a non associative operator.
	NonAssoc
)
