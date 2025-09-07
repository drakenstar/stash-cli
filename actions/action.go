package actions

import (
	"errors"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// action package
//
// This package provides parsing and binding capabilities around a small command DSL.  The DSL is optimised for typing
// and succinctness.

// Action is a stateful representation of the parsing of a line of command input.  This struct can have Next() called
// any number of times (input allowing), before a caller decides to Bind().  Any call after Bind() will error with
// ErrorEOF and the Action should be discarded.
type Action struct {
	input string

	pos int
}

var (
	// TODO this is probably not going to be enough context on it's own, a position for the start of the quoted string
	// is probably a useful aid.
	ErrorUnterminatedQuote = errors.New("unterminated quote")
)

// New returns a pointer to a new instance of Action.  Input is provided as a string value.  io.Reader is not used
// since inputs are expected to be small single lines of text.
func New(input string) *Action {
	return &Action{input: input}
}

type Token struct {
	Label string
	Value string
}

func (a *Action) Next() (Token, error) {
	// Consume all unicode whitespace
	for a.pos < len(a.input) {
		r, size := utf8.DecodeRuneInString(a.input[a.pos:])
		if !unicode.IsSpace(r) {
			break
		}
		a.pos += size
	}

	// If we're reached the end of our input, we want to return an io.EOF to indicate there are no more tokens.
	if a.pos >= len(a.input) {
		return Token{}, io.EOF
	}

	var value string

	// Quoted string, either double or single quoted.
	if a.input[a.pos] == '"' || a.input[a.pos] == '\'' {
		quote := a.input[a.pos]
		start := a.pos
		a.pos += 1

		for a.pos < len(a.input) {
			if a.input[a.pos] == quote && a.input[a.pos-1] != '\\' {
				break
			}
			a.pos += 1
		}

		if a.pos >= len(a.input) {
			return Token{}, ErrorUnterminatedQuote
		}

		a.pos += 1
		value = unquote(a.input[start:a.pos])
	} else {
		next := strings.IndexFunc(a.input[a.pos:], unicode.IsSpace)
		if next == -1 {
			value = a.input[a.pos:]
			a.pos = len(a.input)
		} else {
			value = a.input[a.pos : a.pos+next]
			a.pos = a.pos + next
		}
	}

	return Token{
		Label: value,
		Value: value,
	}, nil
}

func (*Action) Bind(dest any) error {
	return nil
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
