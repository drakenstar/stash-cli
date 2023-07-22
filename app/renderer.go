package app

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
	"golang.org/x/term"
)

type Filer interface {
	io.Writer
	Fd() uintptr
}

type Renderer struct {
	Out Filer
}

func (r Renderer) Prompt(a *App) {
	stats := a.Stats()
	fmt.Fprint(r.Out, promptStyle.Render(fmt.Sprintf("%s (%d/%d) >> ", a.mode.Icon(), stats.Index+1, stats.Total)))
}

var (
	ColorGreenCheck = lipgloss.Color("#32CD32")
	ColorGreen      = lipgloss.Color("#8FCB9B")
	ColorYellow     = lipgloss.Color("#FAD689")
	ColorPurple     = lipgloss.Color("#B39DDC")
	ColorGrey       = lipgloss.Color("#D3D3D3")
	ColorOffWhite   = lipgloss.Color("#FAF0E6")
	ColorMidGrey    = lipgloss.Color("#808080")
	ColorSalmon     = lipgloss.Color("#FF9C8A")

	check = lipgloss.NewStyle().
		Foreground(ColorGreenCheck).
		SetString("✓").
		String()
	circle = lipgloss.NewStyle().
		Foreground(ColorGrey).
		SetString("○").
		Render()
)

func (r Renderer) ContentList(a *App) {
	screenWidth, _, _ := term.GetSize(int(r.Out.Fd()))
	if a.mode == FilterModeScenes {
		var rows []ui.Row
		for _, s := range a.scenesState.content {
			scene := scenePresenter{s}
			rows = append(rows, []string{
				scene.organised(),
				scene.Date,
				scene.Title,
				scene.Studio.Name,
				scene.performerList(),
				scene.tagList(),
				scene.Details,
			})
		}
		fmt.Fprint(r.Out, ui.RenderTable(screenWidth, []ui.Column{
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
		}, rows)+"\n")
	} else {
		for _, g := range a.galleriesState.content {
			fmt.Fprintf(r.Out, "%s %s %s\n", g.ID, g.Title, g.FilePath())
		}
	}
}

func (r Renderer) ContentRow(a *App) {
	screenWidth, _, _ := term.GetSize(int(r.Out.Fd()))
	if a.mode == FilterModeScenes {
		scene := scenePresenter{a.Current().(stash.Scene)}
		fmt.Fprint(r.Out, ui.RenderTable(screenWidth, []ui.Column{
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
		}, []ui.Row{
			{
				scene.organised(),
				scene.Date,
				scene.Title,
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
