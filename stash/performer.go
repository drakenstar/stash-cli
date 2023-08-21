package stash

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

type Performer struct {
	ID        string  `graphql:"id"`
	Name      string  `graphql:"name"`
	URL       string  `graphql:"url"`
	Birthdate string  `graphql:"birthdate"`
	Gender    Gender  `graphql:"gender"`
	Country   Country `graphql:"country"`
	Favorite  bool    `graphql:"favorite"`
}

func (p Performer) EntityID() string {
	return p.ID
}

type Gender int

const (
	GenderNotSpecified = iota
	GenderMale
	GenderFemale
	GenderTransMale
	GenderTransFemale
	GenderIntersex
	GenderNonBinary
)

func (g Gender) MarshalJSON() ([]byte, error) {
	switch g {
	case GenderNotSpecified:
		return json.Marshal("")
	case GenderMale:
		return json.Marshal("MALE")
	case GenderFemale:
		return json.Marshal("FEMALE")
	case GenderTransMale:
		return json.Marshal("TRANSGENDER_MALE")
	case GenderTransFemale:
		return json.Marshal("TRANSGENDER_FEMALE")
	case GenderIntersex:
		return json.Marshal("INTERSEX")
	case GenderNonBinary:
		return json.Marshal("NON_BINARY")
	default:
		return nil, errors.New("invalid Gender value")
	}
}

func (g *Gender) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "":
		*g = GenderNotSpecified
	case "MALE":
		*g = GenderMale
	case "FEMALE":
		*g = GenderFemale
	case "TRANSGENDER_MALE":
		*g = GenderTransMale
	case "TRANSGENDER_FEMALE":
		*g = GenderTransFemale
	case "INTERSEX":
		*g = GenderIntersex
	case "NON_BINARY":
		*g = GenderNonBinary
	default:
		return errors.New("invalid Gender string")
	}

	return nil
}

func (g Gender) String() string {
	switch g {
	case GenderNotSpecified:
		return ""
	case GenderMale:
		return "♂️"
	case GenderFemale:
		return "♂️"
	case GenderTransMale:
		return "⚧️"
	case GenderTransFemale:
		return "⚧️"
	case GenderIntersex:
		return "⚧️"
	case GenderNonBinary:
		return "⚧️"
	default:
		panic("invalid Gender value")
	}
}

type Country string

func (c Country) String() string {
	if c == "" {
		return ""
	}
	code := strings.ToUpper(string(c))
	return string(0x1F1E6+rune(code[0])-'A') + string(0x1F1E6+rune(code[1])-'A')
}

type PerformerSummary struct {
	ID             string   `graphql:"id"`
	Name           string   `graphql:"name"`
	Disambiguation string   `graphql:"disambiguation"`
	Aliases        []string `graphql:"alias_list"`
}

func (PerformerSummary) GetGraphQLType() string {
	return "Performer"
}

type allPerformersQuery struct {
	Performers []PerformerSummary `graphql:"allPerformers"`
}

// PerformersAll returns a slice containing all performers.  Due to the potential size of this slice, only ID, Name,
// and Alises are requested.
func (s stash) PerformersAll(ctx context.Context) ([]PerformerSummary, error) {
	resp := allPerformersQuery{
		Performers: make([]PerformerSummary, 0),
	}
	err := s.client.Query(ctx, &resp, nil)
	if err != nil {
		return nil, err
	}
	return resp.Performers, nil
}

type PerformerCreate struct {
	Name           string  `json:"name"`
	Disambiguation string  `json:"disambiguation,omitempty"`
	URL            *string `json:"url,omitempty"`
}

func (PerformerCreate) GetGraphQLType() string {
	return "PerformerCreateInput"
}

// PerformerCreate creates a new performer in the stash instance and returns it with it's ID value.
func (s stash) PerformerCreate(ctx context.Context, p PerformerCreate) (Performer, error) {
	var m struct {
		Performer Performer `graphql:"performerCreate(input: $input)"`
	}
	err := s.client.Mutate(ctx, &m, map[string]any{"input": p})
	return m.Performer, err
}
