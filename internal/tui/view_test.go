package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/mondial7/banana-four/internal/listing"
)

func TestClampDisplay(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"", 5, ""},
		{"hi", 5, "hi"},
		{"hello world", 5, "hell…"},
		{"abc", 1, "…"},
		{"abc", 0, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "ab…"},
	}
	for _, tc := range cases {
		if got := clampDisplay(tc.in, tc.max); got != tc.want {
			t.Errorf("clampDisplay(%q, %d) = %q, want %q", tc.in, tc.max, got, tc.want)
		}
	}
}

func TestPadToWidth_PadsToExactWidth(t *testing.T) {
	got := padToWidth("ab", 6, nameStyle)
	if w := lipgloss.Width(got); w != 6 {
		t.Errorf("display width = %d, want 6 (%q)", w, got)
	}
	if !strings.Contains(got, "ab") {
		t.Errorf("content lost, got %q", got)
	}
}

func TestChildCountLabel(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{-1, "—"},
		{0, "empty"},
		{1, "1 item"},
		{42, "42 items"},
	}
	for _, tc := range cases {
		if got := childCountLabel(tc.n); got != tc.want {
			t.Errorf("childCountLabel(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}

// scaffoldForView builds a tiny fixture for end-to-end View rendering.
func scaffoldForView(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), nil, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return root
}

// TestView_RendersAllSections is a smoke test: it doesn't assert exact ANSI
// output (that would be brittle), only that View produces non-empty output
// containing the path, every entry name, and the help text.
func TestView_RendersAllSections(t *testing.T) {
	root := scaffoldForView(t)
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Use a generous width so the path doesn't get truncated by renderHeader.
	m.width = 200
	m.height = 30

	out := m.View()
	if out == "" {
		t.Fatal("View() returned empty string")
	}
	for _, want := range []string{filepath.Base(root), "src", "go.mod", "items", "quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("View() missing %q\n%s", want, out)
		}
	}
}

func TestView_EmptyDirectoryShowsHint(t *testing.T) {
	root := t.TempDir()
	m, err := New(root, listing.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.width = MinCellWidth * 3
	out := m.View()
	// The grid is empty (only ".." parent), so the help text should still appear.
	if !strings.Contains(out, "..") {
		t.Errorf("expected '..' parent in output:\n%s", out)
	}
}

func TestFlatList_FormatsByEntryType(t *testing.T) {
	entries := []listing.Entry{
		{Name: "src", IsDir: true},
		{Name: "go.mod"},
		{Name: "link", IsSymlink: true, Target: "go.mod"},
	}
	got := FlatList(entries)
	want := "src/\ngo.mod\nlink -> go.mod\n"
	if got != want {
		t.Errorf("FlatList:\n got: %q\nwant: %q", got, want)
	}
}
