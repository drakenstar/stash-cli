package action

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testDest struct {
	Foo    string
	Bar    *string `action:"baz"`
	Qux    int
	Quux   time.Time
	Quzz   *bool
	Corge  *bool
	Grault bool
}

type binderDest string

func (d *binderDest) Bind(l ArgumentList) error {
	if len(l) == 0 {
		return errors.New("no arguments provided")
	}
	*d = binderDest(l[0].Value)
	return nil
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
				{Name: "corge", Value: "true"},
				{Name: "", Value: "1234"},
				{Name: "", Value: "2025-07-27"},
				{Name: "grault", Value: "0"},
			},
			ptr(testDest{}),
			ptr(testDest{
				Foo:    "bar",
				Bar:    ptr("baz"),
				Qux:    1234,
				Quux:   time.Date(2025, 07, 27, 0, 0, 0, 0, time.UTC),
				Quzz:   nil,
				Corge:  ptr(true),
				Grault: false,
			}),
			nil,
		},
		{
			"should error when positional arguments are not consumed",
			ArgumentList{{Value: "boo"}},
			&struct{}{},
			nil,
			ErrUnusedArgument,
		},
		{
			"should error when named arguments are not consumed",
			ArgumentList{{Name: "boo"}},
			&struct{}{},
			nil,
			ErrUnusedArgument,
		},
		{
			"should skip named arguments for fields with a -",
			ArgumentList{{Name: "foo", Value: "bar"}},
			&struct {
				Foo string `action:"-"`
			}{},
			nil,
			ErrUnusedArgument,
		},
		{
			"should not skip positional arguments for fields with a -",
			ArgumentList{{Value: "bar"}},
			&struct {
				Foo string `action:"-"`
			}{},
			&struct {
				Foo string `action:"-"`
			}{
				Foo: "bar",
			},
			nil,
		},
		{
			"should match Binder implementors and call their logic",
			ArgumentList{{Value: "foo"}},
			ptr(binderDest("")),
			ptr(binderDest("foo")),
			nil,
		},
		{
			"should return errors from Binder implementors",
			ArgumentList{},
			ptr(binderDest("")),
			nil,
			errors.New("no arguments provided"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.args.Bind(test.dest)
			if test.error == nil {
				require.NoError(t, err)
				require.Equal(t, test.expected, test.dest)
			} else {
				require.Equal(t, test.error, err)
			}
		})
	}
}
