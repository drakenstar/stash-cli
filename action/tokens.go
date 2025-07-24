package action

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TOKEN_WHITESPACE TokenType = iota
	TOKEN_STRING
	TOKEN_QUOTED_STRING
	TOKEN_ARGUMENT_SEPARATOR
)

type Token struct {
	Type    TokenType
	Literal string
}

func (t Token) String() string {
	if t.Type == TOKEN_QUOTED_STRING {
		return unquote(t.Literal)
	}
	return t.Literal
}

func Tokenize(input string) ([]Token, error) {
	var tokens []Token
	pos := 0

	for pos < len(input) {

		// Top level whitespace
		if unicode.IsSpace(rune(input[pos])) {
			start := pos
			pos++
			for pos < len(input) && unicode.IsSpace(rune(input[pos])) {
				pos++
			}
			tokens = append(tokens, Token{TOKEN_WHITESPACE, input[start:pos]})
			continue
		}

		// Unquoted string
		if isWordStart(input[pos]) {
			start := pos
			pos++
			for pos < len(input) && isWordPart(input[pos]) {
				pos++
			}
			tokens = append(tokens, Token{TOKEN_STRING, input[start:pos]})
			continue
		}

		// Named argument separators
		if input[pos] == '=' {
			pos++
			tokens = append(tokens, Token{TOKEN_ARGUMENT_SEPARATOR, "="})
			continue
		}

		// Quoted strings, can be either double or single
		if input[pos] == '"' || input[pos] == '\'' {
			quote := input[pos]
			start := pos
			pos++
			for pos < len(input) {
				if input[pos] == quote && input[pos-1] != '\\' {
					break
				}
				pos++
			}
			if pos == len(input) {
				return nil, fmt.Errorf("unterminated string literal '%c' as pos %d", quote, pos)
			}
			pos++
			tokens = append(tokens, Token{TOKEN_QUOTED_STRING, input[start:pos]})
			continue
		}

		return nil, fmt.Errorf("unexpected character at pos %d: '%c'", pos, input[pos])
	}

	return tokens, nil
}

func isWordStart(c byte) bool {
	return unicode.IsLetter(rune(c))
}

func isWordPart(c byte) bool {
	return unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) || c == '_' || c == '-'
}

// unquote removes quotes and escape characters from a string found during tokenization. Assumes that the string is a
// valid quoted string, so no errors will be returned.
func unquote(in string) string {
	quote := in[0]
	var sb strings.Builder
	for i := 1; i < len(in)-1; i++ {
		if in[i] == '\\' && i+1 < len(in) && (in[i+1] == quote || in[i+1] == '\'') {
			sb.WriteByte(in[i+1])
			i++
		} else {
			sb.WriteByte(in[i])
		}
	}
	return sb.String()
}
