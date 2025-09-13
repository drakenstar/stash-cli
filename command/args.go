package command

import (
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Argument is a unit of command input.  An argument always has a Value, but can optionally also have a Name prefixed
// by NameSeparator.  Argument inputs can be quoted if they contain spaces or name separators.
type Argument struct {
	Raw   string
	Name  string
	Value string
}

// IsName returns a boolean indicating if this is possibly a "name" argument, that is an argument that is defined
// without an explicit name, and is not a quoted value.
func (a Argument) IsName() bool {
	return a.Name == "" && len(a.Raw) > 0 && a.Raw[0] != '"' && a.Raw[0] != '\''
}

// Iterator is an interface that yields successive arguments for each call to Next.  It is assumed to be stateful,
// such that once exhausted there is no mechanism for rewinding arguments.  An io.EOF should be returned to indicate
// that no more arguments are available.
type Iterator interface {
	Next() (Argument, error)
}

const NameSeparator = '='

// parser is a stateful representation of the parsing of a line of command input.  This struct can have Next() called
// any number of times (input allowing), before a caller decides to Bind().  Any call after Bind() will error with
// ErrorEOF and the parser should be discarded.
type parser struct {
	input string

	pos int
}

// Parser returns a pointer to a new instance of Action.  Input is provided as a string value.  io.Reader is not used
// since inputs are expected to be small single lines of text.
func Parser(input string) *parser {
	return &parser{input: input}
}

// Next returns the next token
func (a *parser) Next() (Argument, error) {
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
		return Argument{}, io.EOF
	}

	rawStart := a.pos

	value, parsedName, err := a.parseValue()
	if err != nil {
		return Argument{}, err
	}

	var t Argument
	if parsedName {
		t.Name = value
		value, parsedName, err := a.parseValue()
		if parsedName {
			return Argument{}, fmt.Errorf("argument contains multiple name separators = as position %d", a.pos)
		}
		if err != nil {
			return Argument{}, err
		}
		t.Value = value
	} else {
		t.Value = value
	}

	t.Raw = a.input[rawStart:a.pos]

	return t, nil
}

func (a *parser) parseValue() (string, bool, error) {
	// Quoted string, either double or single quoted.
	if a.input[a.pos] == '"' || a.input[a.pos] == '\'' {
		value, err := a.parseQuotedString()
		if err != nil {
			return "", false, err
		}
		return unquote(value), false, nil
	}

	// Semantically it's unclear what an argument starting with a name separator would mean so exit early.
	if a.input[a.pos] == NameSeparator {
		return "", false, fmt.Errorf("arguments may not start with the name separator = at position %d", a.pos)
	}

	start := a.pos
	parsedName := false
	for a.pos < len(a.input) {
		r, size := utf8.DecodeRuneInString(a.input[a.pos:])
		if unicode.IsSpace(r) {
			break
		}
		if r == NameSeparator {
			parsedName = true
			break
		}
		a.pos += size
	}

	end := a.pos
	// If we found a name, we want to advance the position to the start of the value.
	if parsedName {
		a.pos += 1
	}

	return a.input[start:end], parsedName, nil
}

// parseQuotedString takes a string of input, that starts with a quote
func (a *parser) parseQuotedString() (string, error) {
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
		return "", fmt.Errorf("unterminated quote as position %d", a.pos)
	}

	a.pos += 1
	return a.input[start:a.pos], nil
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
