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
				Token{Raw: `command`, Label: "", Value: "command"},
				Token{Raw: `test`, Label: "", Value: "test"},
				Token{Raw: `bar`, Label: "", Value: "bar"},
				Token{Raw: `"quoted\" input	"`, Label: "", Value: "quoted\" input\t"},
				Token{Raw: `foo=bar`, Label: "foo", Value: "bar"},
				Token{Raw: `foo="quoted bar"`, Label: "foo", Value: "quoted bar"},
				io.EOF,
			},
		},
		{
			"quoted edges",
			`"\"foo" "bar\""`,
			[]any{
				Token{Raw: `"\"foo"`, Label: "", Value: `"foo`},
				Token{Raw: `"bar\""`, Label: "", Value: `bar"`},
				io.EOF,
			},
		},
		{
			"unterminated quote",
			`foo "bar`,
			[]any{
				Token{Raw: `foo`, Label: "", Value: `foo`},
				errors.New("unterminated quote as position 8"),
			},
		},
		{
			"separator in value",
			`foo=bar=`,
			[]any{
				errors.New("argument contains multiple label separators = as position 8"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			a := New(test.input)
			for _, next := range test.expected {
				tok, err := a.Next()
				if _, ok := next.(Token); ok {
					require.NoError(t, err)
					require.Equal(t, next, tok)
				} else {
					require.Equal(t, next, err)
				}
			}
		})
	}
}
