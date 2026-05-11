package tui

import (
	"errors"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mondial7/mira/internal/listing"
)

// withFakeOpener swaps openPathFunc for a recording stub and restores it
// when the test ends.
func withFakeOpener(t *testing.T) *[]string {
	t.Helper()
	prev := openPathFunc
	calls := make([]string, 0, 1)
	openPathFunc = func(path string) error {
		calls = append(calls, path)
		return nil
	}
	t.Cleanup(func() { openPathFunc = prev })
	return &calls
}

func TestModel_OpenKeyOpensHighlightedFile(t *testing.T) {
	root := scaffoldDir(t)
	calls := withFakeOpener(t)

	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = MinCellWidth * 5

	// Cursor on the first file ("one.txt") — dirs sort first, so files
	// start at index 1+len(dirs) = 3.
	m.cursor = 3
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = next.(Model)

	if len(*calls) != 1 {
		t.Fatalf("opener invocations = %d, want 1", len(*calls))
	}
	got := (*calls)[0]
	if filepath.Base(got) != "one.txt" {
		t.Errorf("opened %q, want it to end in one.txt", got)
	}
	if m.cwd != root {
		t.Errorf("cwd changed to %q, want unchanged %q", m.cwd, root)
	}
}

func TestModel_OpenKeyOpensHighlightedDir(t *testing.T) {
	root := scaffoldDir(t)
	calls := withFakeOpener(t)

	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = MinCellWidth * 5

	// Index 1 is the first non-parent entry, "alpha".
	m.cursor = 1
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = next.(Model)

	if len(*calls) != 1 {
		t.Fatalf("opener invocations = %d, want 1", len(*calls))
	}
	if filepath.Base((*calls)[0]) != "alpha" {
		t.Errorf("opened %q, want it to end in alpha", (*calls)[0])
	}
	if m.cwd != root {
		t.Errorf("'o' should not descend into a directory; cwd = %q", m.cwd)
	}
}

func TestModel_OpenKeySkipsParent(t *testing.T) {
	root := scaffoldDir(t)
	calls := withFakeOpener(t)

	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = MinCellWidth * 5
	m.cursor = 0 // ".."

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = next.(Model)

	if len(*calls) != 0 {
		t.Errorf("opener should be a no-op on the parent entry; got %v", *calls)
	}
	if m.err != nil {
		t.Errorf("unexpected error: %v", m.err)
	}
}

func TestModel_OpenKeySurfacesError(t *testing.T) {
	root := scaffoldDir(t)
	want := errors.New("opener offline")

	prev := openPathFunc
	openPathFunc = func(string) error { return want }
	t.Cleanup(func() { openPathFunc = prev })

	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = MinCellWidth * 5
	m.cursor = 1

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = next.(Model)

	if !errors.Is(m.err, want) {
		t.Errorf("m.err = %v, want %v on status line", m.err, want)
	}
}
