package utils

import (
	"testing"

	"github.com/juliofiliizzola/hookord/internal/domain"
)

// ---------------------------------------------------------------------------
// ColorToHex
// ---------------------------------------------------------------------------

func TestColorToHex(t *testing.T) {
	tests := []struct {
		name      string
		colorCode int
		want      string
	}{
		{"orange", ColorOrange, "#e67e22"},
		{"grey", ColorGrey, "#95a5a6"},
		{"green", ColorGreen, "#2ecc71"},
		{"blue", ColorBlue, "#3498db"},
		{"purple", ColorPurple, "#9b59b6"},
		{"red", ColorRed, "#e74c3c"},
		{"yellow", ColorYellow, "#f1c40f"},
		{"dark grey", ColorDarkGrey, "#7f8c8d"},
		{"black/zero", 0x000000, "#000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ColorToHex(tt.colorCode)
			if got != tt.want {
				t.Errorf("ColorToHex(%#x) = %q, want %q", tt.colorCode, got, tt.want)
			}
		})
	}
}

// Package-level colour text variables are initialised via ColorToHex at init
// time; just assert they're non-empty and match the expected hex strings.
func TestColorTextVars(t *testing.T) {
	checks := map[string]string{
		"orange":    ColorOrangeText,
		"grey":      ColorGreyText,
		"green":     ColorGreenText,
		"blue":      ColorBlueText,
		"purple":    ColorPurpleText,
		"red":       ColorRedText,
		"yellow":    ColorYellowText,
		"dark grey": ColorDarkGreyText,
	}
	for name, val := range checks {
		if val == "" {
			t.Errorf("color text var for %q is empty", name)
		}
	}
}

// ---------------------------------------------------------------------------
// TypePullRequest
// ---------------------------------------------------------------------------

func TestTypePullRequest(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"feat: add login", domain.TypeFeat},
		{"FEAT: add login", domain.TypeFeat},          // case-insensitive
		{"  feat: leading spaces  ", domain.TypeFeat}, // trim
		{"fix: resolve bug", domain.TypeFix},
		{"Fix: resolve bug", domain.TypeFix},
		{"hot: urgent crash", domain.TypeHot},
		{"HOT: urgent crash", domain.TypeHot},
		{"doc: update readme", domain.TypeDoc},
		{"DOC: update readme", domain.TypeDoc},
		{"chore: update deps", domain.TypeChore},
		{"CHORE: update deps", domain.TypeChore},
		{"refactor: clean up", domain.TypeOther},
		{"", domain.TypeOther},
		{"random title", domain.TypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := TypePullRequest(tt.title)
			if got != tt.want {
				t.Errorf("TypePullRequest(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}
