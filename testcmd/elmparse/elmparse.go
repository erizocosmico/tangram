package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/fatih/color"
	"github.com/mvader/elm-compiler/ast"
	"github.com/mvader/elm-compiler/parser"
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
	case *ast.UnionDecl:
		printUnionDecl(d)
	case *ast.AliasDecl:
		printAliasDecl(d)
	default:
		fmt.Println("Not implemented decl printer of type:", reflect.TypeOf(d))
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

func printUnionDecl(decl *ast.UnionDecl) {
	printIndent(1)
	fmt.Println(color.YellowString("Union:"), decl.Name.Name)

	for _, c := range decl.Types {
		printIndent(2)
		fmt.Println("- " + c.Name.Name)
		for _, a := range c.Args {
			printType(3, a)
		}
	}
}

func printAliasDecl(decl *ast.AliasDecl) {
	printIndent(1)
	fmt.Println(color.YellowString("Alias:"), decl.Name.Name)
	printType(2, decl.Type)
}

func printType(indent int, typ ast.Type) {
	switch t := typ.(type) {
	case *ast.TupleType:
		printTupleType(indent, t)
	case *ast.RecordType:
		printRecordType(indent, t)
	case *ast.BasicType:
		printBasicType(indent, t)
	}
}

func printTupleType(indent int, tuple *ast.TupleType) {
	printIndent(indent)
	color.Yellow("Tuple:")
	for _, e := range tuple.Elems {
		printType(indent+1, e)
	}
}

func printBasicType(indent int, t *ast.BasicType) {
	fmt.Println(t.Name.Name)
	for _, a := range t.Args {
		printType(indent+1, a)
	}
}

func printRecordType(indent int, t *ast.RecordType) {
	printIndent(indent)
	color.Yellow("Record:")
	for _, f := range t.Fields {
		printIndent(indent + 1)
		fmt.Printf("- %s: ", f.Name.Name)
		printType(indent+2, f.Type)
	}
}

const helpText = `Display a parsed AST

usage: elmparse < /path/to/file.elm`

func printUsage() {
	fmt.Println(helpText)
}
