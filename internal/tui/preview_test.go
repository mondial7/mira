package tui

import (
	"fmt"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mondial7/banana-four/internal/listing"
)

// TestPreview is a manual visualisation aid: run `go test ./internal/tui
// -run TestPreview -v -preview` to print a rendered View() to stdout.
// It's skipped by default so CI stays clean.
func TestPreview(t *testing.T) {
	if os.Getenv("BF_PREVIEW") == "" {
		t.Skip("set BF_PREVIEW=1 to render a preview")
	}
	dir := os.Getenv("BF_PREVIEW_DIR")
	if dir == "" {
		dir = "."
	}
	m, err := New(dir, listing.Options{UseGitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	next, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 40})
	model := next.(Model)
	if envCursor := os.Getenv("BF_PREVIEW_CURSOR"); envCursor != "" {
		fmt.Sscanf(envCursor, "%d", &model.cursor)
	}
	fmt.Println("\n" + model.View())
}
