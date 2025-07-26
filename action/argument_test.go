package action

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testDest struct {
	Foo string
	Bar string `action:"baz"`
	Qux string
}

func ptr[T any](t T) *T {
	return &t
}

func TestArgumentList(t *testing.T) {
	tests := []struct {
		name     string
		args     ArgumentList
		dest     any
		expected any
		error    error
	}{
		{"should error non-pointer", ArgumentList{}, testDest{}, nil, ErrNonPointerStruct},
		{"should error on nil", ArgumentList{}, nil, nil, ErrNonPointerStruct},
		{"should error on non-struct", ArgumentList{}, ptr("value"), nil, ErrNonPointerStruct},
		{
			"should bind simple types",
			ArgumentList{
				{Name: "foo", Value: "bar"},
				{Name: "baz", Value: "baz"},
				{Name: "", Value: "qux"},
			},
			ptr(testDest{}),
			ptr(testDest{
				Foo: "bar",
				Bar: "baz",
				Qux: "qux",
			}),
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.args.Bind(test.dest)
			fmt.Printf("%#v", test.dest)
			if test.error == nil {
				require.NoError(t, err)
				require.Equal(t, test.expected, test.dest)
			} else {
				require.Equal(t, test.error, err)
			}
		})
	}
}
