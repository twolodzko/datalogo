package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/twolodzko/datalogo/datalog"
	"github.com/twolodzko/datalogo/eval"
	"github.com/twolodzko/datalogo/parser"
)

func main() {
	db := make(datalog.Database)

	if len(os.Args) > 1 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			fmt.Printf("usage: %s [-h|--help] [FILE]...\n", os.Args[0])
			return
		}
		evalFiles(os.Args[1:], db)
	} else {
		repl(db)
	}
}

func repl(db datalog.Database) {
	fmt.Println("Press ^C to exit.")
	fmt.Println()

	parser := parser.NewParser(os.Stdin)
	for {
		fmt.Print("| ")
		expr, err := parser.Next()
		if err != nil {
			printError(err)
			// flush the reader
			// see: https://stackoverflow.com/a/14640839
			parser.Reader.ReadString('\n')
			continue
		}

		out := make(chan datalog.Atom)
		if err := eval.Eval(expr, db, out); err != nil {
			printError(err)
		} else {
			for result := range out {
				fmt.Println(result)
			}
		}
	}
}

func evalFiles(paths []string, db datalog.Database) {
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			printError(err)
			return
		}
		reader := bufio.NewReader(file)
		parser := parser.NewParser(reader)
		for {
			expr, err := parser.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				printError(err)
				return
			}
			out := make(chan datalog.Atom)
			if err := eval.Eval(expr, db, out); err != nil {
				printError(err)
				return
			} else {
				for result := range out {
					fmt.Println(result)
				}
			}
		}
	}
}

func printError(msg error) {
	fmt.Printf("error: %s\n", msg)
}
