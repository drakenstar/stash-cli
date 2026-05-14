package app

import tea "github.com/charmbracelet/bubbletea"

type deleteRequestMsg struct {
	Entity      string
	Title       string
	Path        string
	SkipConfirm bool
	DeleteCmd   tea.Cmd
}

type confirmDeleteMsg struct {
	Request deleteRequestMsg
}

type dismissModalMsg struct{}

type pendingDeleteState struct {
	tabID   tabID
	request deleteRequestMsg
}
