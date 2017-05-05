package parser

import (
	"fmt"

	"github.com/elm-tangram/tangram/ast"
)

// opTable is the implementation of an operator table. It contains the operators
// and their info.
type opTable struct {
	ops         map[operator]*operatorInfo
	opsByModule map[string]map[string]string
}

// newOpTable creates a new empty operator table.
func newOpTable() *opTable {
	return &opTable{
		ops:         make(map[operator]*operatorInfo),
		opsByModule: make(map[string]map[string]string),
	}
}

// builtinOpTable creates a new operator table with all the builtin operators
// loaded. This is used for parsing single modules. If all the imports are not
// parsed, core will not be parsed and thus all expressions using the builtin
// operators might not be correctly parsed.
func builtinOpTable() *opTable {
	t := newOpTable()
	for _, op := range builtinOps {
		t.add(op.name, "Basics", op.assoc, op.prec)
	}
	return t
}

var builtinOps = []struct {
	name  string
	assoc ast.Associativity
	prec  uint
}{
	{">>", ast.Left, 9},
	{"<<", ast.Right, 9},
	{"^", ast.Right, 8},
	{"*", ast.Left, 7},
	{"%", ast.Left, 7},
	{"/", ast.Left, 7},
	{"//", ast.Left, 7},
	{"+", ast.Left, 6},
	{"-", ast.Left, 6},
	{"++", ast.Right, 5},
	{"::", ast.Right, 5},
	{"==", ast.NonAssoc, 4},
	{"/=", ast.NonAssoc, 4},
	{"<", ast.NonAssoc, 4},
	{">", ast.NonAssoc, 4},
	{"<=", ast.NonAssoc, 4},
	{">=", ast.NonAssoc, 4},
	{"&&", ast.Right, 3},
	{"||", ast.Right, 2},
	{"<|", ast.Right, 0},
	{"|>", ast.Left, 0},
}

// add inserts the given operator and its data in the operator table. It
// returns an error if the operator is a builtin or has already been defined.
func (t *opTable) add(name, path string, assoc ast.Associativity, precedence uint) error {
	if _, ok := t.ops[operator{name, path}]; ok {
		return fmt.Errorf("operator %s is already defined somewhere else", name)
	}

	t.ops[operator{name, path}] = &operatorInfo{assoc, precedence}
	return nil
}

// addToModule adds an operator as available in the given `module`.
func (t *opTable) addToModule(module, opModule, opName string) {
	if _, ok := t.opsByModule[module]; !ok {
		t.opsByModule[module] = make(map[string]string)
	}

	t.opsByModule[module][opName] = opModule
}

// find finds a specific operator and returns its info. Will return nil if
// the operator does not exist.
func (t *opTable) find(name, path string) *operatorInfo {
	return t.ops[operator{name, path}]
}

// lookup finds an operator that is available (imported or defined) in the current module.
func (t *opTable) lookup(name string, currentModule string) *operatorInfo {
	if ops, ok := t.opsByModule[currentModule]; ok {
		if mod, ok := ops[name]; ok {
			return t.find(name, mod)
		}
	}

	return nil
}

// operator represents a qualified operator with the module where it was defined.
type operator struct {
	// Name of the operator.
	Name string
	// Module of the operator.
	Module string
}

// operatorInfo contains the info about an operator.
type operatorInfo struct {
	// Associativity of the operator.
	Associativity ast.Associativity
	// Precedence of the operator.
	Precedence uint
}
