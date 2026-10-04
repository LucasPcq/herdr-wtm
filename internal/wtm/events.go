package wtm

import (
	"context"
	"encoding/json"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
	"github.com/LucasPcq/herdr-wtm/internal/execx"
)

// Events runs `wtm events --output json` for repo and calls handle for each
// line until the stream ends or ctx is cancelled. It returns the process's
// exit error; a clean end of stream is nil.
func (c Client) Events(ctx context.Context, repo string, handle func(domain.Event)) error {
	cmd := execx.Cmd{Dir: repo, Name: c.Bin, Args: []string{"events", "--repo", repo, "--output", "json"}}
	return c.Runner.Stream(ctx, cmd, func(line []byte) {
		var ev domain.Event
		if err := json.Unmarshal(line, &ev); err != nil {
			return
		}
		handle(ev)
	})
}
