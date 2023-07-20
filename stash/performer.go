package stash

import (
	"encoding/json"
	"errors"
)

type Performer struct {
	ID        string `graphql:"id"`
	Name      string `graphql:"name"`
	Birthdate string `graphql:"birthdate"`
	Gender    Gender `graphql:"gender"`
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
