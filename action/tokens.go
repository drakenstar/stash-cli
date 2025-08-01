package action

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type TokenType int

const (
	TOKEN_WHITESPACE         TokenType = iota // Matches any sort of whitespace encountered outside of another token type.
	TOKEN_STRING                              // A string without any internal whitespace.
	TOKEN_QUOTED_STRING                       // A quoted string, which may contain both whitespace and escaped quotes.
	TOKEN_ARGUMENT_SEPARATOR                  // Separator between an argument name, and a value
	TOKEN_IDENTIFIER                          // A command/argument name, which must be a letter followed by alphanumeric characters.
)

func (t TokenType) String() string {
	switch t {
	case TOKEN_WHITESPACE:
		return "TOKEN_WHITESPACE"
	case TOKEN_STRING:
		return "TOKEN_STRING"
	case TOKEN_QUOTED_STRING:
		return "TOKEN_QUOTED_STRING"
	case TOKEN_ARGUMENT_SEPARATOR:
		return "TOKEN_ARGUMENT_SEPARATOR"
	case TOKEN_IDENTIFIER:
		return "TOKEN_IDENTIFIER"
	default:
		return ""
	}
}

const ARGUMENT_SEPARATOR = '='

var identifierRegex = regexp.MustCompile(`^[a-zA-Z]\w+$`)

type Token struct {
	Type    TokenType
	Literal string
	Pos     int
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
			tokens = append(tokens, Token{TOKEN_WHITESPACE, input[start:pos], start})
			continue
		}

		// Named argument separators
		if input[pos] == ARGUMENT_SEPARATOR {
			tokens = append(tokens, Token{TOKEN_ARGUMENT_SEPARATOR, "=", pos})
			pos++
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
			tokens = append(tokens, Token{TOKEN_QUOTED_STRING, input[start:pos], start})
			continue
		}

		// Unquoted string
		start := pos
		pos++
		for pos < len(input) && !unicode.IsSpace(rune(input[pos])) && input[pos] != ARGUMENT_SEPARATOR {
			pos++
		}
		if identifierRegex.Match([]byte(input[start:pos])) {
			tokens = append(tokens, Token{TOKEN_IDENTIFIER, input[start:pos], start})
		} else {
			tokens = append(tokens, Token{TOKEN_STRING, input[start:pos], start})
		}
	}

	return tokens, nil
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
