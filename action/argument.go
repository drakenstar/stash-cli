package action

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ArgumentValue struct {
	Name  string
	Value string
}

type ArgumentList []ArgumentValue

// Positional returns only positional arguments in a list.
func (l ArgumentList) Named() ArgumentList {
	var a ArgumentList
	for _, v := range l {
		if v.Name != "" {
			a = append(a, v)
		}
	}
	return a
}

// Positional returns only positional arguments in a list as a slice of stirngs.
func (l ArgumentList) Positional() []string {
	var a []string
	for _, v := range l {
		if v.Name == "" {
			a = append(a, v.Value)
		}
	}
	return a
}

// Binder is an interface that can be implemented by destinations passed to ArgumentList.Bind.  This allows destinations
// to define their own Bind method separately from the default matching.
type Binder interface {
	Bind(ArgumentList) error
}

var (
	ErrNonPointerStruct   = errors.New("bind destination must be a non-nil pointer to a struct value")
	ErrUnusedArgument     = errors.New("not all arguments were consumed")
	ErrNonPointerReceiver = errors.New("Binder interface must be implemented on a pointer receiver")
)

// Bind takes a pointer destination value, and populates it from the argument list.
//
// By default the destination must be a struct and will have it's fields populated from the argument list.  Fields can
// be populated by name or by position.  Names are taken from an "action" struct tag, and fallback to the lowercased
// field name.  If a named argument is not found, then a positional argument will be matched instead.  Named argument
// matching can be skipped by passing a dash as the action tag, e.g. `action:"-"`
//
// Argument lists are expected to match up with provided structs, so additional arguments will result in an error.
// Unmatched fields in the target struct will not result in an error, as they may be optional fields.
//
// Destination values may also implement the Binder interface to determine their own binding logic.
func (l ArgumentList) Bind(dest any) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return ErrNonPointerStruct
	}

	if binder, ok := dest.(Binder); ok {
		if reflect.TypeOf(dest).Kind() != reflect.Ptr {
			return ErrNonPointerReceiver
		}
		return binder.Bind(l)
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
		// Attempt to match a name argument to this field and set it.  Fields can opt out of name matching with `action:"-"`
		name := name(f)
		if name != "" {
			if arg, ok := named[name]; ok {
				arg.bound = true
				if err := set(v.Field(i), arg.Value); err != nil {
					return err
				}
				continue
			}
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
	if len(positional) > 0 || unused(named) {
		return ErrUnusedArgument
	}

	return nil
}

func name(f reflect.StructField) string {
	name := f.Tag.Get("action")
	// A single - indicates that this field should not be matched by name, i.e. positional only
	if name == "-" {
		return ""
	}
	if name == "" {
		name = strings.ToLower(f.Name)
	}
	return name
}

// set attempts to parse a string into the provided reflect.Value. It supports a few different types, as well as
// pointers to those types.  Error will be returned whenever an unsupported type is encountered.
func set(f reflect.Value, s string) error {
	// If we have a pointer, then set it so a zero value if it's nil, and then call set on it's value.
	if f.Kind() == reflect.Pointer {
		if f.IsNil() {
			f.Set(reflect.New(f.Type().Elem()))
		}
		return set(f.Elem(), s)
	}

	switch f.Kind() {
	case reflect.String:
		f.SetString(s)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("failed to parse '%s' as integer: %v", s, err)
		}
		f.SetInt(int64(intValue))
		return nil

	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("failed to parse '%s' as float: %v", s, err)
		}
		f.SetFloat(floatValue)
		return nil

	case reflect.Bool:
		boolValue, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("failed to parse '%s' as bool: %v", s, err)
		}
		f.SetBool(boolValue)
		return nil

	case reflect.Struct:
		if f.Type() == reflect.TypeOf(time.Time{}) {
			dateValue, err := time.Parse("2006-01-02", s)
			if err != nil {
				return fmt.Errorf("failed to parse '%s' as date: %v", s, err)
			}
			f.Set(reflect.ValueOf(dateValue))
			return nil
		}
	}

	return fmt.Errorf("unsupported type: %v", f.Kind())
}

func unused(l map[string]*struct {
	ArgumentValue
	bound bool
}) bool {
	for _, v := range l {
		if !v.bound {
			return true
		}
	}
	return false
}
