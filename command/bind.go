package command

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Setter is an interface to allow fields in a destination struct to define their own logic for parsing strings into
// values.
type Setter interface {
	Set(string) error
}

// TagKey is the key of the struct tag used to configure binding.
const TagKey = "command"

var (
	// Runtime errors related to bind destination, likely to be logic errors.
	ErrNonPointerStruct           = errors.New("bind destination must be a non-nil pointer to a struct value")
	ErrUnsupportedDestinationKind = errors.New("unsupported destination kind")

	// Input errors, likely to be errors user can correct.
	ErrUnrecognisedArgument = errors.New("unrecognised argument")
	ErrInvalidValue         = errors.New("invalid value")
)

// Bind consumes all arguments from the given Iterator until Next returns io.EOF and applies them to dest, which
// must be a pointer to a struct.
//
// Each argument is matched to a struct field by name or position and used to mutate dest. Most scalar types and some
// well-known stdlib types are supported.  Pointer fields are initialized to a zero value before being set. Boolean
// fields may be specified by name alone to imply true. Slice fields accept repeated arguments and each value is
// appended.
//
// By default, arguments are matched by name to the lower-case field name.  If a custom name is required it can be
// defined by struct name, e.g. `actions="custom-name"`.  This style of matching cannot be opted out of.
//
// Positional matching can be enabled by adding ",positional" to the field tag. They are bound in the order fields
// appear in the struct; reorder fields to change positional order. If the final positional field is a slice, any extra
// arguments are appended to that slice.

func Bind(a Iterator, dst any) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return ErrNonPointerStruct
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return ErrNonPointerStruct
	}

	// Do a pass through the fields of the destination and resolve a set of mappings of names to reflect.Value, as well
	// as a set of positional arguments.
	t := v.Type()
	named := make(map[string]reflect.Value)
	var positional []reflect.Value
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// Skip private and anonymous fields.
		if f.PkgPath != "" || f.Anonymous {
			continue
		}

		n, p := parseArgDetails(f)
		fv := v.Field(i)
		named[n] = fv
		if p {
			positional = append(positional, fv)
		}
	}

	pos := 0 // Keep state on which positional argument we should write to.
	for {
		arg, err := a.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// This is a positional argument, so attempt to write to the next positional input.
		if arg.Name == "" {
			// Special case: if an argument value has the name of a boolean field in the destination, then we treat
			// this as an implicit arg=true.  This is a convenience to allow having actions like "open no-skip"
			if f, ok := named[arg.Value]; ok && isBoolOrBoolPtr(f) {
				set(f, "true")
				continue
			}

			if len(positional) <= pos {
				return fmt.Errorf("%w: '%s'", ErrUnrecognisedArgument, arg.Raw)
			}
			set(positional[pos], arg.Value)
			// Special case: if the final positional value is a slice, that means we can continue to append to it for
			// any additional arguments that we encounter.
			if positional[pos].Kind() != reflect.Slice {
				pos++
			}
			continue
		}

		// Otherwise we have a named value, so all we need to do here is write.
		f, ok := named[arg.Name]
		if !ok {
			return fmt.Errorf("%w: %s", ErrUnrecognisedArgument, arg.Name)
		}
		set(f, arg.Value)
	}
}

// parseArgDetails takes a struct for this package and returns a name for the field, and a boolean indicating if this
// argument supports positional inputs.  Default is non-positional and the lower-cased name.
func parseArgDetails(f reflect.StructField) (string, bool) {
	name := strings.ToLower(f.Name)
	positional := false

	tag := f.Tag.Get(TagKey)
	if tag == "" {
		return name, false
	}

	parts := strings.Split(tag, ",")
	if len(parts) > 2 {
		panic(fmt.Sprintf(`misconfigured struct tag %s, format "name[,positional]"`, TagKey))
	}
	if len(parts) == 2 {
		if parts[1] != "positional" {
			panic(fmt.Sprintf(`misconfigured struct tag %s, format "name[,positional]"`, TagKey))
		}
		positional = true
	}
	if len(parts) == 1 {
		name = parts[0]
	}

	return name, positional
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

	// Supports the ability to implement Setter on a destination value for custom parsing logic.  Will return an error
	// in cases where the interface is implemented on a value receiver.
	if f.CanAddr() {
		if ss, ok := f.Addr().Interface().(Setter); ok {
			return ss.Set(s)
		}
	}

	switch f.Kind() {
	case reflect.Slice:
		elemT := f.Type().Elem()
		elemV := reflect.New(elemT).Elem()
		if err := set(elemV, s); err != nil {
			return err
		}
		f.Set(reflect.Append(f, elemV))
		return nil

	case reflect.String:
		f.SetString(s)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("%w: failed to parse '%s' as integer: %w", ErrInvalidValue, s, err)
		}
		f.SetInt(int64(intValue))
		return nil

	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("%w: failed to parse '%s' as float: %w", ErrInvalidValue, s, err)
		}
		f.SetFloat(floatValue)
		return nil

	case reflect.Bool:
		boolValue, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("%w: failed to parse '%s' as bool: %w", ErrInvalidValue, s, err)
		}
		f.SetBool(boolValue)
		return nil

	case reflect.Struct:
		if f.Type() == reflect.TypeOf(time.Time{}) {
			dateValue, err := time.Parse("2006-01-02", s) // TODO currently only handles dates, not times as well.
			if err != nil {
				return fmt.Errorf("%w: failed to parse '%s' as date: %w", ErrInvalidValue, s, err)
			}
			f.Set(reflect.ValueOf(dateValue))
			return nil
		}
	}

	return fmt.Errorf("%w: %v", ErrUnsupportedDestinationKind, f.Kind())
}

func isBoolOrBoolPtr(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return true
	case reflect.Pointer:
		return v.Type().Elem().Kind() == reflect.Bool
	default:
		return false
	}
}
