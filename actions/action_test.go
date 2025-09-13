package actions

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
				Argument{Raw: `command`, Label: "", Value: "command"},
				Argument{Raw: `test`, Label: "", Value: "test"},
				Argument{Raw: `bar`, Label: "", Value: "bar"},
				Argument{Raw: `"quoted\" input	"`, Label: "", Value: "quoted\" input\t"},
				Argument{Raw: `foo=bar`, Label: "foo", Value: "bar"},
				Argument{Raw: `foo="quoted bar"`, Label: "foo", Value: "quoted bar"},
				io.EOF,
			},
		},
		{
			"quoted edges",
			`"\"foo" "bar\""`,
			[]any{
				Argument{Raw: `"\"foo"`, Label: "", Value: `"foo`},
				Argument{Raw: `"bar\""`, Label: "", Value: `bar"`},
				io.EOF,
			},
		},
		{
			"unterminated quote",
			`foo "bar`,
			[]any{
				Argument{Raw: `foo`, Label: "", Value: `foo`},
				errors.New("unterminated quote as position 8"),
			},
		},
		{
			"separator in value",
			`foo="bar=" foo=bar=`,
			[]any{
				Argument{Raw: `foo="bar="`, Label: "foo", Value: `bar=`},
				errors.New("argument contains multiple label separators = as position 19"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			a := New(test.input)
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
