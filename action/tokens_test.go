package action

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input  string
		tokens []Token
		error  error
	}{
		{`foo bar 0b_-`, []Token{{TOKEN_IDENTIFIER, `foo`, 0}, {TOKEN_WHITESPACE, ` `, 3}, {TOKEN_IDENTIFIER, `bar`, 4}, {TOKEN_WHITESPACE, ` `, 7}, {TOKEN_STRING, `0b_-`, 8}}, nil},
		{`foo bar=baz`, []Token{{TOKEN_IDENTIFIER, `foo`, 0}, {TOKEN_WHITESPACE, ` `, 3}, {TOKEN_IDENTIFIER, `bar`, 4}, {TOKEN_ARGUMENT_SEPARATOR, `=`, 7}, {TOKEN_IDENTIFIER, `baz`, 8}}, nil},
		{`foo "bar =baz" 'bim bam'`, []Token{{TOKEN_IDENTIFIER, `foo`, 0}, {TOKEN_WHITESPACE, ` `, 3}, {TOKEN_QUOTED_STRING, `"bar =baz"`, 4}, {TOKEN_WHITESPACE, ` `, 14}, {TOKEN_QUOTED_STRING, `'bim bam'`, 15}}, nil},
		{` 0oo`, []Token{{TOKEN_WHITESPACE, ` `, 0}, {TOKEN_STRING, `0oo`, 1}}, nil},
		{`"0oo`, nil, fmt.Errorf("unterminated string literal '\"' as pos 4")},
		{` "0oo\"`, nil, fmt.Errorf("unterminated string literal '\"' as pos 7")},
		{`"foo \" bar"`, []Token{{TOKEN_QUOTED_STRING, `"foo \" bar"`, 0}}, nil},
		{`🦁`, []Token{{TOKEN_STRING, `🦁`, 0}}, nil},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			tokens, err := Tokenize(test.input)
			if test.error == nil {
				require.NoError(t, err)
				require.Equal(t, test.tokens, tokens)
			} else {
				require.Equal(t, test.error, err)
			}
		})
	}
}

func TestTokenString(t *testing.T) {
	tests := []struct {
		token    Token
		expected string
	}{
		{Token{TOKEN_STRING, `foo`, 0}, `foo`},
		{Token{TOKEN_WHITESPACE, " \r\n\t", 0}, " \r\n\t"},
		{Token{TOKEN_ARGUMENT_SEPARATOR, `=`, 0}, `=`},
		{Token{TOKEN_QUOTED_STRING, `"foo"`, 0}, `foo`},
		{Token{TOKEN_QUOTED_STRING, `"foo \" bar"`, 0}, `foo " bar`},
	}
	for _, test := range tests {
		t.Run(test.token.Literal, func(t *testing.T) {
			require.Equal(t, test.expected, test.token.String())
		})
	}
}
