package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mvader/elm-compiler/lexer"
	"github.com/mvader/elm-compiler/token"
)

var help = flag.Bool("help", false, "display help")

func main() {
	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	lexer := lexer.New("stdin", os.Stdin)
	go lexer.Run()

	for {
		t := lexer.Next()
		if t == nil || t.Type == token.EOF {
			break
		}

		fmt.Printf(
			"LINE: %4d POS: %4d TYPE: %-30s %s\n",
			t.Line,
			t.Column,
			t.Type,
			t.Value,
		)
	}
}

const helpText = `Display a list of tokens with their properties

usage: elmlex < /path/to/file.elm

To enter lines interactively use: cat | elmlex`

func printUsage() {
	fmt.Println(helpText)
}
