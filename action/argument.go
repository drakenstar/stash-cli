package action

import (
	"errors"
	"reflect"
	"strings"
)

type ArgumentValue struct {
	Name  string
	Value string
}

type ArgumentList []ArgumentValue

var (
	ErrNonPointerStruct = errors.New("Bind destination must be a non-nil pointer to a struct value")
	ErrUnusedArgument   = errors.New("Not all arguments were consumed")
)

// Bind takes a pointer to a struct value, and populates it's fields from the argument list.  Fields can be populated
// by name or by position.  Names are taken from an "action" struct tag, and fallback to the lowercased field name.
// If a named argument is not found, then a positional argument will be matched instead.
//
// Argument lists are expected to match up with provided structs, so additional arguments will result in an error.
// Unmatched fields in the target struct will not result in an error, as they may be optional fields.
func (l ArgumentList) Bind(dest any) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return ErrNonPointerStruct
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return ErrNonPointerStruct
	}

	// Denormalize arguments into named and positional.  This will aid us tracking which arguments get consumed later.
	named := make(map[string]*struct {
		ArgumentValue
		bound bool
	})
	positional := []ArgumentValue{}
	for i := 0; i < len(l); i++ {
		a := l[i]
		if a.Name == "" {
			positional = append(positional, a)
		} else {
			named[a.Name] = &struct {
				ArgumentValue
				bound bool
			}{a, false}
		}
	}

	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		// Attempt to match a name argument to this field and set it.
		name := name(f)
		if arg, ok := named[name]; ok {
			arg.bound = true
			if err := set(v.Field(i), arg.Value); err != nil {
				return err
			}
			continue
		}

		// Otherwise consume a positional argument and attempt to set it.
		if len(positional) > 0 {
			arg := positional[0]
			positional = positional[1:]
			if err := set(v.Field(i), arg.Value); err != nil {
				return err
			}
		}

		// No match could be found in the arguments for this field, so it remains unchanged.
	}

	// Check whether we have any additional arguments.
	if len(positional) > 0 {
		return ErrUnusedArgument
	}

	return nil
}

func name(f reflect.StructField) string {
	name := f.Tag.Get("action")
	if name == "" {
		name = strings.ToLower(f.Name)
	}
	return name
}

func set(f reflect.Value, s string) error {
	v := reflect.ValueOf(s)
	f.Set(v)
	return nil
}
