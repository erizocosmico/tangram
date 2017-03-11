package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/parser"
	"github.com/fatih/color"
)

var help = flag.Bool("help", false, "display help")

func main() {
	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	file, err := parser.ParseFile("stdin", os.Stdin, parser.FullParse)
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
	fmt.Println(color.YellowString("Module:"), mod.Name.(fmt.Stringer).String())
	printExposing(mod.Exposing)
}

func printImport(i *ast.ImportDecl) {
	printIndent(1)
	fmt.Println(color.YellowString("Import:"), i.Module.(fmt.Stringer).String())
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
	case *ast.Definition:
		printDef(d)
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
			printIndent(3)
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

func printRecordField(indent int, f *ast.RecordField) {
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
	fmt.Println("-", t.Name.(fmt.Stringer).String())
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
	fmt.Println("- Precedence:", decl.Precedence.Value)
}

func printDef(def *ast.Definition) {
	printIndent(1)
	fmt.Println(color.YellowString("Definition:"), def.Name.Name)
	printIndent(2)
	if def.Annotation == nil {
		fmt.Println("- No type annotation")
	} else {
		fmt.Println("- Type annotation:")
		printType(3, def.Annotation.Type)
	}

	if len(def.Args) > 0 {
		printIndent(2)
		fmt.Println("- Arguments:")
		for _, arg := range def.Args {
			printPattern(3, arg)
		}
	}

	printIndent(2)
	fmt.Println("- Body:")
	printExpr(3, def.Body)
}

func printExpr(indent int, expr ast.Expr) {
	printIndent(indent)

	switch e := expr.(type) {
	case *ast.BasicLit:
		fmt.Println("- Basic Literal:", e.Type, e.Value)
	default:
		fmt.Println(reflect.TypeOf(expr), "is not supported")
	}
}

func printPattern(indent int, pattern ast.Pattern) {
	printIndent(indent)
	switch pat := pattern.(type) {
	case *ast.VarPattern:
		fmt.Println("- Var:", pat.Name.Name)
	case *ast.AnythingPattern:
		fmt.Println("- Anything")
	case *ast.CtorPattern:
		fmt.Println("- Ctor:")
		for _, p := range pat.Patterns {
			printPattern(indent+1, p)
		}
	case *ast.TuplePattern:
		fmt.Println("- Tuple:")
		for _, p := range pat.Patterns {
			printPattern(indent+1, p)
		}
	case *ast.RecordPattern:
		fmt.Println("- Record:")
		for _, p := range pat.Patterns {
			printPattern(indent+1, p)
		}
	case *ast.AliasPattern:
		fmt.Println("- Alias:", pat.Name.Name)
		printPattern(indent+1, pat.Pattern)
	case *ast.ListPattern:
		fmt.Println("- List:")
		for _, p := range pat.Patterns {
			printPattern(indent+1, p)
		}
	default:
		fmt.Println(reflect.TypeOf(pattern), "is not a valid pattern")
	}
}

const helpText = `Display a parsed AST

usage: elmparse < /path/to/file.elm`

func printUsage() {
	fmt.Println(helpText)
}
