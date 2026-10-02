// Package menu is the wtm action menu shown in the herdr popup.
package menu

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Item is one menu entry: the plugin command it runs and its label.
type Item struct {
	Cmd   string
	Label string
}

// ItemsTop is the screen row of the first entry: row 0 is the title, row 1 blank.
const ItemsTop = 2

// Items returns the menu entries. cleanBranch names the worktree the popup was
// opened from, or "" from the main checkout (wtm then shows its picker).
func Items(cleanBranch string) []Item {
	clean := "Clean a worktree…"
	if cleanBranch != "" {
		clean = fmt.Sprintf("Clean this worktree (%s)", cleanBranch)
	}
	return []Item{
		{Cmd: "create", Label: "New worktree"},
		{Cmd: "open", Label: "Open a worktree"},
		{Cmd: "checkout", Label: "Checkout a pull request"},
		{Cmd: "clean", Label: clean},
		{Cmd: "prune", Label: "Prune finished worktrees"},
		{Cmd: "ui", Label: "Dashboard"},
		{Cmd: "sync", Label: "Sync workspaces"},
	}
}

// Model is the menu state. Chosen is set when an entry is launched, Quit when
// the menu is cancelled.
type Model struct {
	Title  string
	Items  []Item
	Cursor int
	Chosen string
	Quit   bool
}

func New(title string, items []Item) Model {
	return Model{Title: title, Items: items}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s := msg.String(); s {
		case "up", "k":
			m.move(-1)
		case "down", "j":
			m.move(1)
		case "enter":
			return m.choose(m.Cursor)
		case "esc", "q", "ctrl+c":
			m.Quit = true
			return m, tea.Quit
		default:
			if len(s) == 1 && s[0] >= '1' && s[0] <= '9' {
				if i := int(s[0] - '1'); i < len(m.Items) {
					return m.choose(i)
				}
			}
		}
	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.move(-1)
		case tea.MouseButtonWheelDown:
			m.move(1)
		case tea.MouseButtonLeft:
			if i := msg.Y - ItemsTop; i >= 0 && i < len(m.Items) {
				return m.choose(i)
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.Title + "\n\n")
	for i, it := range m.Items {
		marker := "  "
		if i == m.Cursor {
			marker = "▸ "
		}
		fmt.Fprintf(&b, "%s%d  %s\n", marker, i+1, it.Label)
	}
	b.WriteString("\n↑↓ / click: choose   ⏎: run   1-7: run   esc: close\n")
	return b.String()
}

func (m *Model) move(delta int) {
	m.Cursor = min(max(m.Cursor+delta, 0), len(m.Items)-1)
}

func (m Model) choose(i int) (tea.Model, tea.Cmd) {
	m.Cursor = i
	m.Chosen = m.Items[i].Cmd
	return m, tea.Quit
}

// Choose shows the menu full-screen with mouse support and returns the chosen
// command, or "" when the user cancels.
func Choose(title string, items []Item) (string, error) {
	final, err := tea.NewProgram(New(title, items), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	if err != nil {
		return "", fmt.Errorf("menu: %w", err)
	}
	return final.(Model).Chosen, nil
}
