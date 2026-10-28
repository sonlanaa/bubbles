package main

import (
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sonlanaa/bubbles/dialog"
)

type model struct {
	dialog dialog.Model
	help   help.Model
	open   bool
	status string
	width  int
	height int
}

const (
	actionDelete = iota
	actionCancel
)

type deletedMsg struct{}

func newModel() model {
	d := dialog.New(
		dialog.WithID("delete"),
		dialog.WithTitle("Delete file?"),
		dialog.WithDescription("report.pdf will be removed permanently. This can't be undone."),
		dialog.WithAction("Delete", func() tea.Msg { time.Sleep(500 * time.Millisecond); return deletedMsg{} }),
		dialog.WithAction("Cancel", nil),
	)

	return model{dialog: d, help: help.New(), open: false}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if !m.open {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "o":
				m.open = true
				m.dialog.SetCursor(actionCancel)
			}

			return m, nil
		}

	case dialog.ClosedMsg:
		m.open = false

		switch msg.Action {
		case actionDelete:
			m.status = "Deleting…"
		case actionCancel:
			m.status = "Canceled."
		default:
			m.status = "Closed without choosing."
		}

		return m, nil

	case deletedMsg:
		m.status = "Deleted."

		return m, nil
	}

	if !m.open {
		return m, nil
	}

	var cmd tea.Cmd

	m.dialog, cmd = m.dialog.Update(msg)

	return m, cmd
}

func (m model) View() tea.View {
	content := m.status + "\n\n" + "o open dialog · q quit"

	if m.open {
		content = m.dialog.View() + "\n\n" + m.help.View(m.dialog.KeyMap)
	}

	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true

	return v
}

func main() {
	p := tea.NewProgram(newModel())

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
