package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/drakenstar/stash-cli/stash"
)

type dateFilterValue struct {
	Modifier stash.CriterionModifier
	Value    time.Time
}

func (v *dateFilterValue) Set(s string) error {
	parsed, err := parseDateFilterValue(s)
	if err != nil {
		return err
	}
	*v = parsed
	return nil
}

func (v dateFilterValue) DateCriterion() *stash.DateCriterion {
	return &stash.DateCriterion{
		Modifier: v.Modifier,
		Value:    v.Value,
	}
}

func (v dateFilterValue) TimestampCriterion() *stash.TimestampCriterion {
	return &stash.TimestampCriterion{
		Modifier: v.Modifier,
		Value:    v.Value,
	}
}

func parseDateFilterValue(s string) (dateFilterValue, error) {
	modifier := stash.CriterionModifierEquals
	switch {
	case strings.HasPrefix(s, ">"):
		modifier = stash.CriterionModifierGreaterThan
		s = s[1:]
	case strings.HasPrefix(s, "<"):
		modifier = stash.CriterionModifierLessThan
		s = s[1:]
	}

	if s == "" {
		return dateFilterValue{}, fmt.Errorf("date filter value is empty")
	}

	value, err := parseDateValue(s)
	if err != nil {
		return dateFilterValue{}, err
	}

	return dateFilterValue{
		Modifier: modifier,
		Value:    value,
	}, nil
}

func parseDateValue(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	if len(s) < 2 {
		return time.Time{}, fmt.Errorf("invalid date filter %q", s)
	}

	unit := s[len(s)-1]
	magnitude, err := strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date filter %q", s)
	}

	now := time.Now().UTC()
	switch unit {
	case 'h':
		return now.Add(time.Duration(magnitude) * time.Hour), nil
	case 'd':
		return now.AddDate(0, 0, magnitude), nil
	case 'w':
		return now.AddDate(0, 0, magnitude*7), nil
	case 'm':
		return now.AddDate(0, magnitude, 0), nil
	case 'y':
		return now.AddDate(magnitude, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported date filter unit %q", string(unit))
	}
}
