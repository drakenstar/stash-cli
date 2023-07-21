package app

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/stash"
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
	ColorGreen    = lipgloss.Color("#8FCB9B")
	ColorYellow   = lipgloss.Color("#FAD689")
	ColorPurple   = lipgloss.Color("#B39DDC")
	ColorGrey     = lipgloss.Color("#D3D3D3")
	ColorOffWhite = lipgloss.Color("#FAF0E6")
	ColorMidGrey  = lipgloss.Color("#808080")
	ColorSalmon   = lipgloss.Color("#FF9C8A")
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(ColorOffWhite).
			Bold(true).
			PaddingRight(1)

	performerStyle = lipgloss.NewStyle().
			Foreground(ColorYellow).
			PaddingRight(2)

	tagStyle = lipgloss.NewStyle().
			Foreground(ColorPurple).
			PaddingRight(1)

	emptyStyle = lipgloss.NewStyle().
			Foreground(ColorMidGrey).
			PaddingRight(1)

	studioStyle = lipgloss.NewStyle().
			Foreground(ColorSalmon).
			PaddingRight(1)

	organized = lipgloss.NewStyle().
			SetString("✓").
			Foreground(lipgloss.Color("#32CD32")).
			PaddingRight(1).
			String()
	notOrganized = emptyStyle.Copy().
			SetString("○").
			String()
)

func printLine(w io.Writer, s ...string) {
	fmt.Fprint(w, lipgloss.JoinHorizontal(lipgloss.Top, s...), "\n")
}

func (r Renderer) ContentList(a *App) {
	// width, _, _ := term.GetSize(int(r.Out.Fd()))
	if a.mode == FilterModeScenes {
		rw := sceneColWidths{}
		for _, scene := range a.scenesState.content {
			s := scenePresenter{scene}
			rw.title = max(rw.title, lipgloss.Width(s.Title))
			rw.studio = max(rw.studio, lipgloss.Width(s.Studio.Name))
			rw.performers = max(rw.performers, lipgloss.Width(s.performerList()))
			rw.tags = max(rw.tags, lipgloss.Width(s.tagList()))
		}
		for _, scene := range a.scenesState.content {
			s := scenePresenter{scene}
			printLine(r.Out, sceneRow(s, &rw)...)
		}
	} else {
		for _, g := range a.galleriesState.content {
			fmt.Fprintf(r.Out, "%s %s %s\n", g.ID, g.Title, g.FilePath())
		}
	}
}

func (r Renderer) ContentRow(a *App) {
	if a.mode == FilterModeScenes {
		printLine(r.Out, sceneRow(scenePresenter{a.Current().(stash.Scene)}, nil)...)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type sceneColWidths struct {
	title      int
	performers int
	studio     int
	tags       int
}

func sceneRow(s scenePresenter, colWidths *sceneColWidths) []string {
	row := []string{}

	if colWidths == nil {
		colWidths = &sceneColWidths{
			title:      lipgloss.Width(s.Title),
			studio:     lipgloss.Width(s.Studio.Name),
			performers: lipgloss.Width(s.performerList()),
			tags:       lipgloss.Width(s.tagList()),
		}
	}

	// Organized
	{
		if s.Organized {
			row = append(row, organized)
		} else {
			row = append(row, notOrganized)
		}
	}

	// Date
	{
		row = append(row, renderDate(s.Date))
	}

	// Title
	{
		if s.Title != "" {
			style := titleStyle.Copy()
			row = append(row, style.
				Width(colWidths.title+1).
				Render(s.Title))
		} else {
			style := emptyStyle.Copy()
			row = append(row, style.
				Width(colWidths.title+1).
				Render(truncate(colWidths.title, filepath.Base(s.FilePath()), "…")))
		}
	}

	// Studio
	{
		style := studioStyle.Copy()
		row = append(row, style.
			Width(colWidths.studio+1).
			Render(s.Studio.Name))
	}

	// Performers
	{
		style := performerStyle.Copy()
		row = append(row, style.
			Width(colWidths.performers+2).
			Render(s.performerList()))
	}

	// Tags
	{
		style := tagStyle.Copy()
		row = append(row, style.
			Width(colWidths.tags+1).
			Render(s.tagList()))
	}

	return row
}

const dateColWidth = 11

var dateStyle = lipgloss.NewStyle().
	Foreground(ColorGrey).
	Width(dateColWidth).
	PaddingRight(1)

func renderDate(d string) string {
	return dateStyle.Render(d)
}

type scenePresenter struct {
	stash.Scene
}

func (s scenePresenter) performerList() string {
	var names []string
	for _, p := range s.Performers {
		name := p.Name + " "
		if p.Gender != stash.GenderFemale {
			name += p.Gender.String() + "  "
		}
		name += p.Country.String()
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

func truncate(length int, s, suffix string) string {
	if lipgloss.Width(s) < length {
		return s
	}
	sw := lipgloss.Width(suffix)
	r := []rune(s)
	return string(r[:length-sw]) + suffix
}
