package menu_test

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LucasPcq/herdr-wtm/internal/menu"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func click(y int) tea.MouseMsg {
	return tea.MouseMsg{X: 4, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
}

func send(m menu.Model, msgs ...tea.Msg) (menu.Model, tea.Cmd) {
	var cmd tea.Cmd
	for _, msg := range msgs {
		var next tea.Model
		next, cmd = m.Update(msg)
		m = next.(menu.Model)
	}
	return m, cmd
}

func newMenu() menu.Model {
	return menu.New(menu.ChooseParams{Title: "app", Subtitle: "feat/a", Items: menu.Items("feat/a")})
}

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func viewLines(m menu.Model) []string {
	return strings.Split(ansi.ReplaceAllString(m.View(), ""), "\n")
}

func lineOf(t *testing.T, m menu.Model, text string) int {
	t.Helper()
	for i, l := range viewLines(m) {
		if strings.Contains(l, text) {
			return i
		}
	}
	t.Fatalf("%q not in view:\n%s", text, m.View())
	return -1
}

func TestItemsOrderGroupsAndCleanLabel(t *testing.T) {
	items := menu.Items("feat/a")
	var cmds, groups []string
	for _, it := range items {
		cmds = append(cmds, it.Cmd)
		groups = append(groups, it.Group)
	}
	if strings.Join(cmds, ",") != "create,open,checkout,clean,prune,ui,sync" {
		t.Fatalf("order %v", cmds)
	}
	if strings.Join(groups, ",") != "Worktrees,Worktrees,Worktrees,Clean up,Clean up,More,More" {
		t.Fatalf("groups %v", groups)
	}
	if items[3].Label != "Clean this worktree" || items[3].Detail != "feat/a" {
		t.Fatalf("clean item %+v", items[3])
	}
	if got := menu.Items("")[3]; got.Label != "Clean a worktree…" || got.Detail != "" {
		t.Fatalf("clean item from main %+v", got)
	}
}

func TestEnterChoosesSelection(t *testing.T) {
	m, cmd := send(newMenu(), key("down"), key("j"), key("enter"))
	if m.Chosen != "checkout" || cmd == nil {
		t.Fatalf("chosen %q, cmd %v", m.Chosen, cmd)
	}
}

func TestCursorIsClamped(t *testing.T) {
	m, _ := send(newMenu(), key("up"), key("k"))
	if m.Cursor != 0 {
		t.Fatalf("cursor %d", m.Cursor)
	}
	m, _ = send(newMenu(), key("down"), key("down"), key("down"), key("down"), key("down"), key("down"), key("down"), key("down"))
	if m.Cursor != 6 {
		t.Fatalf("cursor %d", m.Cursor)
	}
}

func TestDigitChoosesDirectly(t *testing.T) {
	m, cmd := send(newMenu(), key("6"))
	if m.Chosen != "ui" || cmd == nil {
		t.Fatalf("chosen %q", m.Chosen)
	}
}

func TestDigitOutOfRangeIgnored(t *testing.T) {
	for _, d := range []string{"0", "8", "9"} {
		m, cmd := send(newMenu(), key(d))
		if m.Chosen != "" || m.Quit || cmd != nil {
			t.Fatalf("%s: chosen %q quit %v", d, m.Chosen, m.Quit)
		}
	}
}

func TestCancelKeys(t *testing.T) {
	for _, k := range []string{"esc", "q", "ctrl+c"} {
		m, cmd := send(newMenu(), key(k))
		if !m.Quit || m.Chosen != "" || cmd == nil {
			t.Fatalf("%s: quit %v chosen %q", k, m.Quit, m.Chosen)
		}
	}
}

func TestClickOnItemChoosesIt(t *testing.T) {
	m := newMenu()
	m, cmd := send(m, click(lineOf(t, m, "Prune finished worktrees")))
	if m.Chosen != "prune" || cmd == nil {
		t.Fatalf("chosen %q", m.Chosen)
	}
}

func TestClickOutsideItemsDoesNothing(t *testing.T) {
	m := newMenu()
	for _, y := range []int{0, lineOf(t, m, "app"), lineOf(t, m, "CLEAN UP"), lineOf(t, m, "esc close"), 40} {
		got, cmd := send(m, click(y))
		if got.Chosen != "" || got.Quit || cmd != nil {
			t.Fatalf("y=%d: chosen %q", y, got.Chosen)
		}
	}
}

func TestWheelMovesSelection(t *testing.T) {
	down := tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress}
	up := tea.MouseMsg{Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress}
	m, _ := send(newMenu(), down, down, up)
	if m.Cursor != 1 || m.Chosen != "" {
		t.Fatalf("cursor %d chosen %q", m.Cursor, m.Chosen)
	}
}

func TestViewGroupsItemsUnderHeaders(t *testing.T) {
	m := newMenu()
	order := []string{"app", "feat/a", "WORKTREES", "New worktree", "Checkout a pull request", "CLEAN UP", "Clean this worktree", "Prune finished worktrees", "MORE", "Dashboard", "Sync workspaces", "esc close"}
	prev := -1
	for _, text := range order {
		at := lineOf(t, m, text)
		if at < prev {
			t.Fatalf("%q at line %d, before line %d", text, at, prev)
		}
		prev = at
	}
}

func TestViewMarksSelectionAndNumbersItems(t *testing.T) {
	lines := viewLines(newMenu())
	first := lines[lineOf(t, newMenu(), "New worktree")]
	if !strings.Contains(first, "▌") || !strings.HasSuffix(strings.TrimRight(first, " "), "1") {
		t.Fatalf("selected row %q", first)
	}
	other := lines[lineOf(t, newMenu(), "Dashboard")]
	if strings.Contains(other, "▌") || !strings.HasSuffix(strings.TrimRight(other, " "), "6") {
		t.Fatalf("row %q", other)
	}
}

func TestViewFitsTheMenuPopup(t *testing.T) {
	lines := viewLines(newMenu())
	if len(lines) > menu.Height {
		t.Fatalf("%d lines, popup holds %d", len(lines), menu.Height)
	}
	for _, l := range lines {
		if n := len([]rune(l)); n > menu.Width {
			t.Fatalf("%d columns, popup holds %d: %q", n, menu.Width, l)
		}
	}
}

func TestFailureNamesTheErrorAndHowToClose(t *testing.T) {
	got := ansi.ReplaceAllString(menu.Failure("wtm prune: exit status 1"), "")
	if !strings.Contains(got, "✗ wtm prune: exit status 1") || !strings.Contains(got, "Press Enter to close") {
		t.Fatalf("failure %q", got)
	}
}
