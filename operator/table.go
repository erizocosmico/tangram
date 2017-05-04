package operator

import "fmt"

// Table is the implementation of an operator table. It contains the operators
// and their info.
type Table struct {
	ops         map[Op]*OpInfo
	opsByModule map[string]map[string]string
}

// NewTable creates a new empty operator table.
func NewTable() *Table {
	return &Table{
		ops:         make(map[Op]*OpInfo),
		opsByModule: make(map[string]map[string]string),
	}
}

// BuiltinTable creates a new operator table with all the builtin operators
// loaded. This is used for parsing single modules. If all the imports are not
// parsed, core will not be parsed and thus all expressions using the builtin
// operators might not be correctly parsed.
func BuiltinTable() *Table {
	t := NewTable()
	for _, op := range builtinOps {
		t.Add(op.name, "Basics", op.assoc, op.prec)
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
	if _, ok := t.ops[Op{name, path}]; ok {
		return fmt.Errorf("operator %s is already defined somewhere else", name)
	}

	t.ops[Op{name, path}] = &OpInfo{assoc, precedence}
	return nil
}

// AddToModule adds an operator as available in the given `module`.
func (t *Table) AddToModule(module, opModule, opName string) {
	if _, ok := t.opsByModule[module]; !ok {
		t.opsByModule[module] = make(map[string]string)
	}

	t.opsByModule[module][opName] = opModule
}

// lookup finds a specific operator and returns its info. Will return nil if
// the operator does not exist.
func (t *Table) lookup(name, path string) *OpInfo {
	return t.ops[Op{name, path}]
}

// Lookup finds an operator that is available (imported or defined) in the current module.
func (t *Table) Lookup(name string, currentModule string) *OpInfo {
	if ops, ok := t.opsByModule[currentModule]; ok {
		if mod, ok := ops[name]; ok {
			return t.lookup(name, mod)
		}
	}

	return nil
}

// Op represents a qualified operator with the module where it was defined.
type Op struct {
	// Name of the operator.
	Name string
	// Module of the operator.
	Module string
}

// OpInfo contains the info about an operator.
type OpInfo struct {
	// Associativity of the operator.
	Associativity Associativity
	// Precedence of the operator.
	Precedence uint
}

// Associativity is the type of associativity of the operator.
type Associativity byte

const (
	// NonAssoc is a non associative operator.
	NonAssoc Associativity = iota
	// Left associativity.
	Left
	// Right associativity.
	Right
)
