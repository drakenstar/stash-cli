package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Footer struct {
	Background lipgloss.Color

	loadingCount uint
	spinner      spinner.Model
}

func NewFooter() Footer {
	return Footer{
		spinner: spinner.New(spinner.WithSpinner(spinner.Globe)),
	}
}

func (f Footer) Init() tea.Cmd {
	return f.spinner.Tick
}

func (f Footer) Update(msg tea.Msg) (Footer, tea.Cmd) {
	var cmd tea.Cmd
	f.spinner, cmd = f.spinner.Update(msg)
	return f, cmd
}

func (f Footer) Render(width int, loading bool) string {
	style := lipgloss.NewStyle().
		Background(f.Background)

	l := ""
	if loading {
		l += f.spinner.View()
	}

	return style.
		Padding(0, 1).
		Width(width).
		Render(l)
}

type LoadingBeginMsg struct{}

func LoadingBeginCmd() tea.Msg {
	return LoadingBeginMsg{}
}

type LoadingEndMsg struct{}

func LoadingEndCmd() tea.Msg {
	return LoadingEndMsg{}
}
