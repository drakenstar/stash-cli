package stash

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
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
	ID                 *IntCriterion               `json:"id,omitempty"`
	Title              *StringCriterion            `json:"title,omitempty"`
	Code               *StringCriterion            `json:"code,omitempty"`
	Details            *StringCriterion            `json:"details,omitempty"`
	Director           *StringCriterion            `json:"director,omitempty"`
	OSHash             *StringCriterion            `json:"oshash,omitempty"`
	Checksum           *StringCriterion            `json:"checksum,omitempty"`
	PHash              *StringCriterion            `json:"phash,omitempty"`
	PHashDistance      *PHashDistanceCriterion     `json:"phash_distance,omitempty"`
	Path               *StringCriterion            `json:"path,omitempty"`
	FileCount          *IntCriterion               `json:"file_count,omitempty"`
	Rating100          *IntCriterion               `json:"rating100,omitempty"`
	Organized          *bool                       `json:"organized,omitempty"`
	OCounter           *IntCriterion               `json:"o_counter,omitempty"`
	Resolution         *ResolutionCriterion        `json:"resolution,omitempty"`
	FrameRate          *IntCriterion               `json:"framerate,omitempty"`
	VideoCodec         *StringCriterion            `json:"video_codec,omitempty"`
	AudioCodec         *StringCriterion            `json:"audio_codec,omitempty"`
	Duration           *IntCriterion               `json:"duration,omitempty"`
	HasMarkers         *string                     `json:"has_markers,omitempty"`
	IsMissing          string                      `json:"is_missing,omitempty"`
	Studios            *HierarchicalMultiCriterion `json:"studios,omitempty"`
	Movies             *MultiCriterion             `json:"movies,omitempty"`
	Tags               *HierarchicalMultiCriterion `json:"tags,omitempty"`
	TagCount           *IntCriterion               `json:"tag_count,omitempty"`
	PerformerTags      *HierarchicalMultiCriterion `json:"performer_tags,omitempty"`
	PerformerFavourite *bool                       `json:"performer_favorite,omitempty"`
	PerformerAge       *IntCriterion               `json:"performer_age,omitempty"`
	Performers         *MultiCriterion             `json:"performers,omitempty"`
	PerformerCount     *IntCriterion               `json:"performer_count,omitempty"`
	URL                *StringCriterion            `json:"url,omitempty"`
	Interactive        *bool                       `json:"interactive,omitempty"`
	InteractiveSpeed   *IntCriterion               `json:"interactive_speed,omitempty"`
	Captions           *StringCriterion            `json:"captions,omitempty"`
	ResumeTime         *IntCriterion               `json:"resume_time,omitempty"`
	PlayCount          *IntCriterion               `json:"play_count,omitempty"`
	PlayDuration       *IntCriterion               `json:"play_duration,omitempty"`
	Date               *DateCriterion              `json:"date,omitempty"`
	CreatedAt          *TimestampCriterion         `json:"created_at,omitempty"`
	UpdatedAt          *TimestampCriterion         `json:"updated_at,omitempty"`
}

func (SceneFilter) GetGraphQLType() string {
	return "SceneFilterType"
}

type GalleryFilter struct {
	FilterCombinator[GalleryFilter]
	ID                 *IntCriterion               `json:"id,omitempty"`
	Title              *StringCriterion            `json:"title,omitempty"`
	Details            *StringCriterion            `json:"details,omitempty"`
	Checksum           *StringCriterion            `json:"checksum,omitempty"`
	Path               *StringCriterion            `json:"path,omitempty"`
	FileCount          *IntCriterion               `json:"file_count,omitempty"`
	IsMissing          string                      `json:"is_missing,omitempty"`
	IsZip              *bool                       `json:"is_zip,omitempty"`
	Rating100          *IntCriterion               `json:"rating100,omitempty"`
	Organized          *bool                       `json:"organized,omitempty"`
	AverageResolution  *ResolutionCriterion        `json:"average_resolution,omitempty"`
	HasChapters        *string                     `json:"has_chapters,omitempty"`
	Studios            *HierarchicalMultiCriterion `json:"studios,omitempty"`
	Tags               *HierarchicalMultiCriterion `json:"tags,omitempty"`
	TagCount           *IntCriterion               `json:"tag_count,omitempty"`
	PerformerTags      *HierarchicalMultiCriterion `json:"performer_tags,omitempty"`
	Performers         *MultiCriterion             `json:"performers,omitempty"`
	PerformerCount     *IntCriterion               `json:"performer_count,omitempty"`
	PerformerFavourite *bool                       `json:"performer_favorite,omitempty"`
	PerformerAge       *IntCriterion               `json:"performer_age,omitempty"`
	ImageCount         *IntCriterion               `json:"image_count,omitempty"`
	URL                *StringCriterion            `json:"url,omitempty"`
	Date               *DateCriterion              `json:"date,omitempty"`
	CreatedAt          *TimestampCriterion         `json:"created_at,omitempty"`
	UpdatedAt          *TimestampCriterion         `json:"updated_at,omitempty"`
	Code               *StringCriterion            `json:"code,omitempty"`
	Photographer       *StringCriterion            `json:"photographer,omitempty"`
}

func (GalleryFilter) GetGraphQLType() string {
	return "GalleryFilterType"
}

type MultiCriterion struct {
	Value    []string          `json:"value"`
	Modifier CriterionModifier `json:"modifier"`
}

type HierarchicalMultiCriterion struct {
	Value    []string          `json:"value"`
	Modifier CriterionModifier `json:"modifier"`
	Depth    int               `json:"depth,omitempty"`
	Excludes []string          `json:"excludes,omitempty"`
}

type IntCriterion struct {
	Value    int               `json:"value"`
	Value2   *int              `json:"value2,omitempty"`
	Modifier CriterionModifier `json:"modifier"`
}

type StringCriterion struct {
	Value    string            `json:"value"`
	Modifier CriterionModifier `json:"modifier"`
}

type DateCriterion struct {
	Value    time.Time
	Value2   *time.Time
	Modifier CriterionModifier
}

type dateCriterion struct {
	Value    string            `json:"value"`
	Value2   *string           `json:"value2,omitempty"`
	Modifier CriterionModifier `json:"modifier"`
}

func (c DateCriterion) MarshalJSON() ([]byte, error) {
	tmp := dateCriterion{
		Value:    c.Value.Format("2006-01-02"),
		Modifier: c.Modifier,
	}
	if c.Value2 != nil {
		value2Str := c.Value2.Format("2006-01-02")
		tmp.Value2 = &value2Str
	}
	return json.Marshal(tmp)
}

func (c *DateCriterion) UnmarshalJSON(data []byte) error {
	var tmp dateCriterion
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	var err error
	c.Value, err = time.Parse("2006-01-02", tmp.Value)
	if err != nil {
		return err
	}

	if tmp.Value2 != nil {
		t, err := time.Parse("2006-01-02", *tmp.Value2)
		if err != nil {
			return err
		}
		c.Value2 = &t
	}

	c.Modifier = tmp.Modifier
	return nil
}

type TimestampCriterion struct {
	Value    time.Time         `json:"value"`
	Value2   *time.Time        `json:"value2,omitempty"`
	Modifier CriterionModifier `json:"modifier"`
}

type PHashDistanceCriterion struct {
	Value    string            `json:"value"`
	Modifier CriterionModifier `json:"modifier"`
	Distance *int              `json:"distance,omitempty"`
}

type ResolutionCriterion struct {
	Value    Resolution        `json:"value"`
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

type Resolution int

const (
	ResolutionVeryLow    Resolution = iota // 144p
	ResolutionLow                          // 240p
	ResolutionR360P                        // 360p
	ResolutionStandard                     // 480p
	ResolutionWebHD                        // 540p
	ResolutionStandardHD                   // 720p
	ResolutionFullHD                       // 1080p
	ResolutionQuadHD                       // 1440p
	ResolutionFourK                        // 4K
	ResolutionFiveK                        // 5K
	ResolutionSixK                         // 6K
	ResolutionSevenK                       // 7K
	ResolutionEightK                       // 8K
	ResolutionHuge                         // 8K+
)

var resolutionNames = []string{
	"VERY_LOW",
	"LOW",
	"R360P",
	"STANDARD",
	"WEB_HD",
	"STANDARD_HD",
	"FULL_HD",
	"QUAD_HD",
	"FOUR_K",
	"FIVE_K",
	"SIX_K",
	"SEVEN_K",
	"EIGHT_K",
	"HUGE",
}

func (r Resolution) String() string {
	if r < ResolutionVeryLow || r > ResolutionHuge {
		return ""
	}
	return resolutionNames[r]
}

func (r Resolution) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *Resolution) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	for i, v := range resolutionNames {
		if strings.EqualFold(s, v) {
			*r = Resolution(i)
			return nil
		}
	}
	return fmt.Errorf("unknown Resolution: %s", s)
}
