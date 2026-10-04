package herdr

import (
	"encoding/json"
	"fmt"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
)

// ParseContext decodes HERDR_PLUGIN_CONTEXT_JSON; an empty string is an empty context.
func ParseContext(raw string) (domain.HerdrContext, error) {
	var ctx domain.HerdrContext
	if raw == "" {
		return ctx, nil
	}
	if err := json.Unmarshal([]byte(raw), &ctx); err != nil {
		return domain.HerdrContext{}, fmt.Errorf("parse herdr plugin context: %w", err)
	}
	return ctx, nil
}
