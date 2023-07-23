package app

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type Output interface {
	io.Writer
	ScreenWidth() int
}

type Renderer struct {
	Out Output
}

func (r Renderer) Prompt(a *App) {
	stats := a.Stats()
	fmt.Fprint(r.Out, promptStyle.Render(fmt.Sprintf("%s (%d/%d) >> ", a.mode.Icon(), stats.Index+1, stats.Total)))
}

var (
	ColorGreen    = lipgloss.Color("#32CD32")
	ColorYellow   = lipgloss.Color("#FAD689")
	ColorPurple   = lipgloss.Color("#B39DDC")
	ColorGrey     = lipgloss.Color("#D3D3D3")
	ColorOffWhite = lipgloss.Color("#FAF0E6")
	ColorSalmon   = lipgloss.Color("#FF9C8A")
	ColorBlue     = lipgloss.Color("#A2D2FF")

	check = lipgloss.NewStyle().
		Foreground(ColorGreen).
		SetString("✓").
		String()
	circle = lipgloss.NewStyle().
		Foreground(ColorGrey).
		SetString("○").
		Render()

	sceneTable = &ui.Table{
		Cols: []ui.Column{
			{
				Name: "Organised",
			},
			{
				Name:       "Date",
				Foreground: &ColorGrey,
			},
			{
				Name:       "Title",
				Foreground: &ColorOffWhite,
				Bold:       true,
				Weight:     1,
			},
			{
				Name:       "Size",
				Foreground: &ColorBlue,
				Align:      lipgloss.Right,
			},
			{
				Name:       "Studio",
				Foreground: &ColorSalmon,
				Weight:     1,
			},
			{
				Name:       "Perfomers",
				Foreground: &ColorYellow,
				Weight:     1,
			},
			{
				Name:       "Tags",
				Foreground: &ColorPurple,
				Weight:     1,
			},
			{
				Name:       "Description",
				Foreground: &ColorGrey,
				Flex:       true,
			},
		},
	}
	sceneRow = &ui.Table{
		Cols: []ui.Column{
			{
				Name: "Organised",
			},
			{
				Name:       "Date",
				Foreground: &ColorGrey,
			},
			{
				Name:       "Title",
				Foreground: &ColorOffWhite,
				Bold:       true,
				Weight:     0,
			},
			{
				Name:       "Size",
				Foreground: &ColorBlue,
				Align:      lipgloss.Right,
			},
			{
				Name:       "Studio",
				Foreground: &ColorSalmon,
				Weight:     1,
			},
			{
				Name:       "Perfomers",
				Foreground: &ColorYellow,
				Weight:     1,
			},
			{
				Name:       "Tags",
				Foreground: &ColorPurple,
				Weight:     1,
			},
		},
	}
)

func (r Renderer) ContentList(a *App) {
	screenWidth := r.Out.ScreenWidth()
	if a.mode == FilterModeScenes {
		var rows []ui.Row
		for _, s := range a.scenesState.content {
			scene := scenePresenter{s}
			rows = append(rows, []string{
				scene.organised(),
				scene.Date,
				scene.title(),
				scene.size(),
				scene.Studio.Name,
				scene.performerList(),
				scene.tagList(),
				scene.details(),
			})
		}
		fmt.Fprint(r.Out, sceneTable.Render(screenWidth, rows)+"\n")
	} else {
		for _, g := range a.galleriesState.content {
			fmt.Fprintf(r.Out, "%s %s %s\n", g.ID, g.Title, g.FilePath())
		}
	}
}

func (r Renderer) ContentRow(a *App) {
	screenWidth := r.Out.ScreenWidth()
	if a.mode == FilterModeScenes {
		scene := scenePresenter{a.Current().(stash.Scene)}
		fmt.Fprint(r.Out, sceneRow.Render(screenWidth, []ui.Row{
			{
				scene.organised(),
				scene.Date,
				scene.title(),
				scene.size(),
				scene.Studio.Name,
				scene.performerList(),
				scene.tagList(),
			},
		})+"\n")
	}
}

type scenePresenter struct {
	stash.Scene
}

func (s scenePresenter) title() string {
	if s.Title != "" {
		return s.Title
	}
	return filepath.Base(s.FilePath())
}

func (s scenePresenter) performerList() string {
	var names []string
	for _, p := range s.Performers {
		name := p.Name + " "
		if p.Gender != stash.GenderFemale {
			name += p.Gender.String() + "  "
		}
		name += p.Country.String() + " "
		if p.Favorite {
			name += " ❤️"
		}
		names = append(names, name)
	}
	return strings.Join(names, "\n")
}

func (s scenePresenter) tagList() string {
	var tags []string
	for _, t := range s.Tags {
		tags = append(tags, t.Name)
	}
	return strings.Join(tags, ", ")
}

func (s scenePresenter) organised() string {
	if s.Scene.Organized {
		return check
	}
	return circle
}

func (s scenePresenter) details() string {
	return strings.ReplaceAll(s.Details, "\n", " ")
}

func (s scenePresenter) size() string {
	if len(s.Files) > 0 {
		return humanBytes(s.Files[0].Size)
	}
	return ""
}

func humanBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
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
