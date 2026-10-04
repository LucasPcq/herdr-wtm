package rules

import (
	"fmt"
	"strings"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
)

// IsOwnCorrelation reports whether a correlation id marks a command the popup started.
func IsOwnCorrelation(id string) bool { return strings.HasPrefix(id, domain.CorrelationPrefix) }

// ProvisionFailure returns the notification for a failed worktree.provisioned.
func ProvisionFailure(ev domain.Event) (string, bool) {
	if ev.Type != domain.EventProvisioned || ev.Worktree == nil || ev.OK == nil || *ev.OK {
		return "", false
	}
	msg := ev.Worktree.Branch + ": on_create failed"
	if ev.Hook != "" {
		msg += " — " + ev.Hook
	}
	if ev.ExitCode != nil {
		msg += fmt.Sprintf(" (exit %d)", *ev.ExitCode)
	}
	return msg, true
}

type SummaryParams struct {
	Opened   []string
	Closed   int
	Failures []string
}

// Summary is one notification for what the plugin did; "" when it did nothing.
func Summary(p SummaryParams) string {
	var parts []string
	if len(p.Opened) > 0 {
		parts = append(parts, "opened "+strings.Join(p.Opened, ", "))
	}
	if p.Closed > 0 {
		parts = append(parts, fmt.Sprintf("closed %d workspace(s)", p.Closed))
	}
	if len(p.Failures) > 0 {
		parts = append(parts, "failed: "+strings.Join(p.Failures, "; "))
	}
	return strings.Join(parts, " · ")
}
