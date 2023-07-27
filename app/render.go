package app

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/stash"
)

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
)

type scenePresenter struct {
	stash.Scene
}

func (s scenePresenter) title() string {
	if s.Title != "" {
		return s.Title
	}
	fileName := filepath.Base(s.FilePath())
	parentDir := filepath.Base(filepath.Dir(s.FilePath()))
	return filepath.Join(parentDir, fileName)
}

func (s scenePresenter) performerList() string {
	return performerList(s.Performers)
}

func performerList(performers []stash.Performer) string {
	var names []string
	for _, p := range performers {
		name := p.Name
		if p.Gender != stash.GenderFemale {
			name += " " + p.Gender.String()
		}
		if p.Country != "" {
			name += " " + p.Country.String()
		}
		if p.Favorite {
			name += " ❤️"
		}
		names = append(names, name)
	}
	return strings.Join(names, "\n")
}

func (s scenePresenter) tagList() string {
	return tagList(s.Tags)
}

func tagList(tags []stash.Tag) string {
	var tagStrings []string
	for _, t := range tags {
		tagStrings = append(tagStrings, t.Name)
	}
	return strings.Join(tagStrings, ", ")
}

func (s scenePresenter) organised() string {
	return organised(s.Organized)
}

func organised(o bool) string {
	if o {
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

type galleryPresenter struct {
	stash.Gallery
}

func (g galleryPresenter) organised() string {
	return organised(g.Organized)
}

func (g galleryPresenter) title() string {
	if g.Title != "" {
		return g.Title
	}
	return filepath.Base(g.FilePath())
}

func (g galleryPresenter) size() string {
	return strconv.Itoa(g.ImageCount)
}

func (g galleryPresenter) performerList() string {
	return performerList(g.Performers)
}

func (g galleryPresenter) tagList() string {
	return tagList(g.Tags)
}

func (g galleryPresenter) details() string {
	return strings.ReplaceAll(g.Details, "\n", " ")
}
