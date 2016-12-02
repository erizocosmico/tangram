package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/fatih/color"
	"github.com/mvader/elmo/ast"
	"github.com/mvader/elmo/parser"
)

var help = flag.Bool("help", false, "display help")

func main() {
	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	file, err := parser.ParseFile("stdin", os.Stdin)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(color.YellowString("File:"), file.Name)
	printModule(file.Module)
	for _, i := range file.Imports {
		printImport(i)
	}

	for _, d := range file.Decls {
		printDecl(d)
	}
}

func printModule(mod *ast.ModuleDecl) {
	printIndent(1)
	fmt.Println(color.YellowString("Module:"), mod.Name.String())
	printExposing(mod.Exposing)
}

func printImport(i *ast.ImportDecl) {
	printIndent(1)
	fmt.Println(color.YellowString("Import:"), i.Module.String())
	if i.Alias != nil {
		printIndent(2)
		fmt.Println(color.YellowString("Alias:"), i.Alias.Name)
	}

	printExposing(i.Exposing)
}

func printExposing(e *ast.ExposingList) {
	if e != nil {
		printIndent(2)
		color.Yellow("Exposing:")

		for _, e := range e.Idents {
			printIndent(3)
			fmt.Println(color.CyanString("-"), e.Name)
			if e.Exposing != nil {
				printIndent(4)
				color.Yellow("Exposing:")

				for _, e := range e.Exposing.Idents {
					printIndent(5)
					fmt.Println(color.CyanString("-"), e.Name)
				}
			}
		}
	}
}

func printIndent(offset int) {
	for i := 0; i < offset; i++ {
		fmt.Print(".   ")
	}
}

func printDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.InfixDecl:
		printInfixDecl(d)
	case *ast.AliasDecl:
		printAliasDecl(d)
	case *ast.UnionDecl:
		printUnionDecl(d)
	default:
		fmt.Println("Not implemented decl printer of type:", reflect.TypeOf(d))
	}
}

func printAliasDecl(decl *ast.AliasDecl) {
	printIndent(1)
	fmt.Println(color.YellowString("Alias Type:"), decl.Name.Name)
	if len(decl.Args) > 0 {
		printIndent(2)
		fmt.Println("- Args:")
		for _, a := range decl.Args {
			printIndent(2)
			fmt.Println("-", a.Name)
		}
	}

	printIndent(2)
	fmt.Println("- Type:")
	printType(3, decl.Type)
}

func printUnionDecl(decl *ast.UnionDecl) {
	printIndent(1)
	fmt.Println(color.YellowString("Union Type:"), decl.Name.Name)
	if len(decl.Args) > 0 {
		printIndent(2)
		fmt.Println("- Args:")
		for _, a := range decl.Args {
			printIndent(3)
			fmt.Println("-", a.Name)
		}
	}

	printIndent(2)
	fmt.Println("- Constructors:")
	for _, t := range decl.Types {
		printConstructor(t)
	}
}

func printConstructor(c *ast.Constructor) {
	printIndent(3)
	fmt.Println("-", c.Name.Name)
	for _, a := range c.Args {
		printType(4, a)
	}
}

func printType(indent int, typ ast.Type) {
	switch t := typ.(type) {
	case *ast.TupleType:
		printTuple(indent, t)
	case *ast.RecordType:
		printRecord(indent, t)
	case *ast.FuncType:
		printFuncType(indent, t)
	case *ast.BasicType:
		printBasicType(indent, t)
	}
}

func printTuple(indent int, t *ast.TupleType) {
	printIndent(indent)
	color.Yellow("- Tuple:")
	for _, el := range t.Elems {
		printType(indent+1, el)
	}
}

func printRecord(indent int, t *ast.RecordType) {
	printIndent(indent)
	color.Yellow("- Record:")
	for _, f := range t.Fields {
		printRecordField(indent+1, f)
	}
}

func printRecordField(indent int, f *ast.RecordTypeField) {
	printIndent(indent)
	fmt.Printf("- %s:\n", f.Name.Name)
	printType(indent+1, f.Type)
}

func printFuncType(indent int, f *ast.FuncType) {
	printIndent(indent)
	color.Yellow("- Function:")

	printIndent(indent + 1)
	fmt.Println("- Args:")
	for _, a := range f.Args {
		printType(indent+2, a)
	}

	printIndent(indent + 1)
	fmt.Println("- Return:")
	printType(indent+2, f.Return)
}

func printBasicType(indent int, t *ast.BasicType) {
	printIndent(indent)
	fmt.Println("-", t.Name.Name)
	for _, a := range t.Args {
		printType(indent, a)
	}
}

func printInfixDecl(decl *ast.InfixDecl) {
	printIndent(1)
	color.Yellow("Infix:")
	printIndent(2)
	fmt.Print("- Associativity: ")
	switch decl.Assoc {
	case ast.LeftAssoc:
		fmt.Println("left")
	case ast.RightAssoc:
		fmt.Println("right")
	case ast.NonAssoc:
		fmt.Println("non-assoc")
	}
	printIndent(2)
	fmt.Println("- Operator:", decl.Op.Name)
	printIndent(2)
	fmt.Println("- Priority:", decl.Priority.Value)
}

const helpText = `Display a parsed AST

usage: elmparse < /path/to/file.elm`

func printUsage() {
	fmt.Println(helpText)
}
