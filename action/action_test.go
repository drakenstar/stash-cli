package action

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input  string
		action *Action
		error  error
	}{
		{
			`command arg1 arg2=foo arg3="bar baz"`,
			&Action{
				Name: "command",
				Arguments: []ArgumentValue{
					{Name: "", Value: "arg1"}, {Name: "arg2", Value: "foo"}, {Name: "arg3", Value: "bar baz"},
				},
			},
			nil,
		},
		{
			`command arg1=`,
			nil,
			ErrIncompleteArgument,
		},
		{
			`command foo = bar`,
			nil,
			errors.New("invalid token found as pos 12: TOKEN_ARGUMENT_SEPARATOR(=), expected: TOKEN_IDENTIFIER"),
		},
		{
			`command`,
			&Action{
				Name: "command",
			},
			nil,
		},
		{
			`command "= value"`,
			&Action{
				Name: "command",
				Arguments: []ArgumentValue{
					{
						Name:  "",
						Value: "= value",
					},
				},
			},
			nil,
		},
		{
			`command "foo bar"=baz`,
			nil,
			errors.New("invalid token found as pos 8: TOKEN_QUOTED_STRING(foo bar), expected: TOKEN_IDENTIFIER"),
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			a, err := Parse(test.input)
			if test.error == nil {
				require.NoError(t, err)
				require.Equal(t, test.action, a)
			} else {
				require.Equal(t, test.error, err)
			}
		})
	}
}
