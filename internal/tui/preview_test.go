package tui

import (
	"fmt"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/mondial7/mira/internal/listing"
)

// TestPreview is a manual visualisation aid: run `go test ./internal/tui
// -run TestPreview -v -preview` to print a rendered View() to stdout.
// It's skipped by default so CI stays clean.
func TestPreview(t *testing.T) {
	if os.Getenv("BF_PREVIEW") == "" {
		t.Skip("set BF_PREVIEW=1 to render a preview")
	}
	// Force lipgloss to emit ANSI even though the test runner isn't a TTY,
	// so the preview output has colors + bold for visual review.
	lipgloss.SetColorProfile(termenv.ANSI256)
	dir := os.Getenv("BF_PREVIEW_DIR")
	if dir == "" {
		dir = "."
	}
	m, err := New(dir, listing.Options{
		UseGitignore: true,
		ShowHidden:   os.Getenv("BF_PREVIEW_HIDDEN") != "",
	})
	if err != nil {
		t.Fatal(err)
	}
	w, h := 90, 40
	if v := os.Getenv("BF_PREVIEW_W"); v != "" {
		_, _ = fmt.Sscanf(v, "%d", &w)
	}
	if v := os.Getenv("BF_PREVIEW_H"); v != "" {
		_, _ = fmt.Sscanf(v, "%d", &h)
	}
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	model := next.(Model)
	if envCursor := os.Getenv("BF_PREVIEW_CURSOR"); envCursor != "" {
		_, _ = fmt.Sscanf(envCursor, "%d", &model.cursor)
		model.ensureCursorVisible()
	}
	if q := os.Getenv("BF_PREVIEW_SEARCH"); q != "" {
		model.startSearch()
		model.searchQuery = q
		model.updateFilter()
	}
	if v := os.Getenv("BF_PREVIEW_THEME"); v != "" {
		switch v {
		case "forest":
			model.settings.Theme = ThemeForest
		case "ocean":
			model.settings.Theme = ThemeOcean
		case "rose":
			model.settings.Theme = ThemeRose
		}
		model.applyTheme()
	}
	if v := os.Getenv("BF_PREVIEW_BORDERS"); v != "" {
		switch v {
		case "thick":
			model.settings.Borders = BorderThick
		case "dotted":
			model.settings.Borders = BorderDotted
		}
	}
	if os.Getenv("BF_PREVIEW_BIONIC_OFF") != "" {
		model.settings.Bionic = false
	}
	if os.Getenv("BF_PREVIEW_SETTINGS") != "" {
		model.openSettings()
		if v := os.Getenv("BF_PREVIEW_SETTINGS_ROW"); v != "" {
			_, _ = fmt.Sscanf(v, "%d", &model.settingsCursor)
		}
	}
	fmt.Println("\n" + model.View())
}
