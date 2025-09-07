package actions

import (
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

// New returns a pointer to a new instance of Action.  Input is provided as a string value.  io.Reader is not used
// since inputs are expected to be small single lines of text.
func New(input string) *Action {
	return &Action{input: input}
}

type TokenType int

type Token struct {
	Type  TokenType
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
	next := strings.IndexFunc(a.input[a.pos:], unicode.IsSpace)
	var value string
	if next == -1 {
		value = a.input[a.pos:]
		a.pos = len(a.input)
	} else {
		value = a.input[a.pos : a.pos+next]
		a.pos = a.pos + next
	}
	return Token{
		Label: value,
	}, nil
}

func (*Action) Bind(dest any) error {
	return nil
}
