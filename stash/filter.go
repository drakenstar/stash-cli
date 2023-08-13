package stash

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
)

type FindFilter struct {
	Query     string `json:"q"`
	Page      int    `json:"page"`
	PerPage   int    `json:"per_page"`
	Sort      string `json:"sort"`
	Direction string `json:"direction"`
}

func (FindFilter) GetGraphQLType() string {
	return "FindFilterType"
}

const (
	SortDate         = "date"
	SortUpdatedAt    = "updated_at"
	SortCreatedAt    = "created_at"
	SortPath         = "path"
	SortRandomPrefix = "random_"

	SortDirectionAsc  = "ASC"
	SortDirectionDesc = "DESC"
)

func RandomSort() string {
	return fmt.Sprintf("%s%08d", SortRandomPrefix, rand.Intn(100000000))
}

type FilterCombinator[T SceneFilter | GalleryFilter] struct {
	AND *T `json:"AND,omitempty"`
	OR  *T `json:"OR,omitempty"`
	NOT *T `json:"NOT,omitempty"`
}

type SceneFilter struct {
	FilterCombinator[SceneFilter]
	Organized  *bool           `json:"organized,omitempty"`
	Performers *MultiCriterion `json:"performers,omitempty"`
}

func (SceneFilter) GetGraphQLType() string {
	return "SceneFilterType"
}

type GalleryFilter struct {
	FilterCombinator[GalleryFilter]
	Organized  *bool               `json:"organized,omitempty"`
	Performers *MultiCriterion     `json:"performers,omitempty"`
	Path       *StringCriterion    `json:"path,omitempty"`
	CreatedAt  *TimestampCriterion `json:"created_at,omitempty"`
}

func (GalleryFilter) GetGraphQLType() string {
	return "GalleryFilterType"
}

type MultiCriterion struct {
	Value    []string          `json:"value"`
	Modifier CriterionModifier `json:"modifier"`
}

type StringCriterion struct {
	Value    string            `json:"value"`
	Modifier CriterionModifier `json:"modifier"`
}

type TimestampCriterion struct {
	Value    string            `json:"value"`
	Value2   *string           `json:"value2"`
	Modifier CriterionModifier `json:"modifier"`
}

type CriterionModifier int

const (
	CriterionModifierEquals CriterionModifier = iota
	CriterionModifierNotEquals
	CriterionModifierGreaterThan
	CriterionModifierLessThan
	CriterionModifierIsNull
	CriterionModifierNotNull
	CriterionModifierIncludesAll
	CriterionModifierIncludes
	CriterionModifierExcludes
	CriterionModifierMatchesRegex
	CriterionModifierNotMatchesRegex
	CriterionModifierBetween
	CriterionModifierNotBetween
)

var criterionModifierNames = []string{
	"EQUALS",
	"NOT_EQUALS",
	"GREATER_THAN",
	"LESS_THAN",
	"IS_NULL",
	"NOT_NULL",
	"INCLUDES_ALL",
	"INCLUDES",
	"EXCLUDES",
	"MATCHES_REGEX",
	"NOT_MATCHES_REGEX",
	"BETWEEN",
	"NOT_BETWEEN",
}

func (c CriterionModifier) String() string {
	if c < CriterionModifierEquals || c > CriterionModifierNotBetween {
		return ""
	}
	return criterionModifierNames[c]
}

func (c CriterionModifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *CriterionModifier) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	for i, v := range criterionModifierNames {
		if strings.EqualFold(s, v) {
			*c = CriterionModifier(i)
			return nil
		}
	}
	return fmt.Errorf("unknown CriterionModifier: %s", s)
}
