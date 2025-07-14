package app

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

var (
	ColorBlack       = lipgloss.Color("#000000")
	ColorGreen       = lipgloss.Color("#32CD32")
	ColorYellow      = lipgloss.Color("#FAD689")
	ColorPurple      = lipgloss.Color("#B39DDC")
	ColorGrey        = lipgloss.Color("#D3D3D3")
	ColorWhite       = lipgloss.Color("#FFFFFF")
	ColorOffWhite    = lipgloss.Color("#FAF0E6")
	ColorSalmon      = lipgloss.Color("#FF9C8A")
	ColorBlue        = lipgloss.Color("#A2D2FF")
	ColorRowSelected = lipgloss.Color("#28664A")
	ColorRed         = lipgloss.Color("#FF0000")

	ColorStatusBar  = lipgloss.Color("#2B2A60")
	ColorStatusCell = lipgloss.Color("#483D8B")

	check = lipgloss.NewStyle().
		Foreground(ColorGreen).
		SetString("\uf00c").
		String()
	circle = lipgloss.NewStyle().
		Foreground(ColorGrey).
		SetString("○").
		Render()

	tabBar = ui.Tabs{
		NumberForeground: ColorWhite,
		TitleForeground:  ColorOffWhite,
		Background:       ColorBlack,
		ActiveBackground: ColorStatusCell,
	}

	statusBar = ui.StatusBar{
		Background:     ColorStatusBar,
		CellBackground: ColorStatusCell,
	}
)

func sceneTitle(s stash.Scene) string {
	if s.Title != "" {
		return s.Title
	}
	fileName := filepath.Base(s.FilePath())
	parentDir := filepath.Base(filepath.Dir(s.FilePath()))
	return filepath.Join(parentDir, fileName)
}

func performerList(performers []stash.Performer) string {
	switch len(performers) {
	case 0:
		return ""
	case 1:
		return performer(performers[0])
	default:
		var show stash.Performer
		for _, p := range performers {
			if p.SceneCount > show.SceneCount {
				show = p
			}
		}
		return fmt.Sprintf("%s (+%d)", performer(show), len(performers)-1)
	}
}

func studioList(studios []stash.Studio) string {
	ss := make([]string, len(studios))
	for _, s := range studios {
		ss = append(ss, s.Name)
	}
	return strings.Join(ss, ", ")
}

func performer(p stash.Performer) string {
	name := p.Name
	if p.Gender != stash.GenderFemale {
		name += " " + p.Gender.String()
	}
	if p.Country != "" {
		name += " " + p.Country.String()
	}
	if p.Favorite {
		name += lipgloss.NewStyle().Foreground(ColorRed).Render(" \U000f02d1")
	}
	return name
}

func tagList(tags []stash.Tag) string {
	var tagStrings []string
	for _, t := range tags {
		tagStrings = append(tagStrings, t.Name)
	}
	return strings.Join(tagStrings, ", ")
}

func organised(o bool) string {
	if o {
		return check
	}
	return circle
}

func details(details string) string {
	return strings.ReplaceAll(details, "\n", " ")
}

func sceneSize(s stash.Scene) string {
	if len(s.Files) > 0 {
		return humanBytes(s.Files[0].Size)
	}
	return ""
}

func humanBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	value := float64(bytes) / float64(div)
	if exp < 2 {
		return fmt.Sprintf("%.0f%c", value, "KMGTPE"[exp])
	} else {
		return fmt.Sprintf("%.1f%c", value, "KMGTPE"[exp])
	}
}

func galleryTitle(g stash.Gallery) string {
	if g.Title != "" {
		return g.Title
	}
	// TODO Fix this hard-coded library path
	rel, err := filepath.Rel("/library", g.FilePath())
	if err == nil {
		return rel
	}
	return filepath.Base(g.FilePath())
}

func gallerySize(g stash.Gallery) string {
	return strconv.Itoa(g.ImageCount)
}

func sort(order, direction string) string {
	var sort string
	if strings.HasPrefix(order, stash.SortRandomPrefix) {
		sort = "random"
	} else {
		sort = order
	}
	if direction == stash.SortDirectionAsc {
		sort += " ▲"
	} else {
		sort += " ▼"
	}
	return sort
}

// StashRenderLookupService is a service to lookup stash entities by ID.  These methods should not require I/O and
// should either return a valid result, or error.  Since this is used in render methods, we do not want to block, nor
// should we have side-effects or error handling.
type StashLookup interface {
	GetStudio(id string) (stash.Studio, error)
	GetTag(id string) (stash.Tag, error)
	GetPerformer(id string) (stash.Performer, error)
}

// sceneFilterStatus takes a stash.SceneFilter and returns a slice of strings, each string representing an enabled
// filter.  This can be used to display UI to the user with regard to what is being filtered on.
func sceneFilterStatus(filter stash.SceneFilter, srv StashLookup) []string {
	var status criterionRenderer

	// FilterCombinator is ignored for this render function.
	status.intCriterion("ID", filter.ID)
	status.stringCriterion("Title", filter.Title)
	status.stringCriterion("Code", filter.Code)
	status.stringCriterion("Details", filter.Details)
	status.stringCriterion("Director", filter.Director)
	status.stringCriterion("OSHash", filter.OSHash)
	status.stringCriterion("Checksum", filter.Checksum)
	status.stringCriterion("PHash", filter.PHash)
	status.stringCriterion("Path", filter.Path)
	status.pHashDistanceCriterion("PHash distance", filter.PHashDistance)
	status.intCriterion("File count", filter.FileCount)
	status.intCriterion("Rating", filter.Rating100)
	status.boolCriterion(filter.Organized, "Organised", "Unorganised")
	status.intCriterion("O-counter", filter.OCounter)
	status.resolutionCriterion("Resolution", filter.Resolution)
	status.intCriterion("Frame rate", filter.FrameRate)
	status.stringCriterion("VideoCodec", filter.VideoCodec)
	status.stringCriterion("AudioCodec", filter.AudioCodec)
	status.intCriterion("Duration", filter.Duration)
	if filter.HasMarkers != nil {
		status = append(status, *filter.HasMarkers)
	}
	// if filter.IsMissing != "" {
	// 	status = append(status, "Is missing "+filter.IsMissing)
	// }
	status.heirarchicalMultiCriterion("Studios", filter.Studios, func(id string) string {
		studio, err := srv.GetStudio(id)
		if err != nil {
			return "error studio"
		}
		return studio.Name

	})
	status.multiCriterion("Movies", filter.Movies, func(id string) string {
		return id // TODO: movie cache
	})
	status.heirarchicalMultiCriterion("Tags", filter.Tags, func(id string) string {
		tag, err := srv.GetTag(id)
		if err != nil {
			return "error tag"
		}
		return tag.Name
	})
	status.intCriterion("Tag #", filter.TagCount)
	status.heirarchicalMultiCriterion("Performer tags", filter.PerformerTags, func(id string) string {
		tag, err := srv.GetTag(id)
		if err != nil {
			return "error tag"
		}
		return tag.Name
	})
	status.boolCriterion(filter.PerformerFavourite, "Favourite", "Non-favourite")
	status.intCriterion("Age", filter.PerformerAge)
	status.multiCriterion("Performers", filter.Performers, func(id string) string {
		performer, err := srv.GetPerformer(id)
		if err != nil {
			return "error performer"
		}
		return performer.Name
	})
	status.intCriterion("Performer #", filter.PerformerCount)
	status.stringCriterion("URL", filter.URL)
	status.boolCriterion(filter.Interactive, "Interactive", "Non-interactive")
	status.intCriterion("Interactive speed", filter.InteractiveSpeed)
	status.stringCriterion("Captions", filter.Captions)
	status.intCriterion("Resume time", filter.ResumeTime)
	status.intCriterion("Play #", filter.PlayCount)
	status.intCriterion("Play duration", filter.PlayDuration)
	status.dateCriterion("Date", filter.Date)
	status.timestampCriterion("Created", filter.CreatedAt)
	status.timestampCriterion("Updated", filter.UpdatedAt)

	return status
}

func galleryFilterStatus(filter stash.GalleryFilter, srv StashLookup) []string {
	var status criterionRenderer

	status.intCriterion("ID", filter.ID)
	status.stringCriterion("Title", filter.Title)
	status.stringCriterion("Details", filter.Details)
	status.stringCriterion("Checksum", filter.Checksum)
	status.stringCriterion("Path", filter.Path)
	status.intCriterion("File #", filter.FileCount)
	if filter.IsMissing != "" {
		status = append(status, "Is missing "+filter.IsMissing)
	}
	status.boolCriterion(filter.IsZip, "Zip", "Non-zip")
	status.intCriterion("Rating", filter.Rating100)
	status.boolCriterion(filter.Organized, "Organised", "Unorganised")
	status.resolutionCriterion("Resolution", filter.AverageResolution)
	// TODO: HasChapters
	status.heirarchicalMultiCriterion("Studios", filter.Studios, func(id string) string {
		studio, err := srv.GetStudio(id)
		if err != nil {
			return "error studio"
		}
		return studio.Name
	})
	status.heirarchicalMultiCriterion("Tags", filter.Tags, func(id string) string {
		tag, err := srv.GetTag(id)
		if err != nil {
			return "error tag"
		}
		return tag.Name
	})
	status.intCriterion("Tag #", filter.TagCount)
	status.heirarchicalMultiCriterion("Performer tags", filter.PerformerTags, func(id string) string {
		tag, err := srv.GetTag(id)
		if err != nil {
			return "error tag"
		}
		return tag.Name
	})
	status.multiCriterion("Performers", filter.Performers, func(id string) string {
		performer, err := srv.GetPerformer(id)
		if err != nil {
			return "error performer"
		}
		return performer.Name
	})
	status.intCriterion("Performer #", filter.PerformerCount)
	status.boolCriterion(filter.PerformerFavourite, "Favourite", "Non-favourite")
	status.intCriterion("Age", filter.PerformerAge)
	status.intCriterion("Image #", filter.ImageCount)
	status.stringCriterion("URL", filter.URL)
	status.dateCriterion("Date", filter.Date)
	status.timestampCriterion("Created", filter.CreatedAt)
	status.timestampCriterion("Updated", filter.UpdatedAt)
	status.stringCriterion("Code", filter.Code)
	status.stringCriterion("Photographer", filter.Photographer)

	return status
}

var criterionModifierTemplates []*template.Template

func init() {
	templates := []string{
		"{{.FieldLabel}} is {{.Value}}",
		"{{.FieldLabel}} is not {{.Value}}",
		"{{.FieldLabel}} greater than {{.Value}}",
		"{{.FieldLabel}} less than {{.Value}}",
		"{{.FieldLabel}} is null",
		"{{.FieldLabel}} is not null",
		"{{.FieldLabel}} includes all {{.Value | join}}",
		"{{.FieldLabel}} includes {{.Value | join}}",
		"{{.FieldLabel}} excludes {{.Value | join}}",
		"{{.FieldLabel}} matches regex {{.Value | join}}",
		"{{.FieldLabel}} doesn't match regex {{.Value | join}}",
		"{{.FieldLabel}} between {{.Value}} and {{.Value2}}",
		"{{.FieldLabel}} not between {{.Value}} and {{.Value2}}",
	}
	for i, t := range templates {
		tmp, err := template.New(stash.CriterionModifier(i).String()).
			Funcs(template.FuncMap{
				"join": tmplJoin,
			}).
			Parse(t)
		if err != nil {
			panic(err)
		}
		criterionModifierTemplates = append(criterionModifierTemplates, tmp)
	}
}

func tmplJoin(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case []string:
		return strings.Join(v, ", ")
	default:
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Slice {
			var s []string
			for i := 0; i < val.Len(); i++ {
				elem := val.Index(i).Interface()
				if str, ok := elem.(fmt.Stringer); ok {
					s = append(s, str.String())
				} else {
					s = append(s, fmt.Sprintf("%v", elem))
				}
			}
			return strings.Join(s, ", ")
		}
		return fmt.Sprintf("%v", v)
	}
}

type criterionRenderer []string

type criterionData struct {
	FieldLabel string
	Value      any
	Value2     any
}

func renderCriterion(m stash.CriterionModifier, data criterionData) string {
	var b []byte
	w := bytes.NewBuffer(b)
	err := criterionModifierTemplates[m].Execute(w, data)
	if err != nil {
		panic(err)
	}
	return w.String()
}

func (r *criterionRenderer) intCriterion(fieldLabel string, c *stash.IntCriterion) {
	if c == nil {
		return
	}
	*r = append(*r, renderCriterion(c.Modifier, criterionData{
		FieldLabel: fieldLabel,
		Value:      c.Value,
		Value2:     c.Value2,
	}))
}

func (r *criterionRenderer) stringCriterion(fieldLabel string, c *stash.StringCriterion) {
	if c == nil {
		return
	}
	*r = append(*r, renderCriterion(c.Modifier, criterionData{
		FieldLabel: fieldLabel,
		Value:      c.Value,
	}))
}

func (r *criterionRenderer) pHashDistanceCriterion(fieldLabel string, c *stash.PHashDistanceCriterion) {
	if c == nil {
		return
	}
	panic("not implemented")
}

var resolutionLabels = []string{
	"144p",
	"240p",
	"360p",
	"480p",
	"540p",
	"720p",
	"1080p",
	"1440p",
	"4K",
	"5K",
	"6K",
	"7K",
	"8K",
	"8K+",
}

func (r *criterionRenderer) resolutionCriterion(fieldLabel string, c *stash.ResolutionCriterion) {
	if c == nil {
		return
	}
	*r = append(*r, renderCriterion(c.Modifier, criterionData{
		FieldLabel: fieldLabel,
		Value:      resolutionLabels[c.Value],
	}))
}

func (r *criterionRenderer) dateCriterion(fieldLabel string, c *stash.DateCriterion) {
	if c == nil {
		return
	}
	data := criterionData{
		FieldLabel: fieldLabel,
		Value:      c.Value.Format("2006-01-02"),
	}
	if c.Value2 != nil {
		data.Value2 = c.Value2.Format("2006-01-02")
	}
	*r = append(*r, renderCriterion(c.Modifier, data))
}

func (r *criterionRenderer) timestampCriterion(fieldLabel string, c *stash.TimestampCriterion) {
	r.dateCriterion(fieldLabel, (*stash.DateCriterion)(c))
}

func (r *criterionRenderer) multiCriterion(fieldLabel string, c *stash.MultiCriterion, labelFunc func(string) string) {
	if c == nil {
		return
	}
	p := make([]string, len(c.Value))
	for i, v := range c.Value {
		p[i] = labelFunc(v)
	}
	*r = append(*r, renderCriterion(c.Modifier, criterionData{
		FieldLabel: fieldLabel,
		Value:      p,
	}))
}

func (r *criterionRenderer) heirarchicalMultiCriterion(fieldLabel string, c *stash.HierarchicalMultiCriterion, labelFunc func(string) string) {
	if c == nil {
		return
	}
	// TODO better understand the function of this criterion, currently just backing onto the existing MultiCriterion.
	r.multiCriterion(fieldLabel, &stash.MultiCriterion{
		Value:    c.Value,
		Modifier: c.Modifier,
	}, labelFunc)
}
func (r *criterionRenderer) boolCriterion(c *bool, trueValue, falseValue string) {
	if c == nil {
		return
	}
	if *c {
		*r = append(*r, trueValue)
	} else {
		*r = append(*r, falseValue)
	}
}

func humanNumber(n int) string {
	switch {
	case n < 1_000:
		return fmt.Sprintf("%d", n)
	case n < 10_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%dK", n/1_000)
	}
}
