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
		{`foo bar b0_-`, []Token{{TOKEN_STRING, `foo`}, {TOKEN_WHITESPACE, ` `}, {TOKEN_STRING, `bar`}, {TOKEN_WHITESPACE, ` `}, {TOKEN_STRING, `b0_-`}}, nil},
		{`foo bar=baz`, []Token{{TOKEN_STRING, `foo`}, {TOKEN_WHITESPACE, ` `}, {TOKEN_STRING, `bar`}, {TOKEN_ARGUMENT_SEPARATOR, `=`}, {TOKEN_STRING, `baz`}}, nil},
		{`foo "bar =baz" 'bim bam'`, []Token{{TOKEN_STRING, `foo`}, {TOKEN_WHITESPACE, ` `}, {TOKEN_QUOTED_STRING, `"bar =baz"`}, {TOKEN_WHITESPACE, ` `}, {TOKEN_QUOTED_STRING, `'bim bam'`}}, nil},
		{` 0oo`, nil, fmt.Errorf("unexpected character at pos 1: '0'")},
		{` "0oo`, nil, fmt.Errorf("unterminated string literal '\"' as pos 5")},
		{` "0oo\"`, nil, fmt.Errorf("unterminated string literal '\"' as pos 7")},
		{`"foo \" bar"`, []Token{{TOKEN_QUOTED_STRING, `"foo \" bar"`}}, nil},
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
		{Token{TOKEN_STRING, `foo`}, `foo`},
		{Token{TOKEN_WHITESPACE, " \r\n\t"}, " \r\n\t"},
		{Token{TOKEN_ARGUMENT_SEPARATOR, `=`}, `=`},
		{Token{TOKEN_QUOTED_STRING, `"foo"`}, `foo`},
		{Token{TOKEN_QUOTED_STRING, `"foo \" bar"`}, `foo " bar`},
	}
	for _, test := range tests {
		t.Run(test.token.Literal, func(t *testing.T) {
			require.Equal(t, test.expected, test.token.String())
		})
	}
}
