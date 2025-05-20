package main

import "strconv"

// keywords maps reserved keyword strings to their corresponding TokenType.
// It is used to distinguish identifiers from language keywords during scanning.
var keywords = map[string]TokenType{
	"and":    AND,
	"class":  CLASS,
	"else":   ELSE,
	"false":  FALSE,
	"for":    FOR,
	"fun":    FUN,
	"if":     IF,
	"nil":    NIL,
	"or":     OR,
	"print":  PRINT,
	"return": RETURN,
	"super":  SUPER,
	"this":   THIS,
	"true":   TRUE,
	"var":    VAR,
	"while":  WHILE,
}

// Scanner performs lexical analysis on the input source code.
type Scanner struct {
	source  []rune   // source code as a slice of runes (Unicode characters)
	tokens  []*Token // the list of scanned tokens
	start   int      // start index of the current lexeme
	current int      // current index in the source
	line    int      // current line number
	lox     *Lox     // reference to the Lox instance for error reporting
}

// NewScanner constructs and returns a new Scanner for the given source code.
func NewScanner(source string, lox *Lox) *Scanner {
	return &Scanner{
		source: []rune(source),
		tokens: []*Token{}, // can probably be omitted but whatever
		line:   1,
		lox:    lox,
	}
}

// ScanTokens scans through the entire source code, producing tokens until the end is reached.
// It returns a slice of pointers to Tokens, including a final EOF token.
func (s *Scanner) ScanTokens() []*Token {
	for !s.isAtEnd() {
		// At the start of a new lexeme, set start index to current position
		s.start = s.current
		s.scanToken()
	}

	// Append a special EOF token signaling end of source
	s.tokens = append(s.tokens, &Token{
		TokenType: EOF,
		Lexeme:    "",
		Literal:   nil,
		Line:      s.line,
	})

	return s.tokens
}

// scanToken consumes the next character from the source,
// determines its token type (including handling two-character tokens),
// and adds the corresponding token to the token list.
func (s *Scanner) scanToken() {
	var c rune = s.advance()
	switch c {
	case '(':
		s.addToken(LEFT_PAREN)
	case ')':
		s.addToken(RIGHT_PAREN)
	case '{':
		s.addToken(LEFT_BRACE)
	case '}':
		s.addToken(RIGHT_BRACE)
	case ',':
		s.addToken(COMMA)
	case '.':
		s.addToken(DOT)
	case '-':
		s.addToken(MINUS)
	case '+':
		s.addToken(PLUS)
	case ';':
		s.addToken(SEMICOLON)
	case '*':
		s.addToken(STAR)
	case '!':
		if s.match('=') {
			s.addToken(BANG_EQUAL)
		} else {
			s.addToken(BANG)
		}
	case '=':
		if s.match('=') {
			s.addToken(EQUAL_EQUAL)
		} else {
			s.addToken(EQUAL)
		}
	case '<':
		if s.match('=') {
			s.addToken(LESS_EQUAL)
		} else {
			s.addToken(LESS)
		}
	case '>':
		if s.match('=') {
			s.addToken(GREATER_EQUAL)
		} else {
			s.addToken(GREATER)
		}
	case '/':
		if s.match('/') {
            // Single-line comment
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
			// Don't add a token for the comment, just skip it
		} else if s.match('*') {
            // C-style block comment (/*...*/)
            s.blockComment()
        } else {
			s.addToken(SLASH)
		}
	case ' ', '\r', '\t':
		// Ignore whitespace. (Do nothing)
	case '\n':
		s.line++
	case '"':
		s.string()
	default:
		if s.isDigit(c) {
			s.number()
		} else if s.isAlpha(c) {
			s.identifier()
		} else {
			s.lox.reportError(s.line, ErrUnexpectedCharacter)
		}
	}
}

// identifier scans a sequence of letters/digits to form an identifier.
// If it matches a reserved keyword, the appropriate token is added.
func (s *Scanner) identifier() {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}

	// Check if it's a reserved keyword
	text := string(s.source[s.start:s.current])
	if tokenType, ok := keywords[text]; ok {
		s.addToken(tokenType)
	} else {
		s.addToken(IDENTIFIER)
	}
}

// number consumes a numeric literal, including an optional fractional part,
// then adds a NUMBER token with the parsed float64 value.
func (s *Scanner) number() {
	// Consume the integer part digits.
	for s.isDigit(s.peek()) {
		s.advance()
	}

	// Look for a fractional followed by at least one digit.
	if s.peek() == '.' && s.isDigit(s.peekNext()) {
		// Consume the ".".
		s.advance()

		// Consume the digits of the fractional part.
		for s.isDigit(s.peek()) {
			s.advance()
		}
	}

	// Convert the numeric substring to a float64 and add it as a token literal.
	text := string(s.source[s.start:s.current])
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		s.lox.reportError(s.line, ErrInvalidNumberLiteral)
		return
	}
	s.addTokenWithLiteral(NUMBER, value)
}

// string scans a string literal, consuming characters until the closing quote or EOF.
// It updates line count on newlines, reports an error if unterminated,
// then adds a STRING token with the extracted value (excluding quotes).
func (s *Scanner) string() {
	// Keep consuming characters until we find a closing quote or reach the end
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	// If we reached the end without finding a closing quote, report error
	if s.isAtEnd() {
		s.lox.reportError(s.line, ErrUnterminatedString)
		return
	}

	// Consume the closing "
	s.advance()

	// Extract the string value without the surrounding quotes
	// s.start is at opening ", s.current is after closing "
	value := string(s.source[s.start+1 : s.current-1])

	// Add the string token with literal value
	s.addTokenWithLiteral(STRING, value)
}

// match checks if the next character in the source matches the expected rune.
// If it matches, advances the current index and returns true.
// Otherwise, returns false without consuming input.
func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}
	if s.source[s.current] != expected {
		return false
	}

	s.current++
	return true
}

// peek returns the current character without consuming it.
// If at the end of the source, returns zero rune (0).
func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return 0 // zero rune, similar to '\0' in Java
	}

	return s.source[s.current]
}

// peekNext returns the next character after the current one without consuming it.
// Returns zero rune (0) if at or past the end of the source.
func (s *Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return 0 // zero rune, similar to '\0' in Java
	}

	return s.source[s.current+1]
}

// isAlpha returns true if the rune is an uppercase or lowercase letter, or underscore.
func (s *Scanner) isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

// isAlphaNumeric returns true if the rune is a letter, digit, or underscore.
func (s *Scanner) isAlphaNumeric(c rune) bool {
	return s.isAlpha(c) || s.isDigit(c)
}

// isDigit returns true if the rune c is a digit character
// ('0' through '9'), false otherwise.
func (s *Scanner) isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

// isAtEnd returns true if the lexer has consumed all characters in the source string.
func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

// advance consumes the next character in the source, returning it,
// and increments the current index.
func (s *Scanner) advance() rune {
	ch := s.source[s.current]
	s.current++
	return ch
}

// addToken creates a new token with the given TokenType and no literal value,
// then appends it to the tokens slice.
func (s *Scanner) addToken(tokenType TokenType) {
	s.addTokenWithLiteral(tokenType, nil)
}

// addTokenWithLiteral creates a new token with the given TokenType and optional
// literal value (e.g. for numbers, strings), then appends it to the tokens slice.
func (s *Scanner) addTokenWithLiteral(tokenType TokenType, literal any) {
	text := string(s.source[s.start:s.current])
	token := &Token{
		TokenType: tokenType,
		Lexeme:    text,
		Literal:   literal,
		Line:      s.line,
	}
	s.tokens = append(s.tokens, token)
}

// blockComment consumes characters until it finds the closing '*/' or EOF.
// It properly updates line numbers and reports an error for unterminated comments.
func (s *Scanner) blockComment() {
	for !s.isAtEnd() {
		if s.peek() == '*' && s.peekNext() == '/' {
			// End of block comment
			s.advance() // consume '*'
			s.advance() // consume '/'
			return
		}
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	// If we reach here, the comment wasn't closed
	s.lox.reportError(s.line, ErrUnterminatedBlockComment)
}

