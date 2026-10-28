package dialog

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	gloss "charm.land/lipgloss/v2"

	"charm.land/bubbles/v2/key"
)

type action struct {
	label string
	cmd   tea.Cmd
}

type KeyMap struct {
	Prev    key.Binding
	Next    key.Binding
	Confirm key.Binding
	Close   key.Binding
}

type Styles struct {
	Box          gloss.Style
	Title        gloss.Style
	Description  gloss.Style
	Button       gloss.Style
	ButtonActive gloss.Style
}

type Model struct {
	KeyMap KeyMap

	id          string
	title       string
	description string
	actions     []action
	cursor      int
	styles      Styles
	width       int
}

type Option func(*Model)

type ClosedMsg struct {
	ID     string
	Action int
}

const DefaultWidth = 50

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Prev:    key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev")),
		Next:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next")),
		Confirm: key.NewBinding(key.WithKeys("enter", "space"), key.WithHelp("enter", "confirm")),
		Close:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close")),
	}
}

func DefaultStyles() Styles {
	return Styles{
		Box:          gloss.NewStyle().Padding(1, 2).Border(gloss.RoundedBorder()),
		Title:        gloss.NewStyle().Bold(true),
		Description:  gloss.NewStyle(),
		Button:       gloss.NewStyle().Padding(0, 2),
		ButtonActive: gloss.NewStyle().Padding(0, 2).Reverse(true),
	}
}

func WithID(id string) Option                 { return func(m *Model) { m.id = id } }
func WithTitle(t string) Option               { return func(m *Model) { m.title = t } }
func WithDescription(d string) Option         { return func(m *Model) { m.description = d } }
func WithAction(l string, cmd tea.Cmd) Option { return func(m *Model) { m.AddAction(l, cmd) } }
func WithStyles(s Styles) Option              { return func(m *Model) { m.styles = s } }
func WithWidth(w int) Option                  { return func(m *Model) { m.width = w } }

func (m *Model) SetID(id string)         { m.id = id }
func (m *Model) SetTitle(t string)       { m.title = t }
func (m *Model) SetDescription(d string) { m.description = d }
func (m *Model) SetCursor(c int)         { m.cursor = clamp(c, 0, len(m.actions)-1) }
func (m *Model) SetStyles(s Styles)      { m.styles = s }
func (m *Model) SetWidth(w int)          { m.width = w }

func (m Model) ID() string          { return m.id }
func (m Model) Title() string       { return m.title }
func (m Model) Description() string { return m.description }
func (m Model) Cursor() int         { return m.cursor }
func (m Model) Width() int          { return m.width }

func (m *Model) AddAction(l string, cmd tea.Cmd) {
	m.actions = append(m.actions, action{label: l, cmd: cmd})
}

func (m *Model) ClearActions() {
	m.actions = nil
	m.cursor = 0
}

func (k KeyMap) ShortHelp() []key.Binding  { return []key.Binding{k.Prev, k.Next, k.Confirm, k.Close} }
func (k KeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

func New(opts ...Option) Model {
	m := Model{
		KeyMap: DefaultKeyMap(),
		styles: DefaultStyles(),
		width:  DefaultWidth,
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)

	if !ok {
		return m, nil
	}

	n := len(m.actions)

	switch {
	case key.Matches(keyMsg, m.KeyMap.Close):
		return m, m.close(-1)

	case key.Matches(keyMsg, m.KeyMap.Prev):
		if n > 0 {
			m.cursor = (m.cursor - 1 + n) % n
		}

	case key.Matches(keyMsg, m.KeyMap.Next):
		if n > 0 {
			m.cursor = (m.cursor + 1) % n
		}

	case key.Matches(keyMsg, m.KeyMap.Confirm):
		if n == 0 {
			return m, m.close(-1)
		}

		if m.cursor < 0 || m.cursor >= n {
			return m, nil
		}

		return m, tea.Sequence(m.close(m.cursor), m.actions[m.cursor].cmd)
	}

	return m, nil
}

func (m Model) View() string {
	s := m.styles
	w := max(m.width-s.Box.GetHorizontalFrameSize(), 0)

	sections := make([]string, 0, 3)

	if m.title != "" {
		sections = append(sections, s.Title.Width(w).Render(m.title))
	}

	if m.description != "" {
		sections = append(sections, s.Description.Width(w).Render(m.description))
	}

	if len(m.actions) > 0 {
		parts := make([]string, 0, len(m.actions))

		for i, a := range m.actions {
			if i == m.cursor {
				parts = append(parts, s.ButtonActive.Render(a.label))
			} else {
				parts = append(parts, s.Button.Render(a.label))
			}
		}

		buttons := gloss.JoinHorizontal(gloss.Top, parts...)

		sections = append(sections, gloss.NewStyle().Width(w).Align(gloss.Center).Render(buttons))
	}

	return s.Box.Render(strings.Join(sections, "\n\n"))
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}

func (m Model) close(index int) tea.Cmd {
	return func() tea.Msg { return ClosedMsg{ID: m.id, Action: index} }
}
