package app

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

var (
	ColorGreen       = lipgloss.Color("#32CD32")
	ColorYellow      = lipgloss.Color("#FAD689")
	ColorPurple      = lipgloss.Color("#B39DDC")
	ColorGrey        = lipgloss.Color("#D3D3D3")
	ColorOffWhite    = lipgloss.Color("#FAF0E6")
	ColorSalmon      = lipgloss.Color("#FF9C8A")
	ColorBlue        = lipgloss.Color("#A2D2FF")
	ColorRowSelected = lipgloss.Color("#28664A")

	ColorStatusBar  = lipgloss.Color("#2B2A60")
	ColorStatusCell = lipgloss.Color("#483D8B")

	check = lipgloss.NewStyle().
		Foreground(ColorGreen).
		SetString("✓").
		String()
	circle = lipgloss.NewStyle().
		Foreground(ColorGrey).
		SetString("○").
		Render()

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
