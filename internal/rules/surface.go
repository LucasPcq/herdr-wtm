package rules

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/LucasPcq/herdr-wtm/internal/domain"
)

// Surface returns the popup's background for a terminal background, as
// #rrggbb: lifted on a dark theme, lowered on a light one, so the popup stands
// out from the panes behind it whatever the theme. It reads #rrggbb and the
// rgb:rrrr/gggg/bbbb an OSC 11 query answers with.
func Surface(background string) (string, bool) {
	rgb, ok := parseColor(background)
	if !ok {
		return "", false
	}
	luma := (0.299*rgb[0] + 0.587*rgb[1] + 0.114*rgb[2]) / 255
	for i, c := range rgb {
		if luma < 0.5 {
			rgb[i] = c + (255-c)*domain.SurfaceShift
			continue
		}
		rgb[i] = c * (1 - domain.SurfaceShift)
	}
	return fmt.Sprintf("#%02x%02x%02x", round(rgb[0]), round(rgb[1]), round(rgb[2])), true
}

func parseColor(s string) ([3]float64, bool) {
	var parts []string
	switch {
	case strings.HasPrefix(s, "#") && len(s) == 7:
		parts = []string{s[1:3], s[3:5], s[5:7]}
	case strings.HasPrefix(s, "rgb:"):
		parts = strings.Split(strings.TrimPrefix(s, "rgb:"), "/")
	}
	if len(parts) != 3 {
		return [3]float64{}, false
	}
	var rgb [3]float64
	for i, p := range parts {
		if len(p) < 2 {
			return [3]float64{}, false
		}
		v, err := strconv.ParseUint(p[:2], 16, 8)
		if err != nil {
			return [3]float64{}, false
		}
		rgb[i] = float64(v)
	}
	return rgb, true
}

func round(c float64) int { return int(math.Round(c)) }
