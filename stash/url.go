package stash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// http://mendev.local:9999/scenes?c=(%22type%22:%22organized%22,%22value%22:%22true%22,%22modifier%22:%22EQUALS%22)&sortby=updated_at&sortdir=desc

type Route struct {
	Path          string
	FindFilter    FindFilter
	SceneFilter   *SceneFilter
	GalleryFilter *GalleryFilter
}

type filterQueryTuple struct {
	T        string            `json:"type"`
	Modifier CriterionModifier `json:"modifier"`
	Value    any               `json:"value"`
}

func (t filterQueryTuple) dateCriterion(format string) (*DateCriterion, error) {
	v, ok := t.Value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("date filter must have map[string]any value")
	}
	c := &DateCriterion{
		Modifier: t.Modifier,
	}

	v1, ok := v["value"].(string)
	if !ok {
		return nil, fmt.Errorf("date filter value must have a 'value' key")
	}
	v1t, err := time.Parse(format, v1)
	if err != nil {
		return nil, err
	}
	c.Value = v1t

	if v2, ok := v["value2"].(string); ok {
		v2t, err := time.Parse(format, v2)
		if err != nil {
			return nil, err
		}
		c.Value2 = &v2t
	}

	return c, nil
}

func (t filterQueryTuple) DateCriterion() (*DateCriterion, error) {
	return t.dateCriterion("2006-01-02")
}

func (t filterQueryTuple) TimestampCriterion() (*TimestampCriterion, error) {
	c, err := t.dateCriterion("2006-01-02 03:04")
	return (*TimestampCriterion)(c), err
}

func (t filterQueryTuple) BoolPtr() (*bool, error) {
	v, ok := t.Value.(string)
	if !ok {
		return nil, fmt.Errorf("boolean filter must have string value")
	}
	if v != "true" && v != "false" {
		return nil, fmt.Errorf("boolean filter must have value 'true' or 'false'")
	}
	b := v == "true"
	return &b, nil
}

func (t filterQueryTuple) StringCriterion() (*StringCriterion, error) {
	v, ok := t.Value.(string)
	if !ok {
		return nil, fmt.Errorf("string filter must have string value")
	}
	return &StringCriterion{
		Value:    v,
		Modifier: t.Modifier,
	}, nil
}

func (t filterQueryTuple) IntCriterion() (*IntCriterion, error) {
	v, ok := t.Value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("date filter must have map[string]any value")
	}

	c := &IntCriterion{
		Modifier: t.Modifier,
	}

	v1, ok := v["value"].(float64)
	if !ok {
		return nil, fmt.Errorf("date filter value must have an int 'value' key %T", v["value"])
	}
	c.Value = int(v1)

	if v2, ok := v["value2"].(float64); ok {
		v2i := int(v2)
		c.Value2 = &v2i
	}

	return c, nil
}

func ParseUrl(u *url.URL) (Route, error) {
	r := Route{
		Path: u.Path,
	}
	switch r.Path {
	case "/scenes":
		findFilter, err := findFilter(u.Query())
		if err != nil {
			return Route{}, err
		}
		r.FindFilter = findFilter

		r.SceneFilter = &SceneFilter{}

		filters := make(map[string]*filterQueryTuple, len(u.Query()["c"]))
		for _, c := range u.Query()["c"] {
			t := new(filterQueryTuple)
			err := json.Unmarshal(decodeJSON(c), &t)
			if err != nil {
				return Route{}, err
			}
			filters[t.T] = t
		}
		if err := unmarshalFilters(filters, &r.SceneFilter); err != nil {
			return Route{}, err
		}

	case "/galleries":
		findFilter, err := findFilter(u.Query())
		if err != nil {
			return Route{}, err
		}
		r.FindFilter = findFilter

	default:
		return Route{}, fmt.Errorf("unsupported URL route %s", r.Path)
	}

	return r, nil
}

func findFilter(params url.Values) (FindFilter, error) {
	f := FindFilter{
		Query:     params.Get("q"),
		Sort:      params.Get("sortby"),
		Direction: SortDirectionAsc,
		Page:      1,
		PerPage:   40,
	}

	if params.Has("sortdir") {
		f.Direction = strings.ToUpper(params.Get("sortdir"))
		if f.Direction != SortDirectionAsc && f.Direction != SortDirectionDesc {
			return FindFilter{}, fmt.Errorf("invalid sort direction %s", params.Get("sortdir"))
		}
	}

	if params.Has("p") {
		p, err := strconv.Atoi(params.Get("p"))
		if err != nil {
			return FindFilter{}, err
		}
		f.Page = p
	}

	if params.Has("perPage") {
		perPage, err := strconv.Atoi(params.Get("perPage"))
		if err != nil {
			return FindFilter{}, err
		}
		f.PerPage = perPage
	}

	return f, nil
}

func unmarshalFilters(filters map[string]*filterQueryTuple, filter any) error {
	v := reflect.ValueOf(filter).Elem().Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		fieldKey := strings.Split(fieldType.Tag.Get("json"), ",")[0] // ignore omitempty

		if fieldKey != "" && filters[fieldKey] != nil {
			var f any
			var err error
			switch field.Interface().(type) {
			case *bool:
				f, err = filters[fieldKey].BoolPtr()
			case *StringCriterion:
				f, err = filters[fieldKey].StringCriterion()
			case *DateCriterion:
				f, err = filters[fieldKey].DateCriterion()
			case *TimestampCriterion:
				f, err = filters[fieldKey].TimestampCriterion()
			case *IntCriterion:
				f, err = filters[fieldKey].IntCriterion()
			default:
				err = fmt.Errorf("unsupported field type %T", fieldKey)
			}

			if err != nil {
				return fmt.Errorf("error on field %s: %w", fieldKey, err)
			}
			field.Set(reflect.ValueOf(f))
		}
	}

	return nil
}

func decodeJSON(jsonString string) []byte {
	var b bytes.Buffer
	inString := false
	escape := false

	for _, c := range jsonString {
		if escape {
			// this character has been escaped, skip
			escape = false
			b.WriteRune(c)
			continue
		}

		switch c {
		case '\\':
			// escape the next character if in a string
			if inString {
				escape = true
			}
		case '"':
			// unescaped quote, toggle inString
			inString = !inString
		case '(':
			// restore ( to { if not in a string
			if !inString {
				c = '{'
			}
		case ')':
			// restore ) to } if not in a string
			if !inString {
				c = '}'
			}
		}

		if !escape {
			b.WriteRune(c)
		}
	}

	return b.Bytes()
}
