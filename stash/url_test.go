package stash

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURL(t *testing.T) {
	u, err := url.Parse(
		"http://mendev.local:9999/scenes?" +
			"q=test&" +
			"c=(%22type%22:%22audio_codec%22,%22value%22:%22aac%22,%22modifier%22:%22NOT_EQUALS%22)&" +
			"c=(%22type%22:%22captions%22,%22value%22:%22English%22,%22modifier%22:%22EXCLUDES%22)&" +
			"c=(%22type%22:%22checksum%22,%22value%22:%22test%22,%22modifier%22:%22INCLUDES%22)&" +
			"c=(%22type%22:%22organized%22,%22value%22:%22true%22,%22modifier%22:%22EQUALS%22)&" +
			"c=(%22type%22:%22created_at%22,%22value%22:(%22value%22:%222024-02-01%2000:00%22,%22value2%22:%222024-02-16%2000:00%22),%22modifier%22:%22BETWEEN%22)&" +
			"c=(%22type%22:%22date%22,%22value%22:(%22value%22:%222024-02-16%22),%22modifier%22:%22LESS_THAN%22)&" +
			"c=(%22type%22:%22details%22,%22value%22:%22foo%22,%22modifier%22:%22EQUALS%22)&" +
			"c=(%22type%22:%22director%22,%22value%22:%22scorsese%22,%22modifier%22:%22EQUALS%22)&" +
			"c=(%22type%22:%22duplicated%22,%22value%22:%22true%22,%22modifier%22:%22EQUALS%22)&" + // TODO
			"c=(%22type%22:%22duration%22,%22value%22:(%22value%22:60),%22modifier%22:%22GREATER_THAN%22)&" +
			"c=(%22type%22:%22file_count%22,%22value%22:(%22value%22:3,%22value2%22:10),%22modifier%22:%22NOT_BETWEEN%22)&" +
			"c=(%22type%22:%22framerate%22,%22value%22:(%22value%22:60),%22modifier%22:%22GREATER_THAN%22)&" +
			"c=(%22type%22:%22has_markers%22,%22value%22:%22true%22,%22modifier%22:%22EQUALS%22)&" +
			"sortby=updated_at&" +
			"sortdir=desc&" +
			"perPage=60&" +
			"p=2",
	)
	require.NoError(t, err)

	r, err := ParseUrl(u)
	require.NoError(t, err)
	assert.Equal(t, "/scenes", r.Path)
	assert.Equal(t, FindFilter{
		Query:     "test",
		Page:      2,
		PerPage:   60,
		Sort:      "updated_at",
		Direction: SortDirectionDesc,
	}, r.FindFilter)
	assert.Equal(t, &SceneFilter{
		AudioCodec: &StringCriterion{
			Value:    "aac",
			Modifier: CriterionModifierNotEquals,
		},
		Captions: &StringCriterion{
			Value:    "English",
			Modifier: CriterionModifierExcludes,
		},
		Checksum: &StringCriterion{
			Value:    "test",
			Modifier: CriterionModifierIncludes,
		},
		CreatedAt: &TimestampCriterion{
			Value:    time.Date(2024, 02, 1, 0, 0, 0, 0, time.UTC),
			Value2:   ptr(time.Date(2024, 02, 16, 0, 0, 0, 0, time.UTC)),
			Modifier: CriterionModifierBetween,
		},
		Date: &DateCriterion{
			Value:    time.Date(2024, 02, 16, 0, 0, 0, 0, time.UTC),
			Modifier: CriterionModifierLessThan,
		},
		Details: &StringCriterion{
			Value:    "foo",
			Modifier: CriterionModifierEquals,
		},
		Director: &StringCriterion{
			Value:    "scorsese",
			Modifier: CriterionModifierEquals,
		},
		Duration: &IntCriterion{
			Value:    60,
			Modifier: CriterionModifierGreaterThan,
		},
		FileCount: &IntCriterion{
			Value:    3,
			Value2:   ptr(10),
			Modifier: CriterionModifierNotBetween,
		},
		FrameRate: &IntCriterion{
			Value:    60,
			Modifier: CriterionModifierGreaterThan,
		},
		HasMarkers: ptr("true"),
		Organized:  ptr(true),
	}, r.SceneFilter)
}

func ptr[T any](v T) *T {
	return &v
}
