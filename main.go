package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Lox struct {
	hadError bool
}

// main is the entry point of the program. It processes command line arguments
// and either runs a script file or starts an interactive prompt.
func main() {
	l := &Lox{}
	args := os.Args[1:]

	if len(args) > 1 {
		fmt.Println("Usage: go-lox [script]")
		os.Exit(64)
	} else if len(args) == 1 {
		if err := l.runFile(args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(65)
		}
	} else {
		l.runPrompt()
	}
}

// runFile reads the entire source code from a file specified by path,
// then runs it. Exits the program if any errors occurred during execution.
func (l *Lox) runFile(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	l.run(string(bytes))

	if l.hadError {
		os.Exit(65)
	}

	return nil
}
// runPrompt starts an interactive prompt (REPL) that reads user input line-by-line,
// executes the input, and resets the error flag after each line.
func (l *Lox) runPrompt() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			break // EOF or error exits loop
		}

		// Trim newline and any surrounding whitespace
		line = strings.TrimSpace(line)

		l.run(line)
		l.hadError = false
	}
}
// run takes source code as input, scans it into tokens,
// and currently prints the tokens to standard output.
func (l *Lox) run(source string) {
	scanner := NewScanner(source)
	tokens := scanner.ScanTokens()

	for _, token := range tokens {
		fmt.Println(token)
	}
}

// reportError reports an error on a specific line with a message.
// It delegates to the report helper method.
func (l *Lox) reportError(line int, message string) {
	l.report(line, "", message)
}

// report formats and prints an error message including the line number
// and error location, and sets the hadError flag to true.
func (l *Lox) report(line int, where, message string) {
	fmt.Fprintf(os.Stderr, "[line %d] Error%s: %s\n", line, where, message)
	l.hadError = true
}
