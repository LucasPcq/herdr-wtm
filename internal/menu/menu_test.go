package menu_test

import (
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

func newMenu() menu.Model { return menu.New("wtm · app", menu.Items("feat/a")) }

func TestItemsOrderAndCleanLabel(t *testing.T) {
	items := menu.Items("feat/a")
	var cmds []string
	for _, it := range items {
		cmds = append(cmds, it.Cmd)
	}
	if strings.Join(cmds, ",") != "create,open,checkout,clean,prune,ui,sync" {
		t.Fatalf("order %v", cmds)
	}
	if items[3].Label != "Clean this worktree (feat/a)" {
		t.Fatalf("clean label %q", items[3].Label)
	}
	if got := menu.Items("")[3].Label; got != "Clean a worktree…" {
		t.Fatalf("clean label from main %q", got)
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
	m, cmd := send(newMenu(), click(menu.ItemsTop+4))
	if m.Chosen != "prune" || cmd == nil {
		t.Fatalf("chosen %q", m.Chosen)
	}
}

func TestClickOutsideItemsDoesNothing(t *testing.T) {
	for _, y := range []int{0, 1, menu.ItemsTop + 7, menu.ItemsTop + 9, 40} {
		m, cmd := send(newMenu(), click(y))
		if m.Chosen != "" || m.Quit || cmd != nil {
			t.Fatalf("y=%d: chosen %q", y, m.Chosen)
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

func TestViewLayoutMatchesClickRows(t *testing.T) {
	lines := strings.Split(newMenu().View(), "\n")
	if !strings.Contains(lines[0], "wtm · app") {
		t.Fatalf("title line %q", lines[0])
	}
	if !strings.Contains(lines[menu.ItemsTop], "New worktree") || !strings.Contains(lines[menu.ItemsTop], "▸") {
		t.Fatalf("first item line %q", lines[menu.ItemsTop])
	}
	if !strings.Contains(lines[menu.ItemsTop+6], "Sync workspaces") {
		t.Fatalf("last item line %q", lines[menu.ItemsTop+6])
	}
}
