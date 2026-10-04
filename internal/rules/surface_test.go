package rules_test

import (
	"testing"

	"github.com/LucasPcq/herdr-wtm/internal/rules"
)

func TestSurfaceLiftsADarkBackground(t *testing.T) {
	got, ok := rules.Surface("#1e1e2e")
	if !ok || got != "#30303f" {
		t.Fatalf("got %q ok %v", got, ok)
	}
}

func TestSurfaceDarkensALightBackground(t *testing.T) {
	got, ok := rules.Surface("#eff1f5")
	if !ok || got != "#dcdee1" {
		t.Fatalf("got %q ok %v", got, ok)
	}
}

func TestSurfaceReadsTerminalSpelling(t *testing.T) {
	got, ok := rules.Surface("rgb:1e1e/1e1e/2e2e")
	if !ok || got != "#30303f" {
		t.Fatalf("got %q ok %v", got, ok)
	}
}

func TestSurfaceRefusesWhatItCannotRead(t *testing.T) {
	for _, in := range []string{"", "blue", "#12345", "rgb:zz/00/00"} {
		if _, ok := rules.Surface(in); ok {
			t.Errorf("%q: want not ok", in)
		}
	}
}
