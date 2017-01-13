package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erizocosmico/elmo/scanner"
	"github.com/erizocosmico/elmo/token"
)

var help = flag.Bool("help", false, "display help")

func main() {
	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	scanner := scanner.New("stdin", os.Stdin)
	go scanner.Run()

	for {
		t := scanner.Next()
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

usage: elmscan < /path/to/file.elm

To enter lines interactively use: cat | elmscan`

func printUsage() {
	fmt.Println(helpText)
}
