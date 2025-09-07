package actions

import (
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
			"command \ttest bar\t\"quoted\\\" input\t\" foo",
			[]any{
				Token{Label: "command", Value: "command"},
				Token{Label: "test", Value: "test"},
				Token{Label: "bar", Value: "bar"},
				Token{Label: "quoted\" input\t", Value: "quoted\" input\t"},
				Token{Label: "foo", Value: "foo"},
				io.EOF,
			},
		},
		{
			"quoted edges",
			`"\"foo" "bar\""`,
			[]any{
				Token{Label: `"foo`, Value: `"foo`},
				Token{Label: `bar"`, Value: `bar"`},
				io.EOF,
			},
		},
		{
			"unterminated quote",
			`foo "bar`,
			[]any{
				Token{Label: `foo`, Value: `foo`},
				ErrorUnterminatedQuote,
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
