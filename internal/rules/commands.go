package rules

import (
	"slices"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
)

// IsPopupCommand reports whether the popup runs cmd: a wtm command or the menu.
func IsPopupCommand(cmd string) bool {
	return cmd == domain.CmdMenu || slices.Contains(domain.WtmCommands, cmd)
}
