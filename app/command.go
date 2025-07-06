package app

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Command represents a command mode line of input from the user.
type Command string

func NewCommandCmd(s string) tea.Cmd {
	return func() tea.Msg {
		return Command(strings.TrimSpace(s))
	}
}

// Name returns all characters up to the first encountered space in an input string.  This is to be interpretted
// as the command for the rest of the input.
func (i Command) Name() string {
	idx := strings.Index(string(i), " ")
	if idx == -1 {
		return string(i)
	}
	return string(i[:idx])
}

// ArgString returns all text after the command name.  This may be interpretted in any way an action deems appropriate.
func (i Command) ArgString() string {
	idx := strings.Index(string(i), " ")
	if idx == -1 {
		return ""
	}
	return string(i[idx+1:])
}

// ArgInt attempts to parse any value given after the command as an integer.
func (i Command) ArgInt() (int, error) {
	idx := strings.Index(string(i), " ")
	if idx == -1 {
		return 0, fmt.Errorf("no argument given")
	}
	return strconv.Atoi(string(i[idx+1:]))
}

// Args returns a tokenised set of arguments that come after the initial command, not including the command itself.
// Tokens are split on space, with multiple spaces being ignored.
func (i Command) Args() []string {
	return strings.Fields(i.ArgString())
}
