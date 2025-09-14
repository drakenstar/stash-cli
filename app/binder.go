package app

import (
	"github.com/drakenstar/stash-cli/command"
)

// binder is a small utility function for generating quick command.Command configurations for a specific type.
func binder[T any]() command.Command {
	return command.Command{
		Resolve: func(i command.Iterator) (any, error) {
			var t T
			err := command.Bind(i, &t)
			return t, err
		},
	}
}

// static is a small utility for implementing resolvers that return a static value.
func static(msg any) command.Command {
	return command.Command{
		Resolve: func(i command.Iterator) (any, error) {
			return msg, nil
		},
	}
}
