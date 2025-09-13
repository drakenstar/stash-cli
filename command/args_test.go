package command

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAction(t *testing.T) {
	tests := []struct {
		title    string
		input    string
		expected []any
	}{
		{
			"simple",
			"command \ttest bar\t\"quoted\\\" input\t\" foo=bar foo=\"quoted bar\"",
			[]any{
				Argument{Raw: `command`, Name: "", Value: "command"},
				Argument{Raw: `test`, Name: "", Value: "test"},
				Argument{Raw: `bar`, Name: "", Value: "bar"},
				Argument{Raw: `"quoted\" input	"`, Name: "", Value: "quoted\" input\t"},
				Argument{Raw: `foo=bar`, Name: "foo", Value: "bar"},
				Argument{Raw: `foo="quoted bar"`, Name: "foo", Value: "quoted bar"},
				io.EOF,
			},
		},
		{
			"quoted edges",
			`"\"foo" "bar\""`,
			[]any{
				Argument{Raw: `"\"foo"`, Name: "", Value: `"foo`},
				Argument{Raw: `"bar\""`, Name: "", Value: `bar"`},
				io.EOF,
			},
		},
		{
			"unterminated quote",
			`foo "bar`,
			[]any{
				Argument{Raw: `foo`, Name: "", Value: `foo`},
				errors.New("unterminated quote as position 8"),
			},
		},
		{
			"separator in value",
			`foo="bar=" foo=bar=`,
			[]any{
				Argument{Raw: `foo="bar="`, Name: "foo", Value: `bar=`},
				errors.New("argument contains multiple name separators = as position 19"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			a := Parser(test.input)
			for _, next := range test.expected {
				arg, err := a.Next()
				if _, ok := next.(Argument); ok {
					require.NoError(t, err)
					require.Equal(t, next, arg)
				} else {
					require.Equal(t, next, err)
				}
			}
		})
	}
}
