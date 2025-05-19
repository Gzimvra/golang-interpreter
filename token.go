package main

import "fmt"

// Token represents a lexical token with type, lexeme,
// literal value, and source line.
type Token struct {
	TokenType TokenType
	Lexeme    string
	Literal   any
	Line      int
}

// NewToken constructs and returns a new Token.
func NewToken(tokenType TokenType, lexeme string, literal any, line int) Token {
	return Token{
		TokenType: tokenType,
		Lexeme:    lexeme,
		Literal:   literal,
		Line:      line,
	}
}

// String returns a string representation of the token, similar to Java's toString()
func (t *Token) String() string {
	return fmt.Sprintf("%v %s %v", t.TokenType, t.Lexeme, t.Literal)
}
