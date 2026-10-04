// Package menu is the wtm action menu shown in the herdr popup.
package menu

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
)

// Width and Height are what the menu popup holds inside its border.
const (
	Width  = domain.MenuPopupCols - 2*domain.PopupBorderSize
	Height = domain.MenuPopupRows - 2*domain.PopupBorderSize
)

const (
	groupWorktrees = "Worktrees"
	groupCleanUp   = "Clean up"
	groupMore      = "More"
)

// ANSI colours, so the menu takes herdr's palette whatever the theme.
var (
	accent    = lipgloss.Color("4")
	branch    = lipgloss.Color("5")
	muted     = lipgloss.Color("8")
	titleSt   = lipgloss.NewStyle().Bold(true).Foreground(accent)
	branchSt  = lipgloss.NewStyle().Foreground(branch)
	groupSt   = lipgloss.NewStyle().Bold(true).Foreground(muted)
	markerSt  = lipgloss.NewStyle().Foreground(accent)
	chosenSt  = lipgloss.NewStyle().Bold(true)
	mutedSt   = lipgloss.NewStyle().Foreground(muted)
	keySt     = lipgloss.NewStyle()
	failSt    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	separator = mutedSt.Render(" · ")
)

type Item struct {
	Cmd    string
	Label  string
	Detail string
	Group  string
}

// Items returns the menu entries. cleanBranch names the worktree the popup was
// opened from, or "" from the main checkout (wtm then shows its picker).
func Items(cleanBranch string) []Item {
	clean := Item{Cmd: domain.CmdClean, Label: "Clean a worktree…", Group: groupCleanUp}
	if cleanBranch != "" {
		clean.Label, clean.Detail = "Clean this worktree", cleanBranch
	}
	return []Item{
		{Cmd: domain.CmdCreate, Label: "New worktree", Group: groupWorktrees},
		{Cmd: domain.CmdOpen, Label: "Open a worktree", Group: groupWorktrees},
		{Cmd: domain.CmdCheckout, Label: "Checkout a pull request", Group: groupWorktrees},
		clean,
		{Cmd: domain.CmdPrune, Label: "Prune finished worktrees", Group: groupCleanUp},
		{Cmd: domain.CmdUI, Label: "Dashboard", Group: groupMore},
		{Cmd: domain.CmdSync, Label: "Sync workspaces", Group: groupMore},
	}
}

// Model is the menu state. Chosen is set when an entry is launched, Quit when
// the menu is cancelled.
type Model struct {
	Title    string
	Subtitle string
	Items    []Item
	Cursor   int
	Chosen   string
	Quit     bool
}

type ChooseParams struct {
	Title    string
	Subtitle string
	Items    []Item
}

func New(p ChooseParams) Model { return Model{Title: p.Title, Subtitle: p.Subtitle, Items: p.Items} }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.onKey(msg)
	case tea.MouseMsg:
		return m.onMouse(msg)
	}
	return m, nil
}

func (m Model) onKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= len(m.Items) {
			return m.choose(n - 1)
		}
	}
	return m, nil
}

func (m Model) onMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.move(-1)
	case tea.MouseButtonWheelDown:
		m.move(1)
	case tea.MouseButtonLeft:
		rows := m.layout()
		if msg.Y >= 0 && msg.Y < len(rows) && rows[msg.Y].item >= 0 {
			return m.choose(rows[msg.Y].item)
		}
	}
	return m, nil
}

func (m Model) View() string {
	rows := m.layout()
	lines := make([]string, len(rows))
	for i, r := range rows {
		lines[i] = r.text
	}
	return strings.Join(lines, "\n")
}

// row is one line of the menu, and the item it selects (-1 for none): View and
// the mouse read the same layout, so a click lands on what is drawn there.
type row struct {
	text string
	item int
}

func (m Model) layout() []row {
	rows := []row{{"", -1}, {" " + m.header(), -1}, {"", -1}}
	for i, it := range m.Items {
		if i == 0 || it.Group != m.Items[i-1].Group {
			if i > 0 {
				rows = append(rows, row{"", -1})
			}
			rows = append(rows, row{"  " + groupSt.Render(strings.ToUpper(it.Group)), -1})
		}
		rows = append(rows, row{m.itemLine(i), i})
	}
	return append(rows, row{"", -1}, row{" " + m.footer(), -1})
}

func (m Model) header() string {
	if m.Subtitle == "" {
		return " " + titleSt.Render(m.Title)
	}
	return " " + titleSt.Render(m.Title) + separator + branchSt.Render(m.Subtitle)
}

func (m Model) itemLine(i int) string {
	it := m.Items[i]
	marker, label := "  ", it.Label
	if i == m.Cursor {
		marker, label = markerSt.Render("▌ "), chosenSt.Render(it.Label)
	}
	left := " " + marker + label
	if it.Detail != "" {
		left += "  " + branchSt.Render(it.Detail)
	}
	number := mutedSt.Render(strconv.Itoa(i + 1))
	gap := max(Width-lipgloss.Width(left)-lipgloss.Width(number)-2, 1)
	return left + strings.Repeat(" ", gap) + number
}

func (m Model) footer() string {
	hint := func(key, what string) string { return keySt.Render(key) + " " + mutedSt.Render(what) }
	return " " + strings.Join([]string{
		hint("↑↓", "move"), hint("enter", "run"), hint(fmt.Sprintf("1-%d", len(m.Items)), "run"), hint("esc", "close"),
	}, separator)
}

func (m *Model) move(delta int) {
	m.Cursor = min(max(m.Cursor+delta, 0), len(m.Items)-1)
}

func (m Model) choose(i int) (tea.Model, tea.Cmd) {
	m.Cursor = i
	m.Chosen = m.Items[i].Cmd
	return m, tea.Quit
}

// Failure is the popup's last screen when a command failed: what went wrong,
// and how to close it.
func Failure(msg string) string {
	return "\n " + failSt.Render("✗") + " " + msg + "\n\n " + mutedSt.Render("Press Enter to close.")
}

// Choose shows the menu with mouse support and returns the chosen command, or
// "" when the user cancels. No alternate screen: herdr draws it without the
// popup's tinted background.
func Choose(p ChooseParams) (string, error) {
	final, err := tea.NewProgram(New(p), tea.WithMouseCellMotion()).Run()
	if err != nil {
		return "", fmt.Errorf("menu: %w", err)
	}
	m, ok := final.(Model)
	if !ok {
		return "", fmt.Errorf("menu: unexpected model %T", final)
	}
	return m.Chosen, nil
}
