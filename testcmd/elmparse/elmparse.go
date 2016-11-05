package main

import (
	"flag"
	"fmt"
	"os"

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

const helpText = `Display a parsed AST

usage: elmparse < /path/to/file.elm`

func printUsage() {
	fmt.Println(helpText)
}
